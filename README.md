# RFID Reader

## Dependencies

This application was developed with:
- `go1.17.5 linux/amd64`
- `xclip version 0.13` (if you're on Linux)

Find `go` for your OS and architecture [here](https://go.dev/dl/)

If installing on linux, make sure you've installed `xclip` or `xsel` for [`github.com/atotto/clipboard`](https://github.com/atotto/clipboard)

## Build and Run

`go build` will create an executable for your local platform.
Simply run that executable for further instructions.

If you want to build and deploy this to our office machine, which runs Windows, you can build for a specific architecture by setting `GOOS`.
e.g. `GOOS=windows go build`

## Maintenance

Feel free to submit PRs and modify this package to your hearts content

If you need support on this and cannot code, please contact Kent Brockman through SpaceBar
