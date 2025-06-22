package main

import (
	"context"
	"fmt"
	"log"

	"github.com/avfirsov/golang-backend-masterclass/api"
	db "github.com/avfirsov/golang-backend-masterclass/db/sqlc"
	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := util.LoadConfig("."); 
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}
	connPool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName))
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
	}

	defer connPool.Close()

	store := db.NewStore(connPool)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
