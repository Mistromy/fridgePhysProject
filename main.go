package main

import (
	"encoding/json"
	"net/http"
	"time"

	"charm.land/log/v2"
)

type Status struct {
	Id      uint8   `json:"id"`
	Source  string  `json:"source"`
	Output  bool    `json:"output"`
	Apower  float32 `json:"apower"`
	Voltage float32 `json:"voltage"`
	Current float32 `json:"current"`
	Aenergy struct {
		Total float64 `json:"total"`
	} `json:"aenergy"`
	Temperature struct {
		TC float32 `json:"tC"`
		TF float32 `json:"tF"`
	} `json:"temperature"`
}

const SHELLYADDRESS = "http://192.168.33.1/rpc/Switch.GetStatus?id=0"

func main() {
	for {
		pollShelly()
		time.Sleep(time.Second)
	}
}

func pollShelly() {
	answer, err := http.Get(SHELLYADDRESS)
	if err != nil {
		log.Error("HTTP Get", "error", err)
		return
	}
	defer answer.Body.Close()

	var status Status
	if err := json.NewDecoder(answer.Body).Decode(&status); err != nil {
		log.Error("JSON Decode", "error", err)
		return
	}
	log.Info("Shelly Status", "power", status.Apower, "voltage", status.Voltage, "tempC", status.Temperature.TC)
}
