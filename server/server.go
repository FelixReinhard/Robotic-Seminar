package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	CONN_HOST = "192.168.43.253" // "192.168.2.215" for hotspot or 192.168.43.253
	CONN_PORT = "1269"
	CONN_TYPE = "tcp"
	FILENAME  = "log.txt"
)

/*
DEFINITION:
	PIN 0: Front Thumb # PALM
	PIN 1: Front Middlefinger
	Pin 2: Front pinky
	Pin 3: (Front) Palm
	Pin 4: Back under Ring and middlefinger
	Pin 5: Back bottom under pinky

	High five:
		Touched: 0 1 2 3
	Holding side:
		Touched: 3 5 (0)
	Holding interlinked:
		Touched: 3 4
*/

/*
	func encryptMessage(key string, message string) string {
		c, err := aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
		}
		msgByte := make([]byte, len(message))
		c.Encrypt(msgByte, []byte(message))
		return hex.EncodeToString(msgByte)
	}

	func decryptMessage(key string, message string) string {
		txt, _ := hex.DecodeString(message)
		c, err := aes.NewCipher([]byte(key))
		if err != nil {
			fmt.Println(err)
		}
		msgByte := make([]byte, len(txt))
		c.Decrypt(msgByte, []byte(txt))

		msg := string(msgByte[:])
		return msg
	}
*/
type InputData struct {
	// Maps each sensor to a boolean value
	// meaning either being touched or not
	// Note that we use byte here to make it easier to parse
	touchSensors [12]byte
	// Perhaps other stuff here
}

func (d InputData) at(i int) bool {
	if i >= 0 && i <= 12 {
		return d.touchSensors[i] == 1
	}
	return false
}

func (d InputData) toString() string {
	var ret string = ""
	for _, e := range d.touchSensors {

		if e == 0 {
			ret += "0"
		} else {
			ret += "1"
		}
	}
	return ret
}

const (
	HIGH_FIVE           = 0
	HOLDING_SIDE        = 1
	HOLDING_INTERLINKED = 2
)

type OutputData struct {
	debug string
	// One of the things defined in the const above.
	mode int
}

func (d OutputData) toBytes() []byte {

	var h byte
	var s byte
	var i byte

	if d.mode == HIGH_FIVE {
		h = 1
	}
	if d.mode == HOLDING_SIDE {
		s = 1
	}
	if d.mode == HOLDING_INTERLINKED {
		i = 1
	}

	return []byte{h, s, i}
}

// Remember need to read in reverse order as writing
func parseInput(reader *bytes.Reader, log *Logger) *InputData {

	var data InputData
	var err error
	var b byte
	// Touch Sensor Len
	for i := 0; i < 12; i++ {
		b, err = reader.ReadByte()
		data.touchSensors[i] = b
		if err != nil {
			log.log("Error while parsing input data (Touch Sensor Len): " + err.Error())
			os.Exit(1)
		}
	}

	return &data
}

func main() {

	log := NewLogger()

	go log.loop()
	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)

	if err != nil {
		log.log("Error : " + err.Error())
		time.Sleep(2)
		os.Exit(1)
	}

	log.log(listener.Addr().String())
	defer listener.Close()

	dataChan := make(chan InputData, 100)
	stopChan := make(chan bool)
	httpInputDataChan := make(chan InputData)

	// Host http server
	go host(log, httpInputDataChan)

	// TODO 1 => 2
	for i := 0; i < 1; i++ {
		log.log("Waiting for Client " + strconv.Itoa(i) + "...")
		conn, err := listener.Accept()
		if err != nil {
			log.log("error" + err.Error())
			os.Exit(1)
		}
		log.log("Got connection " + conn.LocalAddr().Network())

		go handle(conn, log, dataChan, stopChan, httpInputDataChan)
	}
	r := bufio.NewReader(os.Stdin)
	for {
		text, _ := r.ReadString('\n')
		if text == "stop" {
			// stop other goroutines, always be 2
			stopChan <- true
			stopChan <- true
			return
		} else if text == "1" {
			// simulate a high five TODO
			dataChan <- InputData{}

		}
	}

}

