package main

import (
	"fmt"
	"log"

	"github.com/tknhs/go-irMagician"
)

func main() {
	m, err := irMagician.New("/dev/tty.usbmodem1421")
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := m.GetTemperature()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tmp)

	if err := m.CaptureSignal(); err != nil {
		log.Fatal(err)
	}
	m.Play()
	if err := m.SaveIrData("./sample.json"); err != nil {
		log.Fatal(err)
	}

	if err := m.SendIrData("./sample.json"); err != nil {
		log.Fatal(err)
	}
}
