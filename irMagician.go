package irMagician

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/goserial"
)

type Serial struct {
	SerialObject     io.ReadWriteCloser
	CaputureDataSize int
}

type IrData struct {
	Data      []int  `json:"data"`
	Format    string `json:"format"`
	Freq      int    `json:"freq"`
	Postscale int    `json:"postscale"`
}

const (
	sleepTime = 10
)

func New(name string) (*Serial, error) {
	c := &serial.Config{
		Name: name,
		Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		return &Serial{}, err
	}
	return &Serial{SerialObject: s}, nil
}

func (ser *Serial) writeSerial(data string) error {
	_, err := ser.SerialObject.Write([]byte(data))
	if err != nil {
		return err
	}
	return nil
}

func (ser *Serial) readSerial() (string, error) {
	buf := make([]byte, 128)
	n, err := ser.SerialObject.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func (ser *Serial) bankSet(bank int) {
	var irCommand string
	irCommand = fmt.Sprintf("b,%d\r\n", bank)
	ser.writeSerial(irCommand)
	time.Sleep(sleepTime * time.Millisecond)
}

// Capture the IR Signal
func (ser *Serial) CaptureSignal() error {
	var e error

	irCommand := "c\r\n"
	ser.writeSerial(irCommand)
	time.Sleep(3000 * time.Millisecond)
	res, err := ser.readSerial()
	if err != nil {
		return err
	}

	res = strings.Split(res, "\r\n")[0]
	data := strings.Split(res, " ")
	if dataSize, err := strconv.Atoi(data[1]); err != nil {
		ser.CaputureDataSize = 0
		e = errors.New(strings.Join(data[1:4], ""))
	} else {
		ser.CaputureDataSize = dataSize
		e = nil
	}

	return e
}

// Play the IR Signal
func (ser *Serial) Play() {
	irCommand := "p\r\n"
	ser.writeSerial(irCommand)
	time.Sleep(sleepTime * time.Millisecond)
	ser.readSerial()
}

// Get the Temperature
func (ser *Serial) GetTemperature() (string, error) {
	var degree float64

	irCommand := "t\r\n"
	ser.writeSerial(irCommand)
	time.Sleep(100 * time.Millisecond)
	res, err := ser.readSerial()
	if err != nil {
		return "Cannot read data", err
	}

	temp := strings.Split(res, "\r\n")[0]
	tf64, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return "", errors.New("Cannot get temperature")
	}
	degree = ((5.0 / 1024.0 * tf64) - 0.4) / (19.53 / 1000.0)
	degree = math.Trunc((degree*100.0+5.0)/10.0) / 10.0

	return fmt.Sprint(degree), nil
}

// Send the IR Signal from Local Data
func (ser *Serial) SendIrData(filepath string) error {
	var irData IrData
	var irCommand string

	jsonString, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonString, &irData)
	if err != nil {
		return err
	}

	recNumber := len(irData.Data)
	postScale := irData.Postscale

	// Data length setting
	irCommand = fmt.Sprintf("n,%d\r\n", recNumber)
	ser.writeSerial(irCommand)
	ser.readSerial()
	time.Sleep(sleepTime * time.Millisecond)

	// Post scale setting
	irCommand = fmt.Sprintf("k,%d\r\n", postScale)
	ser.writeSerial(irCommand)
	ser.readSerial()
	time.Sleep(sleepTime * time.Millisecond)

	// Bank & Data setting
	for n := 0; n < recNumber; n++ {
		// Bank set
		bank := n / 64
		pos := n % 64
		if pos == 0 {
			ser.bankSet(bank)
		}
		// Data set
		irCommand = fmt.Sprintf("w,%d,%d\r\n", pos, irData.Data[n])
		ser.writeSerial(irCommand)
		time.Sleep(sleepTime * time.Millisecond)
	}

	ser.Play()

	return nil
}

// Save the IR Signal in Local
func (ser *Serial) SaveIrData(filepath string) error {
	var irCommand string
	dataSize := ser.CaputureDataSize
	if dataSize == 0 {
		return errors.New("CaputureDataSize Error")
	}
	dataSlice := make([]int, 0, dataSize)

	// Bank & Data setting
	for n := 0; n < dataSize; n++ {
		// Bank set
		bank := n / 64
		pos := n % 64
		if pos == 0 {
			ser.bankSet(bank)
		}

		// Dump Memory
		irCommand = fmt.Sprintf("d,%d\r\n", pos)
		ser.writeSerial(irCommand)
		time.Sleep(sleepTime * time.Millisecond)
		data, _ := ser.readSerial()
		dataHex := strings.Split(data, " ")[0]
		dataInt, _ := strconv.ParseUint(dataHex, 16, 0)
		dataSlice = append(dataSlice, int(dataInt))
	}

	jsonData, err := json.Marshal(&IrData{Data: dataSlice, Format: "raw", Freq: 38, Postscale: 100})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}
