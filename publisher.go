// publisher
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Publisher struct {
	data *Datagram				// type var for data to be published
}

/*
	Start the publisher by opening up an REST endpoint to publish the datagram.
*/
func (p *Publisher) start() {
	
	router := mux.NewRouter()
	router.HandleFunc("/status", p.getDatagram).Methods("GET")
	p.data = NewDatagram()

	log.Fatal(http.ListenAndServe(":5701", router))
}

/*
	Update the latest state. Use 'nil' to clean the old one
*/
func (p *Publisher) updateData(data *Datagram) {
	if data != nil {
		p.data = data
	} else {
		p.data = NewDatagram()
	}
}

/*
	Receive an JSON encoded datagram for publication
*/
func (p *Publisher) getDatagram(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.data)
}
