## PBrelay Subscriber

A collection of subscribers for [pbrelay](https://github.com/psychobummer/pbrelay)

## Installation

`go get github.com/psychobummer/pbrelay-subscriber`

NOTE: there are some issues with the way the btle package used by `x/bluetooth` is bundled; if you get dependency errors when `go mod` tries to resolve them, just `go get tinygo.org/x/bluetooth` and you should be good to go.

## Usage

```
$ go build
$ ./pbrelay-subscriber
Subscribe to events from pbrelay; do various things with them

Usage:
  pbrelay-subscriber [command]

Available Commands:
  help        Help about any command
  midi-btle   Read MIDI data, and control local BTLE devices

Flags:
  -h, --help   help for pbrelay-subscriber

Use "pbrelay-subscriber [command] --help" for more information about a command.
```

## Producers

* [midi-btle](https://github.com/psychobummer/pbrelay-subscriber/blob/master/cmd/midibtle.go) Consume MIDI data from pbrelay, and use it to control local BTLE devices through PsychoBummer's [buttworx](https://github.com/psychobummer/buttwork) library. See the [buttworx README.md](https://github.com/psychobummer/buttwork/blob/master/README.md) for some BTLE-specific caveats/solutions.