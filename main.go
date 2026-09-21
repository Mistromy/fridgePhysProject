package main

import (
	"encoding/json"
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

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		pollShelly()
	}
}

func pollShelly() {
	answer, err := http.Get(SHELLYADDRESS)
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
