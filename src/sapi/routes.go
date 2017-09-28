package sapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func InitRoute() {
	router := mux.NewRouter()

	router.HandleFunc("/search/putsample", PutSamples).Methods("GET")
	router.HandleFunc("/search", Search).Methods("GET")
	router.HandleFunc("/search", PostUser).Methods("POST")

	http.Handle("/", router)
}
