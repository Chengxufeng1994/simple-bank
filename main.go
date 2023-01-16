package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Chengxufeng1994/simple-bank/api"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	_ "github.com/Chengxufeng1994/simple-bank/doc/statik"
	"github.com/Chengxufeng1994/simple-bank/gapi"
	"github.com/Chengxufeng1994/simple-bank/pb"
	"github.com/Chengxufeng1994/simple-bank/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	dbDriverName := config.DBDriver
	dataSourceName := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.PostgresUser, config.PostgresPassword, config.DBHost, config.DBPort, config.PostgresDB)

	conn, err := sql.Open(dbDriverName, dataSourceName)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGrpcSrv(config, store)
}

func runGinSrv(config util.Config, store db.Store) {
	serverAddr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
	srv, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = srv.Start(serverAddr)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func runGrpcSrv(config util.Config, store db.Store) {
	srv, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcSrv, srv)
	reflection.Register(grpcSrv)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}

	log.Printf("gRPC server listening at %v", listener.Addr().String())
	err = grpcSrv.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server: ", err)
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
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
		log.Fatal("cannot register handler server: ", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs", err)
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	HTTPAddr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
	listener, err := net.Listen("tcp", HTTPAddr)
	if err != nil {
		log.Fatal("cannot create listener: ", err)
	}

	log.Printf("HTTP gateway server listening at %v", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server: ", err)
	}
}
