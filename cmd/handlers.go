package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Movie struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type Handlers struct {
	db     *sql.DB
	logger *logrus.Logger
}

func (h *Handlers) getMovies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `SELECT * FROM movies LIMIT 100`)
	defer func() {
		if err = rows.Close(); err != nil {
			h.logger.Warnf("unable to close movie rows: %v", err)
		}
	}()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message": "an error occurred"}`)
		return
	}

	var movies []Movie

	for rows.Next() {
		var row Movie
		if err = rows.Scan(&row.Id, &row.Title); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"message": "an error occurred"}`)
			return
		}

		movies = append(movies, row)
	}

	response, err := json.Marshal(&movies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message": "an error occurred"}`)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf(`{"data": %s}`, response)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message": "an error occurred"}`)
		return
	}
}
