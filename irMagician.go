package irMagician

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/goserial"
)

type Serial struct {
	SerialObject io.ReadWriteCloser
}

type IrData struct {
	Data       []int  `json:"data"`
	Format     string `json:"format"`
	Freq       int    `json:"freq"`
	Postscaler int    `json:"postscale"`
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
	irCommand := "c\r\n"
	ser.writeSerial(irCommand)
	time.Sleep(3000 * time.Millisecond)
	res, err := ser.readSerial()
	if err != nil {
		return err
	}

	res = strings.Split(res, "\r\n")[0]
	res = strings.Replace(res, "... ", "", 1)
	if _, err := strconv.Atoi(res); err != nil {
		return errors.New(res)
	}

	return nil
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
	irCommand := "t\r\n"
	ser.writeSerial(irCommand)
	time.Sleep(100 * time.Millisecond)
	res, err := ser.readSerial()
	if err != nil {
		return "", err
	}

	res = strings.Split(res, "\r\n")[0]
	temp, err := strconv.ParseFloat(res, 64)
	if err != nil {
		return "", errors.New("Cannot get temperature")
	}
	degree := ((5.0 / 1024.0 * temp) - 0.4) / (19.53 / 1000.0)

	return fmt.Sprintf("%.1f", degree), nil
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
	postscaler := irData.Postscaler

	// Data length setting
	irCommand = fmt.Sprintf("n,%d\r\n", recNumber)
	ser.writeSerial(irCommand)
	ser.readSerial()
	time.Sleep(sleepTime * time.Millisecond)

	// Post scale setting
	irCommand = fmt.Sprintf("k,%d\r\n", postscaler)
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

	// Get Capture Data Size
	ser.writeSerial("i,1\r\n")
	time.Sleep(sleepTime * time.Millisecond)
	sSize, err := ser.readSerial()
	if err != nil {
		return err
	}
	sSize = strings.Split(sSize, "\r\n")[0]
	dSize, err := strconv.ParseInt(sSize, 16, 0)
	if err != nil {
		return err
	}
	data := make([]int, int(dSize))

	// Get Postscaler value
	ser.writeSerial("i,6\r\n")
	time.Sleep(sleepTime * time.Millisecond)
	sPostscaler, err := ser.readSerial()
	if err != nil {
		return err
	}
	sPostscaler = strings.Split(sPostscaler, "\r\n")[0]
	postscaler, err := strconv.Atoi(sPostscaler)
	if err != nil {
		return err
	}

	// Bank & Data setting
	for n := 0; n < int(dSize); n++ {
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
		sMem, err := ser.readSerial()
		if err != nil {
			return err
		}
		sMem = strings.Split(sMem, " ")[0]
		mem, err := strconv.ParseInt(sMem, 16, 0)
		if err != nil {
			return err
		}
		data[n] = int(mem)
	}

	jsonData, err := json.Marshal(&IrData{Data: data, Format: "raw", Freq: 38, Postscaler: postscaler})
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}
