package main

import (
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

/*
TODO: coooooool icon?

TODO: CLI parsing with:
  - Select serial device, baud
  - Help menu
  - Keyboard bridge mode or Clipboard bridge mode

TODO: CLI menu implementation with:
  - 'Press q to quit'
  - 'Check out the repo at URLHERE'
  - '<Timestamp> - Value scanned'
*/

func main() {
	log.Println("Welcome to Protospace's RFID Reader Tool")

	// openSerial("/dev/ttyUSB0", 2400)
	// openSerial("COM5", 2400)
	// pressKeys()
	// getKeys()
	clipboard.WriteAll("Hello FUCKAAASsdsadsada:w ")
}

// TODO: dummy serial implementation
// TODO: pass in a channel that takes strings...
func dummySerial() {
	// every record from this particular scanner - or maybe it is the cards? - starts with 10 (LF) and ends with 13 (CR)
	buf := []byte{
		10, 51, 48, 48, 48, 70, 68, 51, 54, 56, 48, 13,
		10, 51, 48, 48, 48, 66, 70, 70, 70, 67, 49, 13,
	}
	for _, v := range buf {
		log.Printf("%d is a %v\n", v, v)
	}
}

// TODO: pass in a channel that takes strings...
func openSerial(device string, baud int) {
	serialDevice := device
	config := &serial.Config{
		Name: serialDevice,
		Baud: baud,
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("Failed to open port to: ", err)
	}

	buf := make([]byte, 102)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Fatal("Failed to read from port: ", err)
		}
		// s := string(buf[:n])
		// log.Printf("Read %d bytes : %s\n", n, s)
		log.Printf("Read %d bytes : %d\n", n, buf[:n])
	}
}

func pressKeys() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
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

func getKeys() {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal("Failed to open keyboard tty: ", err)
	}
	defer tty.Close()

	for {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal("Failed to read from keyboard tty: ", err)
		}

		// quit on user pressing 'q'
		if r == 'q' {
			os.Exit(0)
		}
		fmt.Println("Key press => " + string(r))
	}
}
