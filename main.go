package main

import (
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/mattn/go-tty"
	"github.com/micmonay/keybd_event"
	"github.com/tarm/serial"
	"log"
	"os"
	"runtime"
	"time"
)

// TODO: coooooool icon?

// default device/baud is Protospace offices RFID scanner
var defaultDevice = "COM5"
var defaultBaud = 2400
var testMode = false

// every record from PS's office scanner - or maybe it is the cards? - starts with LF and ends with CR
var startCharacter = byte(10) // ASCII LF
var endCharacter = byte(13)   // ASCII CR

func main() {
	fmt.Println("Welcome to Protospace's RFID Reader Tool")
	fmt.Println("")
	fmt.Println("Visit the repository page for more information and support:")
	fmt.Println("\thttps://github.com/Protospace/rfid-reader")
	fmt.Println("")

	flag.StringVar(&defaultDevice, "device", defaultDevice, "Set name of device (Windows) or path to it (POSIX)")
	flag.IntVar(&defaultBaud, "baud", defaultBaud, "Set the baud of serial device")
	flag.BoolVar(&testMode, "test", testMode, "Set test mode, which simulates a serial device instead of requiring connecting to a real one")
	flag.Parse()

  // set up a channel to transmit bytes from serial device to the "aggregator" function
	scanPipe := make(chan byte)
	if testMode {
		fmt.Println("Test mode activated! Using a simulated device. Happy developing.")
		fmt.Println("")
		defaultDevice = "Test Simulator"
		go dummySerial(scanPipe)
	} else {
		go openSerial(defaultDevice, defaultBaud, scanPipe)
	}
	fmt.Println("Successfully connected to serial device '" + defaultDevice + "'.")
	fmt.Println("Begin scanning!")

	go clipboardBridge(scanPipe)
	// TODO: Implement Keyboard bridge mode and allow user to select it instead of clipboard bridge
	// pressKeys()

	waitForExitKey('q')
}

// dummySerial returns hardcoded serial bytes for local development and testing
// Harded coded bytes are pushed into toAggregator channel
func dummySerial(toAggregator chan<- byte) {
	dummyReadings := [][]byte{
		[]byte{10, 51, 52, 53, 54, 71, 65, 56, 54, 56, 48, 13},
		[]byte{10, 51, 48, 48, 49, 66, 70, 70, 70, 67, 49, 13},
		[]byte{10, 51, 57, 55, 69, 66, 66, 66, 66, 66, 75, 72, 79, 66, 49, 13},
	}

	// send simulated data forever
	for {
		for _, reading := range dummyReadings {
			for _, val := range reading {
				toAggregator <- val
			}
			// send a scan at regular intervals
			time.Sleep(5 * time.Second)
		}
	}
}

// openSerial will read directly from a serial device and push bytes into toAggregator channel
func openSerial(device string, baud int, toAggregator chan<- byte) {
	serialDevice := device
	config := &serial.Config{
		Name: serialDevice,
		Baud: baud,
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("Failed to open port to: ", err)
	}

	buf := make([]byte, 128)

	// read forever. put all data into the aggregator pipe
	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Fatal("Failed to read from port: ", err)
		}
		for _, v := range buf[:n] {
			toAggregator <- v
		}
	}
}

// clipboardBridge will aggregate serial bytes into coherent records and sending them to the users clipboard so that they may use it
func clipboardBridge(fromSerial <-chan byte) {
	var result string
	var err error
	for {
    // parse through record until we have received the stop character
		for {
			v := <-fromSerial
			if v == endCharacter {
				// this scan is done, so stop collection
				break
			} else if v == startCharacter {
				// we are getting a new record
				// reinitialize our collection array and start collecting
				result = ""
				continue
			} else {
				// we have a valid character - build out the string
				result = result + string(v)
			}

			if len(result) > 1024 {
				log.Fatal("Serial scan is far too long - is the baud set properly?")
			}
		}
    // TODO: implement debounce? continue if current result is same as previous result and time elapsed is <500 ms?

    // copy the result to clipboard and notify user
		err = clipboard.WriteAll(result)
		if err != nil {
			log.Fatal("Failed to write to clipboard: ", err)
		}
		fmt.Println("Scan copied to clipboard: " + result)
	}
}

