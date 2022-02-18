package main

import (
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/mattn/go-tty"
	"github.com/micmonay/keybd_event"
	"github.com/tarm/serial"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"
)

// default device/baud is Protospace offices RFID scanner
var defaultDevice = "COM5"
var defaultBaud = 2400
var testMode = false

// every record from PS's office scanner - or maybe it is the cards? - starts with LF and ends with CR
var startCharacter = byte(10) // ASCII LF
var endCharacter = byte(13)   // ASCII CR

const DEV_ENDPOINT string = "https://api.spaceport.dns.t0.vc/stats/autoscan/"
const PROD_ENDPOINT string = "https://my.protospace.ca/stats/autoscan/"

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

	// set the endpoint based on environment
	var endpoint string
	// set up a channel to transmit bytes from serial device to the "aggregator" function
	scanPipe := make(chan byte)
	// set up a channel to let main know if it is safe to continue or now
	proceedChan := make(chan bool)
	if testMode {
		fmt.Println("Test mode activated! Using a simulated device. Happy developing.")
		fmt.Println("")
		defaultDevice = "Test Simulator"
		endpoint = DEV_ENDPOINT
		go dummySerial(scanPipe, proceedChan)
	} else {
		endpoint = PROD_ENDPOINT
		go openSerial(defaultDevice, defaultBaud, scanPipe, proceedChan)
	}
	if !<-proceedChan {
		fmt.Println("Something went wrong connecting the serial device.")
		fmt.Println("Please read above and troubleshoot")
		waitForExitKey('q')
		os.Exit(1)
		return
	}
	fmt.Println("Successfully connected to serial device '" + defaultDevice + "'.")

	// TODO: generalize: bridge w/ pipe so we can setup via config instead of hardcoding
	// this might require making a factory for each bridge that returns the bridge function and a channel? factoring can accept configuration (e.g. endpoint for spaceport API, deduplication parameters, etc)
	clipboardBridgePipe := make(chan string)
	spaceportAPIBridgePipe := make(chan string)
	go scanAggregatorDuplicator(scanPipe, clipboardBridgePipe, spaceportAPIBridgePipe)

	go clipboardBridge(clipboardBridgePipe)
	go spaceportAPIBridge(endpoint, spaceportAPIBridgePipe)
	// keyboardBridge()

	fmt.Println("Begin scanning!")
	waitForExitKey('q')
}

// dummySerial returns hardcoded serial bytes for local development and testing
// Harded coded bytes are pushed into toAggregator channel
func dummySerial(toAggregator chan<- byte, proceed chan<- bool) {
	// set random seed for generating random numbers
	rand.Seed(time.Now().UnixNano())

	// the serial device wont fail, so we are good to proceed
	proceed <- true

	dummyReadings := [][]byte{
		[]byte{10, 51, 52, 53, 54, 71, 65, 56, 54, 56, 48, 13},
		[]byte{10, 51, 48, 48, 49, 66, 70, 70, 70, 67, 49, 13},
		[]byte{10, 51, 57, 55, 69, 66, 66, 66, 66, 66, 75, 72, 79, 66, 49, 13},
	}

	// send simulated data forever
	for {
		for _, reading := range dummyReadings {
			// simulate multiple scans to check debounce/deduplication behaviour
			for i := 1; i < rand.Intn(8)+1; i++ {
				// send each byte from reading
				for _, val := range reading {
					toAggregator <- val
				}
			}
			// send a scan at regular intervals
			time.Sleep(5 * time.Second)
		}
	}
}

// openSerial will read directly from a serial device and push bytes into toAggregator channel
func openSerial(device string, baud int, toAggregator chan<- byte, proceed chan<- bool) {
	serialDevice := device
	config := &serial.Config{
		Name: serialDevice,
		Baud: baud,
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		fail("Failed to open port to ", device, " with err:", err)
		proceed <- false
		return
	}

	// broadcast successful connection to device
	proceed <- true
	buf := make([]byte, 128)

	// read forever. put all data into the aggregator pipe
	for {
		n, err := stream.Read(buf)
		if err != nil {
			fail("Failed to read from port: ", err)
			return
		}
		for _, v := range buf[:n] {
			toAggregator <- v
		}
	}
}

// scanAggregatorDuplicator will read from the serial device, aggregate results and send the result to each bridge
// this function just aggregates bytes, it does nothing to deduplicate multiple scans
// pass off deduplication to bridge functions
// deduplication requirements are dictated by each bridge
func scanAggregatorDuplicator(fromSerial <-chan byte, bridges ...chan<- string) {
	var result string
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
				fail("Serial scan is far too long - is baud set properly?")
				return
			}
		}

		// implement one-to-many - send result to each bridge
		for _, bridge := range bridges {
			bridge <- result
		}
	}
}

// timeElapsedDebounce implements a time-based and value based debounce
// it returns a function that takes a scan result and returns a bool indicating if that scan result was duplicated within the provided debounceTimeout
func timeElapsedDebounce(debounceTimeout time.Duration) func(string) bool {
	var lastResult string
	var lastResultTime time.Time
	return func(result string) bool {
		// debounce/deduplication
		// if the current result is same as last
		// AND the elapsed time is less then out timeout
		// it is a duplicated reading within debounceTimeout
		isDuplicated := result == lastResult && time.Since(lastResultTime) < debounceTimeout
		lastResult = result
		lastResultTime = time.Now()
		return isDuplicated
	}
}

// spaceportAPIBridge will POST scans to the spaceport API
func spaceportAPIBridge(endpoint string, fromSerial <-chan string) {
	var result string
	debounceTimeout, _ := time.ParseDuration("1s")
	debouncer := timeElapsedDebounce(debounceTimeout)
	for {
		result = <-fromSerial

		if debouncer(result) {
			continue
		}

		// set POST parameters
		v := url.Values{}
		v.Set("autoscan", result)

		// POST to API
		resp, err := http.PostForm(endpoint, v)
		if err != nil {
			fail("Failed to sent to API: ", err)
			return
		}
		fmt.Println("Scan sent to Spaceport API: " + result + " -> " + resp.Status)
	}
}

// clipboardBridge will aggregate serial bytes into coherent records and send them to the users clipboard
func clipboardBridge(fromSerial <-chan string) {
	var result string
	var err error
	for {
		result = <-fromSerial
		// opting not to implement debounce here
		// because we overwrite the clipboard, multiple scans are idempotent
		// debounce will make the console output nicer maybe
		// but the functionality isn't improved

		// copy the result to clipboard and notify user
		// BUG: if you pass string([]byte) as result, clipboard.WriteAll will silently fail if []byte contains empty elements
		err = clipboard.WriteAll(result)
		if err != nil {
			fail("Failed to write to clipboard: ", err)
			return
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
		fail("Failed to open keyboard tty: ", err)
		return
	}
	// close teletype when the function ends
	defer tty.Close()

	fmt.Println("Press " + string(exitKey) + " to exit")
	fmt.Println("")

	// wait forever for that one key
	for {
		r, err := tty.ReadRune()
		if err != nil {
			fail("Failed to read from keyboard tty: ", err)
			return
		}

		// quit on user pressing 'q'
		if r == exitKey {
			os.Exit(0)
		}
	}
}

func fail(message ...interface{}) {
	fmt.Println(message...)
	fmt.Println("")
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

// keyboardBridge will simulate key presses
func keyboardBridge() {
	// TODO: Implement Keyboard bridge mode
	panic("NOT IMPLEMENTED")

	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		fail("Failed to construct keyboard: ", err)
		return
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
		fail("Failed to press keys: ", err)
		return
	}
}
