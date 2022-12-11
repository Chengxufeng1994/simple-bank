package main

import (
	"context"
	"database/sql"
	sqlc "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=simple_bank sslmode=disable")
	if err != nil {
		panic(err)
	}

	queries := sqlc.New(db)
	createAccountParams := sqlc.CreateAccountParams{
		Owner:    "test",
		Balance:  1000,
		Currency: "USD",
	}
	account, err := queries.CreateAccount(ctx, createAccountParams)
	if err != nil {
		panic(err)
	}
	log.Println(account)

	listAccountParams := sqlc.ListAccountsParams{
		Limit:  10,
		Offset: 0,
	}
	accounts, err := queries.ListAccounts(ctx, listAccountParams)
	if err != nil {
		panic(err)
	}
	log.Println(accounts)
}
