// publisher
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"./diag"
	"./reader"
	"github.com/gorilla/mux"
)

type Status struct {
	Reader      string
	Interpreter string
	Publisher   string
	Init        string
}

type Publisher struct {
	data   *Datagram // type var for data to be published
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
	diag.Info("Starting server on port " + serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, router))
}

/*
	Listen to the supplier and keep track of statuses
*/
func (p *Publisher) listen(supplier *Interpreter, reader *reader.Reader) {

	var prevStatus string

	for {
		p.status.Interpreter = supplier.status
		p.status.Reader = reader.Status
		p.status.Init = reader.InitStatus

		data := supplier.getDatagram()
		if data != nil {
			if strings.Compare(prevStatus, data.Status) != 0 {
				diag.Info("Set datagram: " + data.Status + " on " + time.Now().Format("15:04:05"))
				prevStatus = data.Status
			}
			p.status.Publisher = "Data:" + data.Status
			p.data = data
		} else {
			p.status.Publisher = "No datagram on " + time.Now().Format("15:04:05")
			p.data = NewDatagram()
		}
		time.Sleep(500 * time.Millisecond)

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
