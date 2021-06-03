package main

import (
	gpio "PiTally/gpio"
	"fmt"
	"os"
)

func main() {
	var args []string
	var color string
	var state string

	args = os.Args[1:]

	for i := 0; i < len(args); i = i + 2 {
		b := i + 1
		if b == len(args) {
			fmt.Println("Insufficient arguments")
			return
		}
		color = args[i]
		state = args[i+1]
		gpio.Leds(color, state)
	}
}
