package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"charm.land/log/v2"
)

type Status struct {
	TimeStamp time.Time `json:"timestamp"`
	Id        uint8     `json:"id"`
	Source    string    `json:"source"`
	Output    bool      `json:"output"`
	Apower    float32   `json:"apower"`
	Voltage   float32   `json:"voltage"`
	Current   float32   `json:"current"`
	Aenergy   struct {
		Total float64 `json:"total"`
	} `json:"aenergy"`
	Temperature struct {
		TC float32 `json:"tC"`
		TF float32 `json:"tF"`
	} `json:"temperature"`
}

const SHELLYADDRESS = "http://192.168.33.1/rpc/Switch.GetStatus?id=0"

var client http.Client

var VictoriaStatus bool

func main() {
	log.Info("Shelly Logger Loading...")
	configClient()
	victoriaUp, grafanaUp := checkRequirements()
	if grafanaUp == true && victoriaUp == true {
		log.Info("Requirements status", "Grafana", grafanaUp, "VictoriaMetrics", victoriaUp)
	} else {
		log.Warn("Requirements status", "Grafana", grafanaUp, "VictoriaMetrics", victoriaUp)
	}
	if victoriaUp == false {
		log.Warn("VictoriaMetrics isn't running. or is not detected. Graphs will not be able to get generated. Json Data will continue Collecting.")
	}

	go tempCheckVictoria()

	log.Info("Shelly Address", "address", SHELLYADDRESS)
	log.Info("Log frequency", "requests per second", 1)
	log.Info("Starting Measurements")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		pollShelly()
	}
}

func pollShelly() {
	answer, err := client.Get(SHELLYADDRESS)
	if err != nil {
		log.Error("HTTP Get", "error", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("Body Close", "error", err)
		}
	}(answer.Body)

	var status Status
	if err := json.NewDecoder(answer.Body).Decode(&status); err != nil {
		log.Error("JSON Decode", "error", err)
		return
	}
	log.Info("Shelly Status", "power", status.Apower, "voltage", status.Voltage, "tempC", status.Temperature.TC)
	status.TimeStamp = time.Now()
	err = saveToFile(status)
	if err != nil {
		log.Error("saveToFile", "error", err)
		return
	}
	go writeToVictoria(status)
}

func saveToFile(v any) error {
	f, err := os.OpenFile("dump.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Error("Close", "error", err)
		}
	}(f)
	return json.NewEncoder(f).Encode(v)
}

// configClient configures settings for the package level http client variable
func configClient() {
	client.Timeout = time.Second * 3
}

func tempCheckVictoria() {
	victoriaup, _ := checkRequirements()
	VictoriaStatus = victoriaup
	time.Sleep(time.Minute)
}

func writeToVictoria(status Status) {
	if VictoriaStatus == false {
		return
	}
	var outputValue uint8
	if status.Output == true {
		outputValue = 1
	} else {
		outputValue = 0
	}
	request := fmt.Sprintf("shelly output=%d,apower=%v,voltage=%v,current=%v,aenergy_total=%v,temperature_tC=%v,temperature_tF=%v %d", outputValue, status.Apower, status.Voltage, status.Current, status.Aenergy.Total, status.Temperature.TC, status.Temperature.TF, status.TimeStamp.UnixNano())

	const victoriaAddress = "http://127.0.0.1:8428/write"
	byteRequest := []byte(request)
	reader := bytes.NewReader(byteRequest)
	victoriaResponse, err := client.Post(victoriaAddress, "text/plain", reader)
	if err != nil {
		log.Error("HTTP Post", "error", err)
		return
	}
	log.Debug("Victoria", "response", victoriaResponse)
}

func checkRequirements() (victoriaUp, grafanaUp bool) {
	victoriaUp = true
	grafanaUp = true
	grafanaStatus, err := client.Get("http://localhost:3000/api/health")
	if err != nil {
		log.Error("Grafana Status Check", "error", err)
		grafanaUp = false
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Error("Body Close", "error", err)
			}
		}(grafanaStatus.Body)
	}
	victoriaStatus, err := client.Get("http://localhost:8428/health")
	if err != nil {
		log.Error("VictoriaMetrics Status Check", "error", err)
		victoriaUp = false
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Error("Body Close", "error", err)
			}
		}(victoriaStatus.Body)
	}
	return victoriaUp, grafanaUp
}
