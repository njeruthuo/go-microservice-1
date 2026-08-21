package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var game Game
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&JsonErr{Error: "unable to create a game, please check your data"})
		return
	}

	now := time.Now()

	var lastInsertID int
	err := db.QueryRow(
		"INSERT INTO games(title, console, rating, completed, created, updated) VALUES ($1, $2, $3, $4, $5, $6) returning id",
		game.Title,
		game.Console,
		game.Rating,
		game.Complete,
		now,
		now,
	).Scan(&lastInsertID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&JsonErr{Error: "unable to write to database: " + err.Error()})
		return
	}

	err = db.QueryRow(
		"SELECT g.id, g.title, g.console, g.rating, g.completed, g.created::text, g.updated::text FROM games g WHERE g.id = $1",
		lastInsertID,
	).Scan(
		&game.ID,
		&game.Title,
		&game.Console,
		&game.Rating,
		&game.Complete,
		&game.Created,
		&game.Updated,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(
			&JsonErr{Error: "Unable to read from the database: " + err.Error()},
		)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&game)
}

func RetrieveGameHandler(w http.ResponseWriter, r *http.Request) {
	var game Game
	params := mux.Vars(r)
	id := params["id"]

	err := db.QueryRow(
		"select g.id, g.title, g.console, g.rating, g.completed, g.created, g.updated from games g where  g.id=$1", id,
	).Scan(
		&game.ID, &game.Title, &game.Console, &game.Rating, &game.Complete, &game.Created, &game.Updated,
	)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(&JsonErr{Error: fmt.Sprintf("a game with ID %s does not exist", id)})
		return
	}

	json.NewEncoder(w).Encode(&game)
}

func UpdateGameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var game Game
	err := db.QueryRow(
		"SELECT * FROM games g WHERE g.id=$1", id,
	).Scan(
		&game.ID,
		&game.Title,
		&game.Console,
		&game.Rating,
		&game.Complete,
		&game.Created,
		&game.Updated,
	)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			&JsonErr{Error: fmt.Sprintf("unable to find a game with ID: ", id) + err.Error()},
		)
		return
	}

	gid := game.ID
	created := game.Created
	updated := time.Now()

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&game)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(
			&JsonErr{Error: "unable to process your JSON" + err.Error()},
		)
		return
	}

	game.ID = gid
	game.Created = created
	game.Updated = updated

	_, err = db.Exec(
		"UPDATE games g SET title=$1, console=$2, rating=$3, completed=$4, created=$5, updated=$6  WHERE g.id=$7",
		game.Title,
		game.Console,
		game.Rating,
		game.Complete,
		game.Created,
		game.Updated,
		game.ID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(
			&JsonErr{Error: "Something went wrong during update" + err.Error()},
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&game)
}

func DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	_, err := db.Exec(
		"DELETE FROM games g WHERE g.id=$1", id,
	)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			&JsonErr{Error: fmt.Sprintf("unable to find a game with ID: %s", id) + err.Error()},
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RetrieveGamesHandler(w http.ResponseWriter, r *http.Request) {
	games := make([]Game, 0)

	rows, err := db.Query("SELECT g.id, g.title, g.console, g.rating, g.completed, g.created, g.updated FROM games g ORDER BY g.id")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&JsonErr{Error: "unable to retrieve games at this moment"})
		return
	}

	for rows.Next() {
		var game Game
		rows.Scan(
			&game.ID,
			&game.Title,
			&game.Console,
			&game.Rating,
			&game.Complete,
			&game.Created,
			&game.Updated,
		)

		games = append(games, game)
	}

	json.NewEncoder(w).Encode(&games)
}
