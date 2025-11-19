package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/emicklei/gi"
)

func main() {
	http.HandleFunc("POST /gi", handleGi)
	log.Println("starting remote exec server on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleGi(w http.ResponseWriter, r *http.Request) {
	src, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg, err := gi.ParseSource(string(src))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	funcName := r.URL.Query().Get("func")
	result, err := gi.Call(pkg, funcName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
