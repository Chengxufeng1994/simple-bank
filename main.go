package main

import (
	"database/sql"
	"github.com/Chengxufeng1994/simple-bank/api"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	_ "github.com/lib/pq"
)

const (
	dbDriverName   = "postgres"
	dataSourceName = "postgresql://postgres:postgres@localhost:5432/simple_bank?sslmode=disable"
	srvAddr        = "localhost:8080"
)

func main() {
	conn, err := sql.Open(dbDriverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	store := db.NewStore(conn)
	srv := api.NewServer(store)

	err = srv.Start(srvAddr)
	if err != nil {
		panic(err)
	}
}
