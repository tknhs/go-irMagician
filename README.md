# go-irMagician
[![GoDoc](https://godoc.org/github.com/tknhs/go-irMagician?status.svg)](https://godoc.org/github.com/tknhs/go-irMagician)

Simple Go Library for [irMagician](http://www.omiya-giken.com/)

## Installation

```
go get github.com/tknhs/go-irMagician
```

## Usage
This example is for irMaigician T.

```go
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

	t, err := m.GetTemperature()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)
}
```

## License
This library is licensed under the MIT license.