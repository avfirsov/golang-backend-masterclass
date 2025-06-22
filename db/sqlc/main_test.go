package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	var err error
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}
	testPool, err = pgxpool.New(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName))
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testPool)
	os.Exit(m.Run())
}
