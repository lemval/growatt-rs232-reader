// publisher
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"./diag"
	"./reader"
	"github.com/gorilla/mux"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	prevData  *Datagram
	opts      *mqtt.ClientOptions
	topicRoot string
}

func NewPublisher() *Publisher {
	p := new(Publisher)
	p.status = new(Status)
	p.data = NewDatagram()
	p.prevData = p.data

	if broker != "" {
		diag.Info("Using MQTT via : " + broker + " on /solar/" + topic)

		p.topicRoot = "/solar/" + topic + "/"
		p.opts = mqtt.NewClientOptions().
			AddBroker(broker).
			SetCleanSession(false).
			SetClientID("Growatt connector")
		// opts.SetUsername(user)
		// opts.SetPassword(password)
	}

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
			p.prevData = p.data
			p.data = data
		} else {
			p.status.Publisher = "No datagram on " + time.Now().Format("15:04:05")
			p.prevData = p.data
			p.data = NewDatagram()
		}

		if p.opts != nil {
			p.publishMQTT()
		}

		time.Sleep(500 * time.Millisecond)
	}
}

/*
	Listen to the supplier and keep track of statuses
*/
func (p *Publisher) publishMQTT() {

	if p.data == nil {
		return
	}

	client := mqtt.NewClient(p.opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		diag.Warn("MQTT connection failed!")
		panic(token.Error())
	}

	// diag.Info("Processing for MQTT: " + p.data.String())

	fields := reflect.TypeOf(*p.data)
	valuesNew := reflect.ValueOf(*p.data)
	valuesOld := reflect.ValueOf(*p.prevData)
	num := fields.NumField()

	const TOLERANCE = 0.00001

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		elemNew := valuesNew.Field(i)
		elemOld := valuesOld.Field(i)

		switch field.Type.Kind() {
		case reflect.Float32:
			oldValue := elemOld.Float()
			newValue := elemNew.Float()
			// diag.Info(fmt.Sprintf("Float value: %f -> %f", oldValue, newValue))
			if diff := math.Abs(oldValue - newValue); diff > TOLERANCE {
				// diag.Info(fmt.Sprintf("Publishing: %f to %s", newValue, p.topicRoot+field.Name))
				token := client.Publish(p.topicRoot+field.Name, 0, false, fmt.Sprintf("%.1f", newValue))
				token.Wait()
			}
		case reflect.Int:
			oldValue := elemOld.Int()
			newValue := elemNew.Int()
			// diag.Info(fmt.Sprintf("Int value: %d -> %d", oldValue, newValue))
			if diff := math.Abs(float64(oldValue - newValue)); diff > TOLERANCE {
				// diag.Info(fmt.Sprintf("Publishing: %d", newValue))
				token := client.Publish(p.topicRoot+field.Name, 0, false, fmt.Sprintf("%d", newValue))
				token.Wait()
			}
		case reflect.String:
			// diag.Info(fmt.Sprintf("String value: %s -> %s", elemOld.String(), elemNew.String()))
			if strings.Compare(elemNew.String(), elemOld.String()) != 0 {
				token := client.Publish(p.topicRoot+field.Name, 0, false, elemNew.String())
				token.Wait()
			}
		default:
			// elemNew.Type().String() is always time.Time
			timeValue,_ :=  elemNew.Interface().(time.Time)
			token := client.Publish(p.topicRoot+field.Name, 0, false, timeValue.Format("2006-01-02 15:04:05"))
			token.Wait()
		}
	}
	client.Disconnect(250)
}

/*
	Receive an JSON encoded datagram for publication
*/
func (p *Publisher) getDatagram(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(p.data)
}

/*
	Receive an JSON encoded datagram for publication
*/
func (p *Publisher) getInfo(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(p.status)
}