func handle(conn net.Conn, log *Logger, c chan InputData, stopChan chan bool, http_c chan InputData) {
	isInput := auth(conn, log)

	defer conn.Close()

	// If input wait for data to be sent. After recv. Send one byte (1) to be ok and otherwise (0) to be an error
	//  then resent. if the byte is any other value (>1) then server wants to halt
	// A max of 1024
	if isInput {
		buffer := make([]byte, 12)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				log.log("Error reading from connection " + conn.LocalAddr().Network())
				os.Exit(1)
			}

			if n >= 1024 {
				log.log("Handle.getInput : Warning: Data limit for reading was maxed out. Maybe there is more data to be read")
			}

			stream := bytes.NewReader(buffer)
			data := parseInput(stream, log)

			// send data to another goroutine which is an output if it can get it otherwise just disregard the information.

			select {
			case c <- *data:
				{
				}
			default:
				{
				}
			}

			// Send the data also to the webserver if no data is in the channel
			select {
			case http_c <- *data:
				{
				}
			default:
				{
				}
			}

			// finally give ok back or if should stop sent a stop
			// No okay anymore
			select {
			case _ = <-stopChan:
				{
					// send a stop to the client
					b := []byte{2}
					conn.Write(b)
					return
				}
			default:
				{
					// Ok TEMP DONT DO THIS HERE FOR PERFORMANCE
					//b := []byte{1}
					//conn.Write(b)
				}
			}
		}
	} else {
		// First recv. data sent from the input. After that send it to the client.
		// then recv a 0 for everything is ok
		for {
			tmp := <-c

			select {
			case _ = <-stopChan:
				{
					// send a stop to the client
					log.log("Stopping recv: Stopping Server and client")
					b := []byte{2}
					conn.Write(b)
					return
				}
			default:
				{
				}
			}
			data := processData(&tmp, log).toBytes()

			ok := false
			b := make([]byte, 1)

			for !ok {
				_, err := conn.Write(data)
				if err != nil {
					log.log("Error writting to connection " + conn.LocalAddr().Network())
					continue
				}

				_, err = conn.Read(b)
				if err != nil {
					log.log("Error reading to connection " + conn.LocalAddr().Network())
					continue
				}

				ok = b[0] == 0
			}
			log.log("Successfully sent data to client " + conn.LocalAddr().Network())
		}

	}
}

func processData(data *InputData, log *Logger) OutputData {
	// Process perhaps

	// High five
	if data.at(0) && data.at(1) && data.at(2) && data.at(3) {
		return OutputData{mode: HIGH_FIVE, debug: data.toString()}
	}
	if data.at(3) && data.at(5) {
		return OutputData{mode: HOLDING_SIDE, debug: data.toString()}
	}
	if data.at(3) && data.at(4) {
		return OutputData{mode: HOLDING_INTERLINKED, debug: data.toString()}
	}
	return OutputData{debug: "NOMODE", mode: -1}
}

func auth(conn net.Conn, log *Logger) bool {
	log.log("Start Auth...")
	// Ask the mode
	conn.Write([]byte("m"))

	buffer := make([]byte, 1)

	// Read the type of the device.
	_, err := conn.Read(buffer)

	if err != nil {
		log.log("Error while Auth: \n Cannot read from connection")
		os.Exit(1)
	}

	isInput := false

	if buffer[0] >= 1 {
		isInput = true
	}

	// Auth

	log.log("Auth Ok!")
	if isInput {
		log.log("Registered Input!")
	} else {
		log.log("Registered Output!")
	}
	return isInput
}

type Logger struct {
	file  *os.File
	input chan string
}

func NewLogger() *Logger {
	f, err := os.OpenFile(FILENAME, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	l := Logger{file: f, input: make(chan string, 10)}
	return &l
}

func (l Logger) pure_log(text string) {
	fmt.Println(text)
	if _, err := l.file.WriteString(text + "\n"); err != nil {
		panic(err)
	}
}

func (l Logger) log(text string) {
	l.input <- text
}

func (l Logger) close() {
	l.file.Close()
}

func (l Logger) loop() {
	message := ""
	for true {
		select {
		case message = <-l.input:
			{
				l.pure_log(message)
			}
		}
	}
}
