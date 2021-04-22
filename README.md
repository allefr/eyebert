# Spectronix Eye-BERT MicroX - CLI

Python and Go (golang) CLI to interface with Spectronix Eye-BERT MicroX bit-error-rate (BER) testers.

## Getting Started

### Python Environment Setup

Create a Python workspace and work isolated using:

- [pyenv](https://github.com/pyenv/pyenv)
- [pyenv-virtualenv](https://github.com/pyenv/pyenv-virtualenv)
- [pipenv](https://pypi.org/project/pipenv/)

Inside the `./python/` folder of this project, ensure virtual environment has all needed packages by typing:

```Shell
pipenv sync
```

Run virtual environment:

```Shell
pipenv shell
```

Now on the left of your command line should be present a `(python)` prefix, identifying the correct execution.

## Usage

See Go implementation [here](./go/README.md)

### Running the Python package

The python package can either be used standalone, or imported.

The user _must_ provide the serial port connected to the tester as argument.

On Linux or MacOs the correct port can be found by typing

```Shell
ls /dev/tty*
```

before and after connecting the USB port of the tester to the computer; after connecting the tester, a new port is displayed. On MacOs, it is usually in the form `/dev/tty.usbmodem142101`.

On Windows, after making sure the correct driver is installed (check Spectronix website), the serial port is detected as a `COM` port, e.g. `COM3`. The Device Manager can be used to see what number is assigned.

#### Run Standalone

`eyeBERT` can be run standalone by calling the `eyeBERT.py` file itself, followed by the serial port.

Once inside the virtual environment, typing

```Shell
python ./eyeBERT.py "/dev/tty.usbmodem142101"
```

will run the CLI in its default configuration. The help for providing a custom configuration can be triggered by typing

```Shell
python ./eyeBERT.py -h
```

which yields the following result:

```
usage: eyeBERT.py [-h] [-r RATE] [-w WAVELEN] [-p PATTERN] [-f FREQUENCY] [-s] [-j] port

positional arguments:
  port                  serial port path

optional arguments:
  -h, --help            show this help message and exit
  -r RATE, --rate RATE  datarate in Gbps. default taken from sfp
  -w WAVELEN, --wavelen WAVELEN
                        wavelength in nm. default taken from sfp
  -p PATTERN, --pattern PATTERN
                        bert pattern as "PRBS<7|9|11|15|23|31|58|63>". default "PRBS7"
  -f FREQUENCY, --frequency FREQUENCY
                        polling frequency. default 1Hz
  -j, --json            output as json
```

Based on the above, the user can force the test to run at a specified datarate, pattern, and by defining a polling frequency for getting the test statistics. For example, if want to run at 10.2Gbps (on a supported SFP+) using `PRBS31` as BERT pattern and by polling stats at 5Hz, the following call is used:

```Shell
python ./eyeBERT.py "/dev/tty.usbmodem142101" -r10.2 -pPRBS31 -f5
```

If no custom `datarate` or `wavelen` is provided, the tester will figure out the correct value based on the eeprom data of the connected SFP.
The default BERT pattern is `PRBS7`.

When providing the `-j` or `--json` flag, the test will run outputting any data in JSON format. If the user wants to store all test results in json format to file, the test can be launched as follow:

```Shell
python ./eyeBERT.py "/dev/tty.usbmodem142101" -j 2>&1 | tee BERTresult.txt
```

##### Interaction while test is running

Once the test is running, the user can interact with it by typing any of the following and press the `<Enter>` key:

- `q` or `quit` to quit the test
- `r` or `reset` to trigger clearing all current stats
- `sfp` to get info from the connected SFP
- `tx on` or `tx off` to turn SFP tx power on or off respectively. Setting `tx off` will obviously prevent the test from running since no data will be output. This will have the effect of momentarily pause the test
- `rate` followed by `<space>` and the value in Gbps (e.g. `rate 1.25` to set 1.25 Gbps). This will automatically trigger a test reset
- `p` followed by `<space>` and the pattern that is wanted (e.g. `p PRBS31`). This will automatically trigger a test reset

#### Run as an Imported Package

The package can be imported with

```Python
import eyeBERT
```

An object of the `EyeBERT_MicroX` class can be created by providing the serial port and optionally

- `wavelen=` wavelength in nm (default `None` - tester will use value from SFP)
- `datarate=` datarate in Gbps (default `None` - tester will use value from SFP)
- `pattern=` datarate as string - see [Supported-BERT-patterns](#Supported-BERT-patterns) (default `PRBS7`)
- `noOutput=` boolean flag to prevent any print to stdout (default `True`)
- `useJson=` boolean flag to force any output formatted as JSON (default `False`). Meaningless if `noOutput=True`

When using all default values, the tester object can be created as follow

```Python
tester = eyeBERT.EyeBERT_MicroX("/dev/tty.usbmodem142101")
```

which immediately triggers the start of the test.
To fetch test progression, the following command can be used

```Python
stats = tester.BERTreadStats()
```

where `stats` is returned as a dictionary containing

- `datetime`: timestamp in RFC3339 format
- `duration`: elapsed time from last test reset
- `ber`: current value Bit Error Rate. Value is `-1` when any error occurs
- `bit-cnt`: increasing counter of the sent bits
- `err-cnt`: increasing counter of the received error bits
- `rate-gpbs`: datarate (Gbps) in use
- `pattern`: BERT pattern in use
- `sfp-rx-pow`: SFP Rx power (dBm)
- `sfp-tx-pow`: SFP Tx power (dBm)
- `sfp-temp`: SFP temperature (degC)
- `eye-horz`: horizontal eye diagram (UI)
- `eye-vert`: vertical eye diagram (mV)
- `status`: showing any warnings or `ok` if all is good

The user is welcome to make use or print these values as he/she pleases, thus the reason to use the `noOutput=True` flag to prevent any additional unwelcomed prints. If however `noOutput` is set to `False`, when calling the `BERTreadStats()` method, an output will instead be printed with minimal data inside. In this case the user just needs to call the method multiple times to see the progression of the test.

Similarly, can get dictionaries for the following other commands:

- `getBERTinfo()`: model and version of the tester
- `getSFPinfo()`: DDM values of the SFP inserted inside the tester

The settings for the test can be changed with

- `setBERTwaveLength(wl)`: sets wavelength `wl` in nm
- `setBERTdataRate(dr)`: sets datarate `dr` in Gbps
- `setBERTpattern(patt)`: sets `patt` as BERT pattern (see [Supported-BERT-patterns](#Supported-BERT-patterns))
- `setSFPtxEnable(on)`: turns SFP tx On or Off by providing `True` or `false`. This can be used to pause (off) or un-pause (on) a test.
- `BERTrestartTest()`: resets all test stats, effectively re-starting the BER test

The EyeBERT tester also support a quick test mode triggered by `BERTrunQuickTest()`, which will run through the min and max datarates supported by the inserted SFP, and some additional datarates in between those.

#### Supported BERT patterns

The eyeBERT tester supports the following Pseudo-Random Binary Sequence (PRBS) patterns:

- `eyeBERT.pattPRBS7` or `"PRBS7"`: PRBS 2^7 -1 (good for testing 1Gbps ethernet data)
- `eyeBERT.pattPRBS9` or `"PRBS9"`: PRBS 2^9 -1
- `eyeBERT.pattPRBS11` or `"PRBS11"`: PRBS 2^11 -1
- `eyeBERT.pattPRBS15` or `"PRBS15"`: PRBS 2^15 -1
- `eyeBERT.pattPRBS23` or `"PRBS23"`: PRBS 2^23 -1
- `eyeBERT.pattPRBS31` or `"PRBS31"`: PRBS 2^31 -1 (good for testing 10Gbps ethernet data)
- `eyeBERT.pattPRBS58` or `"PRBS58"`: PRBS 2^58 -1
- `eyeBERT.pattPRBS63` or `"PRBS63"`: PRBS 2^63 -1

and additionally, can be used to set the following test modes:

- `eyeBERT.pattLowPow` or `"LOW_POW"`: low power
- `eyeBERT.pattLoopBk` or `"LOOPBACK"`: loop back

## License

This project is licensed under the [MIT License :copyright: allefr](LICENSE)
