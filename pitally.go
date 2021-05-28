package main

import (
	"bufio"
	"fmt"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type vmixClientType struct {
	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Reader
	sync.Mutex
	connected bool
}

type tally struct {
	action string // Bus or Input
	value  string
}

type ledReq struct {
	color string //red, yellow, green or all
	mode  string //all, off, rotate, strobe
}

var vmixClient = new(vmixClientType)
var wg sync.WaitGroup
var vmixMessageChan = make(chan string)
var vmixIP = "192.168.1.173"
var ledChan = make(chan ledReq)
var buttonChan = make(chan time.Time, 10)
var t1 = tally{
	action: "Bus",
	value:  "C",
}

func init() {
	//Connect to the vmix API
	err := vmixAPIConnect(vmixIP + ":8099")
	if err != nil {
		fmt.Println("Error connecting to vmix API:")
		panic(err)
	}
}

// vmixAPIConnect connects to the vMix API. apiAddress is a string
// of the format ipaddress:port.  By default, the vMix API is on port 8099.
// If vMix is not up, this function will continue trying to connect, and will
// block until a connection is achieved.
func vmixAPIConnect(apiAddress string) error {
	vmixClient.connected = false
	for vmixClient.connected == false {
		timeout := time.Second * 20
		conn, err := net.DialTimeout("tcp", apiAddress, timeout)

		if err == nil {
			vmixClient.conn = conn
			vmixClient.w = bufio.NewWriter(conn)
			vmixClient.r = bufio.NewReader(conn)
			vmixClient.connected = true
		} else if strings.Contains(err.Error(), "connection timed out") ||
			strings.Contains(err.Error(), "connection refused") {
			fmt.Println("vmix api is inaccessible.  Probably because vMix is not running")
			fmt.Println("Waiting 5 seconds and trying again")
			vmixClient.connected = false
			time.Sleep(time.Second * 5)
		} else {
			fmt.Println("Unable to connect. Error was: ", err)
			return err
		}
	}
	return nil
}

func SendMessage(message string) error {
	fmt.Println(message)
	vmixClient.Lock()
	pub := fmt.Sprintf("%v\r\n", message)
	_, err := vmixClient.w.WriteString(pub)
	if err == nil {
		err = vmixClient.w.Flush()
	}
	vmixClient.Unlock()
	return err
}

func getMessage() {

	// Subscribe to the activator feed in the vMix API
	err := SendMessage("SUBSCRIBE ACTS")
	if err != nil {
		fmt.Println("Error in GetMessage.SendMessage: ", err)
		wg.Done()
	}

	//Capture all responses from the vMix API
	for {
		line, err := vmixClient.r.ReadString('\n')

		if err == nil {
			vmixMessageChan <- line
			fmt.Println(line)
		} else {
			wg.Done()
			fmt.Println("Error in GetMessage.ReadString: ", err)
		}
	}
}

func processVmixMessage() {
	for {
		vmixMessage := <-vmixMessageChan
		messageSlice := strings.Fields(vmixMessage)

		var state int

		// ex:  [ACTS OK InputPlaying 9 1]
		// messageSlice[2] - Action
		// messageSlice[3] - Input
		// messageSlice[4] - State (usually 0 for off, 1 for on)

		if t1.action == "Input" && messageSlice[0] == "ACTS" && messageSlice[1] == "OK" &&
			messageSlice[2] == "Input" && messageSlice[3] == t1.value {
			state, _ = strconv.Atoi(messageSlice[4])
			fmt.Println("Input changed: ", messageSlice)

			if state == 0 {
				fmt.Println(messageSlice[2], " off")
			}
			if state == 1 {
				fmt.Println(messageSlice[2], " on")
			}

		}

		if t1.action == "Bus" {
			if messageSlice[2] == t1.action+t1.value+"Audio" {
				state, _ = strconv.Atoi(messageSlice[3])
				if state == 0 {
					fmt.Println(messageSlice[2], " off")
				}
				if state == 1 {
					fmt.Println(messageSlice[2], " on")
				}
			}
		}

	}
}

func lights() {

	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}

	var redPins = []int{rpi.J8p11, rpi.J8p16, rpi.J8p29, rpi.J8p36}
	var yellowPins = []int{rpi.J8p13, rpi.J8p18, rpi.J8p31, rpi.J8p38}
	var greenPins = []int{rpi.J8p15, rpi.J8p22, rpi.J8p33, rpi.J8p40}

	redLines, err := c.RequestLines(redPins, gpiod.AsOutput(0, 0, 0, 0), gpiod.WithPullDown)
	if err != nil {
		panic(err)
	}

	yellowLines, err := c.RequestLines(yellowPins, gpiod.AsOutput(0, 0, 0, 0), gpiod.WithPullDown)
	if err != nil {
		panic(err)
	}

	greenLines, err := c.RequestLines(greenPins, gpiod.AsOutput(0, 0, 0, 0), gpiod.WithPullDown)
	if err != nil {
		panic(err)
	}

	// Lines are now configured, we can safely close the chip
	err = c.Close()
	if err != nil {
		panic(err)
	}

	for {
		req := <-ledChan
		on := []int{1, 1, 1, 1}
		off := []int{0, 0, 0, 0}

		switch color := req.color; color {
		case "red":
			if req.mode == "all" {
				_ = redLines.SetValues(on)
			}
			if req.mode == "off" {
				_ = redLines.SetValues(off)
			}

		case "yellow":
			if req.mode == "all" {
				_ = yellowLines.SetValues(on)
			}
			if req.mode == "off" {
				_ = yellowLines.SetValues(off)
			}

		case "green":
			if req.mode == "all" {
				_ = greenLines.SetValues(on)
			}
			if req.mode == "off" {
				_ = greenLines.SetValues(off)
			}

		case "all":
			if req.mode == "all" {
				_ = redLines.SetValues(on)
				_ = yellowLines.SetValues(on)
				_ = greenLines.SetValues(on)
			}
			if req.mode == "off" {
				_ = redLines.SetValues(off)
				_ = yellowLines.SetValues(off)
				_ = greenLines.SetValues(off)
			}
		}
	}
}

