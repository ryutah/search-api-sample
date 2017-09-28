package main

import (
	"net/http"
	"sapi"
)

func init() {
	http.HandleFunc("/search/putsample", sapi.PutSamples)
	http.HandleFunc("/search", sapi.Search)
}
