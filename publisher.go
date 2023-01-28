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
		if user != "" {
			diag.Info("Authenticated with '" + user + "'.")
		}
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

	if user != "" {
		p.opts.SetUsername(user)
		p.opts.SetPassword(credential)
	}
}

func Item(name string, value string) string    { return _element(name, "\""+value+"\"", false) }
func ItemEnd(name string, value string) string { return _element(name, "\""+value+"\"", true) }
func Object(name string, value string) string  { return _element(name, "{"+value+"}", false) }
func _element(name string, value string, end bool) string {
	postfix := ","
	if end {
		postfix = ""
	}
	return "\"" + name + "\":" + value + postfix
}

func (p *Publisher) discoveryHomeAssist() {
	client := mqtt.NewClient(p.opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		diag.Warn("MQTT not available!")
		return
	}

	diag.Warn("Discovery for Home Assistant")

	configArray := HomeAssistantConfig()
	for i := 0; i < len(configArray); i++ {
		item := configArray[i]
		device := Object("device", Item("name", "Growatt Reader")+
			Item("sw_version", Version)+
			Item("identifiers", "lemval_growatt_inverter_reader")+
			ItemEnd("manufacturer", "Growatt"))

		class := ""
		if item.device != "" {
			class = Item("device_class", item.device)
		}

		unit := ""
		if item.unit != "" {
			unit = Item("unit_of_measurement", item.unit)
		}
		state := ""
		if item.state != "" {
			state = Item("state_class", item.state)
		}
		payload := "{" +
			Item("name", item.name) +
			class +
			unit +
			state +
			device +
			Item("object_id", topic+"_"+item.name) +
			Item("unique_id", item.id) +
			ItemEnd("state_topic", "/solar/"+topic+"/"+item.name) +
			"}"

		client.Publish(
			"homeassistant/sensor/"+topic+"/"+item.name+"/config",
			0, true,
			payload).Wait()
	}

	client.Disconnect(250)
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

	if p.opts != nil {
		p.discoveryHomeAssist()
	}

	for {
		p.status.Interpreter = supplier.status
		p.status.Reader = reader.Status
		p.status.Init = reader.InitStatus
		statusUpdated = false

		data := supplier.getDatagram()
		if data != nil {
			if strings.Compare(prevStatus, data.Status) != 0 {
				diag.Info("Status updated to " + data.Status + " on " + time.Now().Format("15:04:05"))
				prevStatus = data.Status
				statusUpdated = true
			}
			day := time.Now().Day()
			if day != p.publishDay {
				diag.Warn(fmt.Sprintf("Day updated from %d to %d", p.publishDay, day))
				p.publishDay = day
				data.DayProduction = 0
				statusUpdated = true
			}
			p.status.Publisher = "Data:" + data.Status
			p.prevData = p.data
			p.data = data
		} else {
			diag.Info("Missing data on " + time.Now().Format("15:04:05"))
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

	// if statusUpdated {
	// 	diag.Info("Processing for MQTT: " + p.data.String())
	// 	diag.Info("Previous dataset   : " + p.prevData.String())
	// }

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
			if precision >= 0 {
				factor := math.Pow10(precision)
				oldValue = math.Round(oldValue*factor) / factor
				newValue = math.Round(newValue*factor) / factor
			}
			diff := math.Abs(oldValue - newValue)

			// if statusUpdated {
			// 	diag.Info(fmt.Sprintf("On field: %s; publish: %t", field.Name, diff > TOLERANCE))
			// 	diag.Info(fmt.Sprintf("From    : %f to %f", oldValue, newValue))
			// }
			if statusUpdated || diff > TOLERANCE {
				// diag.Info(fmt.Sprintf("Publishing: %f -> %f to %s", oldValue, newValue, p.topicRoot+field.Name))
				token := client.Publish(p.topicRoot+field.Name, 0, false, fmt.Sprintf("%.1f", newValue))
				token.Wait()
			}
		case reflect.Int:
			oldValue := elemOld.Int()
			newValue := elemNew.Int()
			// diag.Info(fmt.Sprintf("Int value: %d -> %d (%s)", oldValue, newValue, p.topicRoot+field.Name))
			if diff := math.Abs(float64(oldValue - newValue)); statusUpdated || diff > TOLERANCE {
				// diag.Info(fmt.Sprintf("Publishing: %d", newValue))
				token := client.Publish(p.topicRoot+field.Name, 0, false, fmt.Sprintf("%d", newValue))
				token.Wait()
			}
		case reflect.String:
			// diag.Info(fmt.Sprintf("String value: %s -> %s (%s)", elemOld.String(), elemNew.String(), p.topicRoot+field.Name))
			if statusUpdated || strings.Compare(elemNew.String(), elemOld.String()) != 0 {
				// Only one is currently 'Status' which should be retained
				token := client.Publish(p.topicRoot+field.Name, 0, true, elemNew.String())
				token.Wait()
			}
		default:
			// diag.Info(fmt.Sprintf("Other value: %s -> %s (%s)", elemOld.String(), elemNew.String(), p.topicRoot+field.Name))
			// elemNew.Type().String() is always time.Time
			timeValue, _ := elemNew.Interface().(time.Time)
			token := client.Publish(p.topicRoot+field.Name, 0, false, timeValue.Format("2006-01-02T15:04:05-07:00"))
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
