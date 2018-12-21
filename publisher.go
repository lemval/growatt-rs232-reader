// publisher
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Status struct {
	Reader      string
	Interpreter string
	Publisher   string
}

type Publisher struct {
	data   *Datagram // type var for data to be published
	empty  bool
	status *Status
}

func NewPublisher() *Publisher {
	p := new(Publisher)
	p.status = new(Status)
	p.data = NewDatagram()
	return p
}

/*
	Start the publisher by opening up an REST endpoint to publish the datagram.
*/
func (p *Publisher) start(port int) {
	serverPort := strconv.Itoa(port)
	router := mux.NewRouter()
	router.HandleFunc("/status", p.getDatagram).Methods("GET")
	router.HandleFunc("/info", p.getInfo).Methods("GET")
	Info("Starting server on port " + serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, router))
}

/*
	Listen to the supplier
*/
func (p *Publisher) listen(supplier *Interpreter, reader *Reader) {
	for {
		p.status.Interpreter = supplier.status
		p.status.Reader = reader.status
		data := supplier.pop()
		if data != nil {
			Verbose("Set datagram: " + data.Status)
			p.status.Publisher = data.Status
			p.data = data
			p.empty = false
		} else {
			p.status.Publisher = "Sleeping"
			p.data = NewDatagram()
			p.empty = true
			// Only sleep if there are no datagrams currently
			time.Sleep(10 * time.Second)
			if supplier.sleeping {
				reader.poke()
			}
		}
	}
}

/*
	Receive an JSON encoded datagram for publication
*/
func (p *Publisher) getDatagram(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.data)
}

/*
	Receive an JSON encoded datagram for publication
*/
func (p *Publisher) getInfo(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(p.status)
}
