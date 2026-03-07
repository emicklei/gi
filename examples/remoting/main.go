package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

	// capture stdout
	old := os.Stdout // keep backup of the real stdout
	or, ow, _ := os.Pipe()
	os.Stdout = ow

	// call it
	_, err = gi.Call(pkg, funcName)

	// read stdout
	ow.Close()
	os.Stdout = old // restoring the real stdout
	buf := new(bytes.Buffer)
	buf.ReadFrom(or)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	fmt.Fprint(w, buf.String())
}
