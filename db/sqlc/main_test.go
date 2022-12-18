package db

import (
	"database/sql"
	"fmt"
	"github.com/Chengxufeng1994/simple-bank/util"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	dbDriverName := config.DBDriver
	dataSourceName := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		config.PostgresUser, config.PostgresPassword, config.DBHost, config.DBPort, config.PostgresDB)

	testDB, err = sql.Open(dbDriverName, dataSourceName)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
