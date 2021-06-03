package leds

import (
	"fmt"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

func Leds(color, state string) {
	var line *gpiod.Lines
	var intState []int

	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}

	var redPins = []int{rpi.J8p11, rpi.J8p16, rpi.J8p29, rpi.J8p36}
	var yellowPins = []int{rpi.J8p13, rpi.J8p18, rpi.J8p31, rpi.J8p38}
	var greenPins = []int{rpi.J8p15, rpi.J8p22, rpi.J8p33, rpi.J8p40}

	switch state {
	case "on":
		intState = []int{1, 1, 1, 1}
	case "off":
		intState = []int{0, 0, 0, 0}
	default:
		fmt.Println("state must be on or off")
		return
	}

	if color == "all" {
		l1, _ := c.RequestLines(redPins, gpiod.AsOutput(), gpiod.WithPullDown)
		l2, _ := c.RequestLines(yellowPins, gpiod.AsOutput(), gpiod.WithPullDown)
		l3, _ := c.RequestLines(greenPins, gpiod.AsOutput(), gpiod.WithPullDown)
		_ = l1.SetValues(intState)
		_ = l2.SetValues(intState)
		_ = l3.SetValues(intState)
		_ = l1.Close()
		_ = l2.Close()
		_ = l3.Close()
		_ = c.Close()
		return
	}

	switch color {
	case "red":
		line, err = c.RequestLines(redPins, gpiod.AsOutput(), gpiod.WithPullDown)
		if err != nil {
			panic(err)
		}

	case "yellow":
		line, err = c.RequestLines(yellowPins, gpiod.AsOutput(), gpiod.WithPullDown)
		if err != nil {
			panic(err)
		}

	case "green":
		line, err = c.RequestLines(greenPins, gpiod.AsOutput(), gpiod.WithPullDown)
		if err != nil {
			panic(err)
		}

	default:
		fmt.Println("color must be red, yellow, green or all")
		return
	}

	err = line.SetValues(intState)
	if err != nil {
		return
	}

	err = line.Close()
	if err != nil {
		return
	}

	err = c.Close()
	if err != nil {
		return
	}
}
