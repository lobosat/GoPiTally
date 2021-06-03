package main

import (
	gpio "PiTally/gpio"
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
			gpio.Leds("all", "off")
		} else if strings.Contains(err.Error(), "connection timed out") ||
			strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "i/o timeout") {

			fmt.Println("vmix api is inaccessible.  Probably because vMix is not running")
			fmt.Println("Waiting 5 seconds and trying again")
			vmixClient.connected = false
			gpio.Leds("all", "off")
			gpio.Leds("yellow", "on")
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
				gpio.Leds("red", "off")
			}
			if state == 1 {
				gpio.Leds("red", "on")
			}
		}

		if vmixClient.tallyCfg.Action == "Bus" {
			if messageSlice[2] == vmixClient.tallyCfg.Action+vmixClient.tallyCfg.Value+"Audio" {
				state, _ = strconv.Atoi(messageSlice[3])
				if state == 0 {
					gpio.Leds("red", "off")
				}
				if state == 1 {
					gpio.Leds("red", "on")
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
					gpio.Leds("all", "off")
					gpio.Leds("red", "on")
					time.Sleep(time.Second)
					gpio.Leds("yellow", "on")
					time.Sleep(time.Second)
					gpio.Leds("green", "on")

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

					gpio.Leds("all", "off")
					gpio.Leds("red", "on")
					gpio.Leds("green", "on")

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

func main() {
	tallyCfg := getConfig()

	wg.Add(1)

	go initButton()

	gpio.Leds("all", "on")
	time.Sleep(time.Second * 3)
	gpio.Leds("all", "off")

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
