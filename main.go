package main

import (
	"context"
	"fmt"
	"log"

	"github.com/avfirsov/golang-backend-masterclass/api"
	db "github.com/avfirsov/golang-backend-masterclass/db/sqlc"
	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

func main() {
	if err := util.LoadConfig(); err != nil {
		log.Fatal("failed to load config: ", err)
	}
	connPool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", viper.GetString("DB_USER"), viper.GetString("DB_PASSWORD"), viper.GetString("DB_HOST"), viper.GetString("DB_PORT"), viper.GetString("DB_NAME")))
	if err != nil {
		log.Fatal("failed to connect to db: ", err)
	}

	defer connPool.Close()

	store := db.NewStore(connPool)
	server := api.NewServer(store)

	err = server.Start(viper.GetString("SERVER_ADDRESS"))

	if err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
