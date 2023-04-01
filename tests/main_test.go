package tests_test

import (
	"net/http"
	"testing"
)

func TestGetMovies(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/movies")
	if err != nil {
		t.Fatalf("could not perform the request /movies: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d got %d", http.StatusOK, resp.StatusCode)
	}
}
