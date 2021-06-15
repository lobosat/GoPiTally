package main

import (
	gpio "PiTally/gpio"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
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
	VmixIP      string `json:"ip"`
	RedType     string `json:"red_type"`
	RedValue    string `json:"red_value"`
	YellowType  string `json:"yellow_type"`
	YellowValue string `json:"yellow_value"`
	GreenType   string `json:"green_type"`
	GreenValue  string `json:"green_value"`
}

type vmState struct {
	Input     string
	Bus       map[string]string
	Streaming int
	Recording int
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
			panic(err)
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
			processVmixAPIResponse(line, vmixClient.tallyCfg)
		} else {
			fmt.Println("Error in GetMessage.ReadString: ", err)
			// most likely cause is that the connection to vMix went away (EOF).  Try to reconnect
			if strings.Contains(err.Error(), "EOF") {
				// The socket closed.  Most likely scenario is that communication to vMix was broken.
				// Restart the vMix service to attempt to re-connect
				cmd := exec.Command("sudo", "systemctl", "restart", "pitally")
				_ = cmd.Start()
				wg.Done()
				return
			} else {
				// Something else happened.  Exit pitally gracefully
				wg.Done()
				return
			}
		}
	}
}

func processVmixAPIResponse(response string, tallyCfg tally) {
	var state int

	if len(response) < 1 {
		return
	}

	messageSlice := strings.Fields(response)

	// ex:  [ACTS OK InputPlaying 9 1]
	// messageSlice[2] - Action
	// messageSlice[3] - Input
	// messageSlice[4] - State (usually 0 for off, 1 for on)
	if messageSlice[0] == "ACTS" && messageSlice[1] == "OK" {

		if messageSlice[2] == "Input" {
			if tallyCfg.RedType == "Input" || tallyCfg.YellowType == "Input" ||
				tallyCfg.GreenType == "Input" {

				state, _ = strconv.Atoi(messageSlice[4])

				if tallyCfg.RedType == "Input" && messageSlice[3] == tallyCfg.RedValue {

					if state == 0 {
						gpio.Leds("red", "off")
					}
					if state == 1 {
						gpio.Leds("red", "on")
					}
				}

				if tallyCfg.YellowType == "Input" && messageSlice[3] == tallyCfg.YellowValue {

					if state == 0 {
						gpio.Leds("yellow", "off")
					}
					if state == 1 {
						gpio.Leds("yellow", "on")
					}
				}

				if tallyCfg.GreenType == "Input" && messageSlice[0] == "ACTS" && messageSlice[1] == "OK" &&
					messageSlice[2] == "Input" && messageSlice[3] == tallyCfg.GreenValue {

					if state == 0 {
						gpio.Leds("green", "off")
					}
					if state == 1 {
						gpio.Leds("green", "on")
					}
				}
			}
		}

		matched, _ := regexp.Match("Bus.Audio", []byte(messageSlice[2]))

		if matched {
			if tallyCfg.RedType == "Bus" || tallyCfg.YellowType == "Bus" ||
				tallyCfg.GreenType == "Bus" {
				state, _ = strconv.Atoi(messageSlice[3])

				if tallyCfg.RedType == "Bus" {
					if messageSlice[2] == tallyCfg.RedType+tallyCfg.RedValue+"Audio" {

						if state == 0 {
							gpio.Leds("red", "off")
						}
						if state == 1 {
							gpio.Leds("red", "on")
						}
					}
				}

				if tallyCfg.YellowType == "Bus" {
					if messageSlice[2] == tallyCfg.YellowType+tallyCfg.YellowValue+"Audio" {

						if state == 0 {
							gpio.Leds("yellow", "off")
						}
						if state == 1 {
							gpio.Leds("yellow", "on")
						}
					}
				}

				if tallyCfg.GreenType == "Bus" {
					if messageSlice[2] == tallyCfg.GreenType+tallyCfg.GreenValue+"Audio" {

						if state == 0 {
							gpio.Leds("green", "off")
						}
						if state == 1 {
							gpio.Leds("green", "on")
						}
					}
				}
			}
		}

		if messageSlice[2] == "Streaming" {
			if tallyCfg.RedType == "Streaming" || tallyCfg.YellowType == "Streaming" ||
				tallyCfg.GreenType == "Streaming" {

				state, _ = strconv.Atoi(messageSlice[3])

				if tallyCfg.RedType == messageSlice[2] {
					if state == 0 {
						gpio.Leds("red", "off")
					}
					if state == 1 {
						gpio.Leds("red", "on")
					}
				}

				if tallyCfg.YellowType == messageSlice[2] {
					if state == 0 {
						gpio.Leds("yellow", "off")
					}
					if state == 1 {
						gpio.Leds("yellow", "on")
					}
				}

				if tallyCfg.GreenType == messageSlice[2] {
					if state == 0 {
						gpio.Leds("green", "off")
					}
					if state == 1 {
						gpio.Leds("green", "on")
					}
				}
			}
		}

		if messageSlice[2] == "Recording" {
			if tallyCfg.RedType == "Recording" || tallyCfg.YellowType == "Recording" ||
				tallyCfg.GreenType == "Recording" {

				state, _ = strconv.Atoi(messageSlice[3])

				if tallyCfg.RedType == messageSlice[2] {
					if state == 0 {
						gpio.Leds("red", "off")
					}
					if state == 1 {
						gpio.Leds("red", "on")
					}
				}

				if tallyCfg.YellowType == messageSlice[2] {
					if state == 0 {
						gpio.Leds("yellow", "off")
					}
					if state == 1 {
						gpio.Leds("yellow", "on")
					}
				}

				if tallyCfg.GreenType == messageSlice[2] {
					if state == 0 {
						gpio.Leds("green", "off")
					}
					if state == 1 {
						gpio.Leds("green", "on")
					}
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
					err := cmd.Start()
					if err != nil {
						return
					}
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
	var cfg tally

	// Check for /boot/tally_config.json.  If it exists move it to /usr/local/etc/pitally
	if _, err := os.Stat("/boot/tally_config.json"); err == nil {
		fmt.Println("Found a Tally Config file in /boot. Moving to /usr/local/etc/pitally")
		// Open original file
		originalFile, err := os.Open("/boot/tally_config.json")
		if err != nil {
			log.Fatal(err)
		}

		// Create new file
		newFile, err := os.Create("/usr/local/etc/pitally/tally_config.json")
		if err != nil {
			log.Fatal(err)
		}

		// Copy the bytes to destination from source
		_, err = io.Copy(newFile, originalFile)
		if err != nil {
			log.Fatal(err)
		}

		// Commit the file contents
		// Flushes memory to disk
		err = newFile.Sync()
		if err != nil {
			log.Fatal(err)
		}

		err = originalFile.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = newFile.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove("/boot/tally_config.json")
		if err != nil {
			log.Fatal(err)
		}
	}

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
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

// getVmixState will create a connection to the vMix API and query it to update the
// vMix state variables with the current configuration
func getVmixState(vmixIP string) *vmState {
	var vc = new(vmixClients)
	vc.vmixIP = vmixIP
	_ = vmixAPIConnect(vc)
	var vmixState = new(vmState)

	_, err := vc.w.WriteString("XML\r\n")
	if err == nil {
		err = vc.w.Flush()
	}
	var xml string
	var cont bool
	for cont = true; cont; {
		line, _ := vc.r.ReadString('\r')
		if strings.Contains(line, "<vmix>") {
			xml = xml + line
		}
		if strings.Contains(line, "</vmix>") {
			xml = xml + line
			cont = false
		}
	}

	_ = vc.conn.Close()

	doc := etree.NewDocument()
	_ = doc.ReadFromString(xml)

	streaming := doc.FindElement("/vmix/streaming").Text()
	if streaming == "True" {
		vmixState.Streaming = 1
	} else {
		vmixState.Streaming = 0
	}
	recording := doc.FindElement("/vmix/recording").Text()
	if recording == "True" {
		vmixState.Recording = 1
	} else {
		vmixState.Recording = 0
	}

	active := doc.FindElement("/vmix/active").Text()
	vmixState.Input = active

	var bus = make(map[string]string)

	var nameMap = map[string]string{
		"busA":   "A",
		"busB":   "B",
		"busC":   "C",
		"busD":   "D",
		"busE":   "E",
		"busF":   "F",
		"busG":   "G",
		"master": "M",
	}

	for _, audio := range doc.FindElements("./vmix/audio/*") {
		muted := audio.SelectAttrValue("muted", "")
		if muted == "True" {
			bus[nameMap[audio.Tag]] = "0"
		}
		if muted == "False" {
			bus[nameMap[audio.Tag]] = "1"
		}
	}
	vmixState.Bus = bus
	fmt.Println("Active Input: ", vmixState.Input)
	fmt.Println("Buses: ", vmixState.Bus)
	fmt.Println("Streaming: ", vmixState.Streaming)
	fmt.Println("Recording: ", vmixState.Recording)

	return vmixState
}

// setInitState will set the tally lights according to the
// contents of a vmState struct
func setInitState(state vmState, tallyCfg tally) {

	if tallyCfg.RedType == "Input" || tallyCfg.YellowType == "Input" || tallyCfg.GreenType == "Input" {
		if state.Input == tallyCfg.RedValue {
			gpio.Leds("red", "on")
		}
		if state.Input == tallyCfg.YellowValue {
			gpio.Leds("yellow", "on")
		}
		if state.Input == tallyCfg.GreenValue {
			gpio.Leds("green", "on")
		}
	}

	if tallyCfg.RedType == "Bus" || tallyCfg.YellowType == "Bus" || tallyCfg.GreenType == "Bus" {
		for bus, value := range state.Bus {
			if tallyCfg.RedValue == bus {
				if value == "0" {
					gpio.Leds("red", "off")
				}
				if value == "1" {
					gpio.Leds("red", "on")
				}
			}
			if tallyCfg.YellowValue == bus {
				if value == "0" {
					gpio.Leds("yellow", "off")
				}
				if value == "1" {
					gpio.Leds("yellow", "on")
				}
			}
			if tallyCfg.GreenValue == bus {
				if value == "0" {
					gpio.Leds("green", "off")
				}
				if value == "1" {
					gpio.Leds("green", "on")
				}
			}
		}
	}

	if tallyCfg.RedType == "Streaming" {
		if state.Streaming == 1 {
			gpio.Leds("red", "on")
		}
		if state.Streaming == 0 {
			gpio.Leds("red", "off")
		}
	}
	if tallyCfg.YellowType == "Streaming" {
		if state.Streaming == 1 {
			gpio.Leds("yellow", "on")
		}
		if state.Streaming == 0 {
			gpio.Leds("yellow", "off")
		}
	}
	if tallyCfg.GreenType == "Streaming" {
		if state.Streaming == 1 {
			gpio.Leds("green", "on")
		}
		if state.Streaming == 0 {
			gpio.Leds("green", "off")
		}
	}

	if tallyCfg.RedType == "Recording" {
		if state.Recording == 1 {
			gpio.Leds("red", "on")
		}
		if state.Recording == 0 {
			gpio.Leds("red", "off")
		}
	}
	if tallyCfg.YellowType == "Recording" {
		if state.Recording == 1 {
			gpio.Leds("yellow", "on")
		}
		if state.Recording == 0 {
			gpio.Leds("yellow", "off")
		}
	}
	if tallyCfg.GreenType == "Recording" {
		if state.Recording == 1 {
			gpio.Leds("green", "on")
		}
		if state.Recording == 0 {
			gpio.Leds("green", "off")
		}
	}
}

func main() {

	tallyCfg := getConfig()

	wg.Add(1)

	go initButton()

	gpio.Leds("all", "on")
	time.Sleep(time.Second * 3)
	gpio.Leds("all", "off")

	//Get current vmix state and set LEDs accordingly
	vmixState := getVmixState(tallyCfg.VmixIP)
	setInitState(*vmixState, tallyCfg)

	//Connect to the vmix API
	var vmixClient = new(vmixClients)
	vmixClient.vmixIP = tallyCfg.VmixIP
	vmixClient.vmixMessageChan = make(chan string)
	vmixClient.tallyCfg = tallyCfg

	err := vmixAPIConnect(vmixClient)
	if err != nil {
		fmt.Println("Error connecting to vmix API: ", err)
		panic(err)
	}

	go getMessage(vmixClient)

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(vmixClient.conn)
	defer close(vmixClient.vmixMessageChan)
	defer close(buttonChan)

	wg.Wait()
	fmt.Println("PiTally service exiting")
	gpio.Leds("all", "off")
}
