package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	DatabaseUser     = "postgres"
	DatabasePassword = "mypassword"
	DatabaseHost     = "localhost"
	DatabaseName     = "games_db"
)

var (
	db *sql.DB
)

type Game struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Console  string    `json:"console"`
	Rating   float64   `json:"rating"`
	Complete bool      `json:"complete"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updted"`
}

type JsonErr struct {
	Error string `json:"error"`
}

func ConnectDB() error {
	dbInfo := fmt.Sprintf(
		"user=%s password=%s host=%s dbname=%s sslmode=disable",
		DatabaseUser,
		DatabasePassword,
		DatabaseHost,
		DatabaseName,
	)

	var err error
	db, err = sql.Open("postgres", dbInfo)

	if err != nil {
		return err
	}

	return db.Ping()
}

func main() {
	err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/games", CreateGameHandler).Methods(http.MethodPost)
	router.HandleFunc("/games", RetrieveGamesHandler).Methods(http.MethodGet)
	router.HandleFunc("/games/{id:[0-9]+}", RetrieveGameHandler).Methods(http.MethodGet)
	router.HandleFunc("/games/{id:[0-9]+}", UpdateGameHandler).Methods(http.MethodPut, http.MethodPatch)
	router.HandleFunc("/games/{id:[0-9]+}", DeleteGameHandler).Methods(http.MethodDelete)

	log.Println("starting server on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