func initButton() {

	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}

	offset := rpi.J8p7
	l, err := c.RequestLine(offset,
		gpiod.WithPullUp,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(buttonCallback))
	if err != nil {
		panic(err)
	}

	err = c.Close()
	if err != nil {
		return
	}
	defer func(l *gpiod.Line) {
		err := l.Close()
		if err != nil {
			return
		}
	}(l)

	for {
		time.Sleep(10 * time.Second)
	}
}

func buttonCallback(event gpiod.LineEvent) {

	if event.Type == gpiod.LineEventFallingEdge {
		buttonChan <- time.Now()
		return
	}

	if event.Type == gpiod.LineEventRisingEdge {
		select {
		case timePressed, ok := <-buttonChan:
			if ok {
				tDiff := time.Now().Sub(timePressed)
				if tDiff > time.Second*3 {
					ledChan <- ledReq{
						color: "all",
						mode:  "off",
					}

					ledChan <- ledReq{
						color: "red",
						mode:  "all",
					}
					time.Sleep(time.Second)
					ledChan <- ledReq{
						color: "yellow",
						mode:  "all",
					}
					time.Sleep(time.Second)
					ledChan <- ledReq{
						color: "green",
						mode:  "all",
					}
				}

			} else {
				fmt.Println("Channel closed!")
			}
		default:
			return
		}
	}
}

func main() {
	wg.Add(1)

	go initButton()
	go lights()
	ledChan <- ledReq{
		color: "all",
		mode:  "off",
	}

	go getMessage()
	go processVmixMessage()

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(vmixClient.conn)
	defer close(vmixMessageChan)
	defer close(ledChan)
	defer close(buttonChan)

	wg.Wait()

}
