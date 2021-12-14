# RFID Reader

RFID Reader is made for Protospace's directors for setting up vetted user access to the space.
It will copy RFID card scans directly to the system clipboard so you can paste them where needed.

More generally, this console application bridges a serial device to your system clipboard.

## How To Run and Use this Program

### 1. Install Dependencies

This application was developed with:
- `go1.17.5 linux/amd64`
- `xclip version 0.13` (if you're on Linux)

Find `go` for your OS and architecture [here](https://go.dev/dl/)

If installing on linux, make sure you've installed `xclip` or `xsel` for [`github.com/atotto/clipboard`](https://github.com/atotto/clipboard)

Install golang libraries with `go mod download`

### 2. Build

`go build` will create an executable for your local platform.
Simply run that executable for further instructions.

If you want to build and deploy this to our office machine, which runs Windows, you can build for a specific architecture by setting `GOOS`.
e.g. `GOOS=windows go build`

### 3. Use

Simply run the executable.

On Linux: `./rfid-reader`

On Windows: `.\rfid-reader.exe`

Provide the `-h` flag for options available.
All defaults are designed for Protospaces office so this utility can be run without any modifications.

## Maintenance

Feel free to submit PRs and modify this package to your hearts content

If you need support on this and cannot code, please contact Kent Brockman through SpaceBar
