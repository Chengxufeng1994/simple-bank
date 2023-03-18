package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/Chengxufeng1994/simple-bank/api"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	_ "github.com/Chengxufeng1994/simple-bank/doc/statik"
	"github.com/Chengxufeng1994/simple-bank/gapi"
	"github.com/Chengxufeng1994/simple-bank/pb"
	"github.com/Chengxufeng1994/simple-bank/util"
	"github.com/Chengxufeng1994/simple-bank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	dbDriverName := config.DBDriver
	dataSourceName := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.PostgresUser, config.PostgresPassword, config.DBHost, config.DBPort, config.PostgresDB)

	conn, err := sql.Open(dbDriverName, dataSourceName)
	if err != nil {
		log.Fatal().Msg("cannot connect to db")
	}
	runDBMigration(config.MigrationURL, dataSourceName)

	store := db.NewStore(conn)

	redisOpts := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	go runTaskProcessor(config, redisOpts, store)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcSrv(config, store, taskDistributor)
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

func runGinSrv(config util.Config, store db.Store) {
	serverAddr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
	srv, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	err = srv.Start(serverAddr)
	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}

func runGrpcSrv(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	srv, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcSrv := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcSrv, srv)
	reflection.Register(grpcSrv)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}

	log.Info().Msgf("gRPC server listening at %v", listener.Addr().String())
	err = grpcSrv.Serve(listener)
	if err != nil {
		log.Fatal().Msg("cannot start gRPC server")
	}
}

func runTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	processor := worker.NewRedisTaskProcessor(redisOpt, store)
	log.Info().
		Msg("start task processor")
	if err := processor.Start(); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to start task processor")
	}
}

func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Msg("cannot create statik fs")
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	HTTPAddr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
	listener, err := net.Listen("tcp", HTTPAddr)
	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}

	log.Info().Msgf("HTTP gateway server listening at %v", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Msg("cannot start HTTP gateway server")
	}
}
