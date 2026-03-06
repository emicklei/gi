package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/emicklei/gi"
)

var port = flag.Int("port", 7171, "port to listen on")

func main() {
	http.HandleFunc("POST /gi", handleGi)
	log.Printf("starting remote exec server on http://0.0.0.0:%d/gi\n", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}

func handleGi(w http.ResponseWriter, r *http.Request) {
	src, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg, err := gi.Parse(string(src))
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
