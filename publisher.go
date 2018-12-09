// publisher
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Publisher struct {
	data *Datagram
}

func (p *Publisher) start() {
	router := mux.NewRouter()
	router.HandleFunc("/status", p.getDatagram).Methods("GET")

	log.Fatal(http.ListenAndServe(":5701", router))
}

func (p *Publisher) updateData(data *Datagram) {
	p.data = data
}

func (p *Publisher) getDatagram(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.data)
}
