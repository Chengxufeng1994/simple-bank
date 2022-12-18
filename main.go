package main

import (
	"database/sql"
	"fmt"
	"github.com/Chengxufeng1994/simple-bank/api"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"github.com/Chengxufeng1994/simple-bank/util"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	dbDriverName := config.DBDriver
	dataSourceName := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.PostgresUser, config.PostgresPassword, config.DBHost, config.DBPort, config.PostgresDB)
	serverAddr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)

	conn, err := sql.Open(dbDriverName, dataSourceName)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	srv := api.NewServer(store)
	err = srv.Start(serverAddr)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
