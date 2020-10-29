package main

import (
	"fmt"
	"goMinimalAPI/api"
	"goMinimalAPI/store"
	"log"
	"net/http"
)

var (
	DSN = "user=postgres password=testPWD99 host=localhost dbname=usersapi sslmode=disable"
)

type Config struct {
	ListenAddr string
}

func main() {
	srvCfg := &Config{
		ListenAddr: "127.0.0.1:8080",
	}

	usersStore, err := store.NewUserStore(DSN)
	if err != nil {
		log.Printf("Error creating UsersStore: %v", err)
		return
	}

	apiMux, err := api.NewUserAPIMux(usersStore)
	if err != nil {
		log.Printf("Error creating users API Handler: %v ", err)
		return
	}

	log.Printf("Starting API on %s\n", srvCfg.ListenAddr)
	err = http.ListenAndServe(srvCfg.ListenAddr, apiMux)
	fmt.Println(err)
	err = usersStore.CloseDB()
	if err != nil {
		log.Printf("Error closing DB: %v\n", err)
	}
}
