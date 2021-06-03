package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type vmixClients struct {
	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Reader
	sync.Mutex
	connected       bool
	vmixIP          string
	vmixMessageChan chan string
	tallyCfg        tally
}

type tally struct {
	Action string `json:"tally_action"` // Bus or Input
	Value  string `json:"tally_value"`
	IP     string `json:"ip"`
}

var wg sync.WaitGroup
var buttonChan = make(chan time.Time, 10)

// vmixAPIConnect connects to the vMix API. By default, the vMix API is on port 8099.
// If vMix is not up, this function will continue trying to connect, and will
// block until a connection is achieved.
func vmixAPIConnect(vmixClient *vmixClients) error {

	vmixClient.connected = false
	for vmixClient.connected == false {
		timeout := time.Second * 5
		conn, err := net.DialTimeout("tcp", vmixClient.vmixIP+":8099", timeout)

		if err == nil {
			vmixClient.conn = conn
			vmixClient.w = bufio.NewWriter(conn)
			vmixClient.r = bufio.NewReader(conn)
			vmixClient.connected = true
			leds("all", "off")
		} else if strings.Contains(err.Error(), "connection timed out") ||
			strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "i/o timeout") {

			fmt.Println("vmix api is inaccessible.  Probably because vMix is not running")
			fmt.Println("Waiting 5 seconds and trying again")
			vmixClient.connected = false
			leds("all", "off")
			leds("yellow", "on")
			time.Sleep(time.Second * 5)
		} else {
			fmt.Println("Unable to connect. Error was: ", err)
			return err
		}
	}
	return nil
}

func SendMessage(vmixClient *vmixClients, message string) error {

	vmixClient.Lock()
	pub := fmt.Sprintf("%v\r\n", message)
	_, err := vmixClient.w.WriteString(pub)
	if err == nil {
		err = vmixClient.w.Flush()
	}
	vmixClient.Unlock()
	return err
}

func getMessage(vmixClient *vmixClients) {

	// Subscribe to the activator feed in the vMix API
	err := SendMessage(vmixClient, "SUBSCRIBE ACTS")
	if err != nil {
		fmt.Println("Error in GetMessage.SendMessage: ", err)
		wg.Done()
	}

	//Capture all responses from the vMix API
	for {
		line, err := vmixClient.r.ReadString('\n')

		if err == nil {
			vmixClient.vmixMessageChan <- line
		} else {
			fmt.Println("Error in GetMessage.ReadString: ", err)
			// most likely cause is that the connection to vMix went away (EOF).  Try to reconnect
			wg.Done()
			return
		}
	}
}

func processVmixMessage(vmixClient *vmixClients) {
	var state int
	for {
		vmixMessage := <-vmixClient.vmixMessageChan
		if len(vmixMessage) < 1 {
			return
		}

		messageSlice := strings.Fields(vmixMessage)

		// ex:  [ACTS OK InputPlaying 9 1]
		// messageSlice[2] - Action
		// messageSlice[3] - Input
		// messageSlice[4] - State (usually 0 for off, 1 for on)

		if vmixClient.tallyCfg.Action == "Input" && messageSlice[0] == "ACTS" && messageSlice[1] == "OK" &&
			messageSlice[2] == "Input" && messageSlice[3] == vmixClient.tallyCfg.Value {
			state, _ = strconv.Atoi(messageSlice[4])

			if state == 0 {
				leds("red", "off")
			}
			if state == 1 {
				leds("red", "on")
			}
		}

		if vmixClient.tallyCfg.Action == "Bus" {
			if messageSlice[2] == vmixClient.tallyCfg.Action+vmixClient.tallyCfg.Value+"Audio" {
				state, _ = strconv.Atoi(messageSlice[3])
				if state == 0 {
					leds("red", "off")
				}
				if state == 1 {
					leds("red", "on")
				}
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
					leds("all", "off")
					leds("red", "on")
					time.Sleep(time.Second)
					leds("yellow", "on")
					time.Sleep(time.Second)
					leds("green", "on")

					cmd := exec.Command("sudo", "/usr/local/bin/doap.sh", "on")
					stdout, err := cmd.Output()
					if err != nil {
						fmt.Println("stderror:", cmd.Stderr)
						fmt.Println("String:", cmd.String())
						fmt.Println("exec", err.Error())
						fmt.Println("stdout:", stdout)
						return
					}
					fmt.Println("210 String:", cmd.String())
					fmt.Println("211 stdout:", stdout)

					leds("all", "off")
					leds("red", "on")
					leds("green", "on")

				} else {
					fmt.Println("Channel closed!")
				}
			}
		default:
			return
		}
	}
}

func getConfig() tally {
	var tallyCfg tally

	tallyFile, err := os.Open("/usr/local/etc/pitally/tally_config.json")
	if err != nil {
		panic(err)
	}
	defer func(tallyFile *os.File) {
		err := tallyFile.Close()
		if err != nil {
			panic(err)
		}
	}(tallyFile)

	byteValue, _ := ioutil.ReadAll(tallyFile)
	err = json.Unmarshal(byteValue, &tallyCfg)
	if err != nil {
		panic(err)
	}

	return tallyCfg
}

func leds(color, state string) {
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

func main() {
	tallyCfg := getConfig()

	wg.Add(1)

	go initButton()

	leds("all", "on")
	time.Sleep(time.Second * 3)
	leds("all", "off")

	var vmixClient = new(vmixClients)
	vmixClient.vmixIP = tallyCfg.IP
	vmixClient.vmixMessageChan = make(chan string)
	vmixClient.tallyCfg = tallyCfg

	//Connect to the vmix API
	err := vmixAPIConnect(vmixClient)
	if err != nil {
		fmt.Println("Error connecting to vmix API: ", err)
		close(buttonChan)
		close(vmixClient.vmixMessageChan)
		panic(err)
	}

	go getMessage(vmixClient)
	go processVmixMessage(vmixClient)

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(vmixClient.conn)
	defer close(vmixClient.vmixMessageChan)
	defer close(buttonChan)

	wg.Wait()
	fmt.Println("PiTally went boom!")

}
