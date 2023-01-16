package main

import (
	"database/sql"
	"fmt"
	"github.com/Chengxufeng1994/simple-bank/api"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"github.com/Chengxufeng1994/simple-bank/gapi"
	"github.com/Chengxufeng1994/simple-bank/pb"
	"github.com/Chengxufeng1994/simple-bank/util"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
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

	log.Printf("grpc server listening at %v", listener.Addr().String())
	err = grpcSrv.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server: ", err)
	}
}
