# Spectronix Eye-BERT MicroX - GoLang CLI

Go (golang) CLI to interface with Spectronix Eye-BERT MicroX bit-error-rate (BER) testers.

## Getting Started

Setup the Go environmnet on your Windows/Mac/Linux machine using <https://golang.org/doc/install>

## Usage

The user _must_ provide the serial port connected to the tester as argument.

On Linux or MacOs the correct port can be found by typing

```Shell
ls /dev/tty*
```

before and after connecting the USB port of the tester to the computer; after connecting the tester, a new port is displayed. On MacOs, it is usually in the form `/dev/tty.usbmodem142101`.

On Windows, after making sure the correct driver is installed (check Spectronix website), the serial port is detected as a `COM` port, e.g. `COM3`. The Device Manager can be used to see what number is assigned.

### Run Standalone

Inside the `go/` folder, can run the standalon version by typing

```Shell
go run main.go "/dev/tty.usbmodem142101"
```

which will run the CLI in its default configuration. Or can first build, then run the compiled file:

```Shell
go build main.go
./main "/dev/tty.usbmodem142101"
```

The help for providing a custom configuration can be triggered by typing

```Shell
./main -h
```

which yields the following result:

```
usage: ./main [-h] [-f <float>] [-j] [-p <string>] [-r <float>] port

positional arguments:
  port  serial port

optional arguments:
  -f float
        polling frequency [Hz] (default 1)
  -j    output as json
  -p string
        bert pattern as "PRBS<7|9|11|15|23|31|58|63>" (default "PRBS7")
  -r float
        datarate [Gbps] (default taken from sfp)
```

The user can force the test to run at a specified datarate, pattern, and by defining a polling frequency for getting the test statistics. For example, if want to run at 10.2Gbps (on a supported SFP+) using `PRBS31` as BERT pattern and by polling stats at 5Hz, the following call is used:

```Shell
./main  -r=10.2 -p=PRBS31 -f=5 "/dev/tty.usbmodem142101"
```

note that any argument provided _after_ `port` will be ignored. For example

```Shell
# this correctly sets pattern as PRBS31
./main -p PRBS31 "/dev/tty.usbmodem142101"
# this time instead will be ignored
./main "/dev/tty.usbmodem142101" -p PRBS31
```

If no custom `datarate` is provided, the tester will figure out the correct value based on the eeprom data of the connected SFP.
The default BER pattern is `PRBS7`.

When providing the `-j` flag, the test will run outputting data in JSON format. For example, if the user wants to store the test results in json format to file, the test can be launched as follow:

```Shell
./main -j "/dev/tty.usbmodem142101" 2>&1 | tee BERTresult.txt
```

### Run as an Imported Package

The package can be imported with

```Go
import ("github.com/allefr/eyebert/go/eyebert")
```

Create an interface with the tester with

```Go
  p := &eyebert.BERTParams{
    SerPort:  port,
    DataRate: rate, // [Gbps]
    Pattern:  eyebert.Pattern(patt),
  }
  tester, err := eyebert.New(p)
```

`SerPort` is required, while if `DataRate` and `Pattern` are not provided, the default will be used.

The methods exposed are:

- `GetTesterInfo()`, returning a [BERTester](#structs) struct and any errors
- `GetSFPinfo()`, returning a [SFPData](#structs) struct and any errors
- `StartTest()` to start the test, returning any errors
- `GetStats()`, returning a [BERStats](#structs) struct and any errors
- `ResetStats()` to reset stats for ongoing test, returning any errors
- `StopTest()` to stop the test, freezing the last stats, returning any errors
- `SetWaveLength(wl)` to set wavelegnth
- `SetDataRate(dr)` to set datarate `dr` in Gbps. This will effectively restart the stats
- `SetPattern(patt)` to set `patt` as BER pattern (see supported [here](#Supported-BERT-patterns))
- `SetSFPtxEnable(on)` to turn SFP tx On or Off as boolean. This can be used to pause (off) or un-pause (on) a test.
- `Close()` to close serial

#### Structs

The following structures are used as returned values

```Go
type BERTester struct {
  Manufacturer string
  Model        string
  Version      string
}

type BERStats struct {
  Datetime  string
  Duration  time.Duration
  BER       float64
  BitCnt    float64
  ErrCnt    float64
  RateGpbs  float32
  Pattern   string
  SFPrxPow  float32
  SFPtxPow  float32
  SFPtemp   float32
  EyeHorzUI float32
  EyeVertmV float32
  Status    string
}

type SFPData struct {
  Vendor       string
  PartNum      string
  SerialNum    string
  WaveLengthNm float32
  DistanceKm   float32
  MinRateGbps  float32
  MaxRateGbps  float32
  Temperature  float32
  RxPow        float32
  TxPow        float32
}
```

#### Supported BERT patterns

The eyeBERT tester supports the following Pseudo-Random Binary Sequence (PRBS) patterns:

- `eyebert.PattPRBS7` or `"PRBS7"`: PRBS 2^7 -1 (good for testing 1Gbps ethernet data)
- `eyebert.PattPRBS9` or `"PRBS9"`: PRBS 2^9 -1
- `eyebert.PattPRBS11` or `"PRBS11"`: PRBS 2^11 -1
- `eyebert.PattPRBS15` or `"PRBS15"`: PRBS 2^15 -1
- `eyebert.PattPRBS23` or `"PRBS23"`: PRBS 2^23 -1
- `eyebert.PattPRBS31` or `"PRBS31"`: PRBS 2^31 -1 (good for testing 10Gbps ethernet data)
- `eyebert.PattPRBS58` or `"PRBS58"`: PRBS 2^58 -1
- `eyebert.PattPRBS63` or `"PRBS63"`: PRBS 2^63 -1

and additionally, can be used to set the following test modes:

- `eyebert.PattLowPow` or `"LOW_POW"`: low power
- `eyebert.PattLoopBk` or `"LOOPBACK"`: loop back