// waitForExitKey set a key that, when pressed, will exit the program
// This allows the user to quit the program with a non-zero status code and keep the terminal open
func waitForExitKey(exitKey rune) {
	// open teletype to keyboard
	tty, err := tty.Open()
	if err != nil {
		log.Fatal("Failed to open keyboard tty: ", err)
	}
	// close teletype when the function ends
	defer tty.Close()

	fmt.Println("Press " + string(exitKey) + " to exit")
	fmt.Println("")

	// wait forever for that one key
	for {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal("Failed to read from keyboard tty: ", err)
		}

		// quit on user pressing 'q'
		if r == exitKey {
			os.Exit(0)
		}
	}
}

// ascii_to_keydb_lookup maps ASCII codes to keybd Key definitions
// See this for more details: https://raw.githubusercontent.com/micmonay/keybd_event/master/keyboard.png
var ascii_to_keydb_lookup = map[int]int{
	// ascii / and 0-9
	47: keybd_event.VK_SP11,
	48: keybd_event.VK_0,
	49: keybd_event.VK_1,
	50: keybd_event.VK_2,
	51: keybd_event.VK_3,
	52: keybd_event.VK_4,
	53: keybd_event.VK_5,
	54: keybd_event.VK_6,
	55: keybd_event.VK_7,
	56: keybd_event.VK_8,
	57: keybd_event.VK_9,
	// handle uppercase ascii A-Z
	65: keybd_event.VK_A,
	66: keybd_event.VK_B,
	67: keybd_event.VK_C,
	68: keybd_event.VK_D,
	69: keybd_event.VK_E,
	70: keybd_event.VK_F,
	71: keybd_event.VK_G,
	72: keybd_event.VK_H,
	73: keybd_event.VK_I,
	74: keybd_event.VK_J,
	75: keybd_event.VK_K,
	76: keybd_event.VK_L,
	77: keybd_event.VK_M,
	78: keybd_event.VK_N,
	79: keybd_event.VK_O,
	80: keybd_event.VK_P,
	81: keybd_event.VK_Q,
	82: keybd_event.VK_R,
	83: keybd_event.VK_S,
	84: keybd_event.VK_T,
	85: keybd_event.VK_U,
	86: keybd_event.VK_V,
	87: keybd_event.VK_W,
	88: keybd_event.VK_X,
	89: keybd_event.VK_Y,
	90: keybd_event.VK_Z,
	// handle lowercase ascii a-z
	97:  keybd_event.VK_A,
	98:  keybd_event.VK_B,
	99:  keybd_event.VK_C,
	100: keybd_event.VK_D,
	101: keybd_event.VK_E,
	102: keybd_event.VK_F,
	103: keybd_event.VK_G,
	104: keybd_event.VK_H,
	105: keybd_event.VK_I,
	106: keybd_event.VK_J,
	107: keybd_event.VK_K,
	108: keybd_event.VK_L,
	109: keybd_event.VK_M,
	110: keybd_event.VK_N,
	111: keybd_event.VK_O,
	112: keybd_event.VK_P,
	113: keybd_event.VK_Q,
	114: keybd_event.VK_R,
	115: keybd_event.VK_S,
	116: keybd_event.VK_T,
	117: keybd_event.VK_U,
	118: keybd_event.VK_V,
	119: keybd_event.VK_W,
	120: keybd_event.VK_X,
	121: keybd_event.VK_Y,
	122: keybd_event.VK_Z,
}

// pressKeys will simulate key presses
// NOT IMPLEMENTED - implement for "keyboard bridge" mode
func pressKeys() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		log.Fatal("Failed to construct keyboard: ", err)
	}

	// For linux, it is very important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	fmt.Println("In 3 seconds we gonna talk...")
	time.Sleep(3 * time.Second)

	// Select keys to be pressed
	kb.SetKeys(
		[]int{
			keybd_event.VK_O,
			keybd_event.VK_M,
			keybd_event.VK_W,
			keybd_event.VK_2,
			keybd_event.VK_F,
			keybd_event.VK_Y,
			keybd_event.VK_G,
		}...,
	)

	// Press the selected keys
	err = kb.Launching()
	if err != nil {
		log.Fatal("Failed to press keys: ", err)
	}
}
