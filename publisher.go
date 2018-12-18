// publisher
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Publisher struct {
	data  *Datagram // type var for data to be published
	empty bool
}

/*
	Start the publisher by opening up an REST endpoint to publish the datagram.
*/
func (p *Publisher) start(port int) {

	serverPort := strconv.Itoa(port)
	router := mux.NewRouter()
	router.HandleFunc("/status", p.getDatagram).Methods("GET")
	p.data = NewDatagram()
	Info("Starting server on port " + serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, router))
}

/*
	Update the latest state. Use 'nil' to clean the old one (unless previously nil)
*/
func (p *Publisher) updateData(data *Datagram) {
	if data != nil {
		p.data = data
		p.empty = false
	} else if !p.empty {
		p.data = NewDatagram()
		p.empty = true
	}
}

/*
	Receive an JSON encoded datagram for publication
*/
func (p *Publisher) getDatagram(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.data)
}
