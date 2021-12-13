package main

import (
	"fmt"
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
TODO: go.mod
TODO: readme with
  - dependency info
  - install build instructions (cover GOOS and GOARCH)
  - dev tips?

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
  openSerial()
	// pressKeys()
  getKeys()
}

// TODO: dummy serial implementation
func openSerial() {
	serialDevice := "/dev/hidraw1"
	config := &serial.Config{
		Name: serialDevice,
		Baud: 9600,
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("Failed to open port to: ", err)
	}

	buf := make([]byte, 1024)

	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Fatal("Failed to read from port: ", err)
		}
		s := string(buf[:n])
		fmt.Println(s)
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
