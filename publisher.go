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

	"growattrr/diag"
	"growattrr/reader"

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

	prevData   *Datagram
	prevMqtt   *Datagram
	nextUpdate time.Time
	opts       *mqtt.ClientOptions
	topicRoot  string
	period     int
	publishDay int
}

func NewPublisher(delay int) *Publisher {
	p := new(Publisher)
	p.status = new(Status)
	p.data = NewDatagram()
	p.prevData = p.data
	p.prevMqtt = p.data
	p.period = delay
	p.nextUpdate = time.Now()
	p.publishDay = time.Now().Day()

	if broker != "" {
		diag.Info("Using MQTT via : " + broker + " on /solar/" + topic)
		if p.period > 0 {
			diag.Info(fmt.Sprintf("Publish once every %d seconds.", p.period))
		}

		p.topicRoot = "/solar/" + topic + "/"
		p.initMqttConnection()
	}

	return p
}

func (p *Publisher) initMqttConnection() {
	p.opts = mqtt.NewClientOptions().
		AddBroker(broker).
		SetCleanSession(false).
		SetClientID("Growatt connector")
	// opts.SetUsername(user)
	// opts.SetPassword(password)
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
	var statusUpdated bool

	for {
		p.status.Interpreter = supplier.status
		p.status.Reader = reader.Status
		p.status.Init = reader.InitStatus
		statusUpdated = false

		data := supplier.getDatagram()
		if data != nil {
			if strings.Compare(prevStatus, data.Status) != 0 {
				diag.Info("Set datagram: " + data.Status + " on " + time.Now().Format("15:04:05"))
				prevStatus = data.Status
				statusUpdated = true
			}
			day := time.Now().Day()
			if day != p.publishDay {
				p.publishDay = day
				data.DayProduction = 0
				statusUpdated = true
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
			p.publishMQTT(false, statusUpdated)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

/*
	Listen to the supplier and keep track of statuses
*/
func (p *Publisher) publishMQTT(retry bool, statusUpdated bool) {

	if p.data == nil {
		return
	}

	if !statusUpdated && p.period > 0 {
		if time.Now().Before(p.nextUpdate) {
			// Update later
			return
		}
		p.nextUpdate = time.Now().Add(time.Second * time.Duration(delay))
	}

	client := mqtt.NewClient(p.opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		if !retry {
			p.initMqttConnection()
			p.publishMQTT(true, statusUpdated)
		} else {
			diag.Warn("MQTT not available!")
			return
		}
	}

	// diag.Info("Processing for MQTT: " + p.data.String())

	// Use reflection to handle fields in data type
	fields := reflect.TypeOf(*p.data)
	valuesNew := reflect.ValueOf(*p.data)
	valuesOld := reflect.ValueOf(*p.prevMqtt)
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
			// diag.Info(fmt.Sprintf("Float value: %f -> %f (%s)", oldValue, newValue, p.topicRoot+field.Name))
			if diff := math.Abs(oldValue - newValue); diff > TOLERANCE {
				// diag.Info(fmt.Sprintf("Publishing: %f to %s", newValue, p.topicRoot+field.Name))
				token := client.Publish(p.topicRoot+field.Name, 0, false, fmt.Sprintf("%.1f", newValue))
				token.Wait()
			}
		case reflect.Int:
			oldValue := elemOld.Int()
			newValue := elemNew.Int()
			// diag.Info(fmt.Sprintf("Int value: %d -> %d (%s)", oldValue, newValue, p.topicRoot+field.Name))
			if diff := math.Abs(float64(oldValue - newValue)); diff > TOLERANCE {
				// diag.Info(fmt.Sprintf("Publishing: %d", newValue))
				token := client.Publish(p.topicRoot+field.Name, 0, false, fmt.Sprintf("%d", newValue))
				token.Wait()
			}
		case reflect.String:
			// diag.Info(fmt.Sprintf("String value: %s -> %s (%s)", elemOld.String(), elemNew.String(), p.topicRoot+field.Name))
			if strings.Compare(elemNew.String(), elemOld.String()) != 0 {
				token := client.Publish(p.topicRoot+field.Name, 0, false, elemNew.String())
				token.Wait()
			}
		default:
			// diag.Info(fmt.Sprintf("Other value: %s -> %s (%s)", elemOld.String(), elemNew.String(), p.topicRoot+field.Name))
			// elemNew.Type().String() is always time.Time
			timeValue, _ := elemNew.Interface().(time.Time)
			token := client.Publish(p.topicRoot+field.Name, 0, false, timeValue.Format("2006-01-02 15:04:05"))
			token.Wait()
		}
	}
	p.prevMqtt = p.data

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
