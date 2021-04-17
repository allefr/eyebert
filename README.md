# Spectronix Eye-BERT MicroX - CLI

Python and Go (golang) CLI to interface with Spectronix Eye-BERT MicroX bit-error-rate (BER) testers.

## Getting Started

### Python Environment Setup

create a Python workspace and work isolated using:

- [pyenv](https://github.com/pyenv/pyenv)
- [pyenv-virtualenv](https://github.com/pyenv/pyenv-virtualenv)
- [pipenv](https://pypi.org/project/pipenv/)

Inside the `./Python/` folder of this project, ensure virtual environment has all needed packages by typing:

```Shell
pipenv sync
```

Run virtual environment:

```Shell
pipenv shell
```

Now on the left of your command line should be present a `(python)` prefix, identifying the correct execution.

## Usage

### Running the Python package

The python package can either be used standalone, or imported.

#### Run Standalone

Without embedding the package elsewhere, `eyeBERT` can be run standalone by calling the `eyeBERT.py` file itself. The user _must_ provide the serial port connected to the tester as argument. On Linux or MacOs the correct port can be found by typing

```Shell
ls /dev/tty*
```

before and after connecting the USB port of the tester to the computer; after connecting the tester, a new port is displayed, and tha is what we need to provide as argument. An example for MacOs is `/dev/tty.usbmodem142101`.

For example, once inside the virtual environment, typing

```Shell
python ./eyeBERT.py "/dev/tty.usbmodem142101"
```

will run the CLI in its default configuration. The help for providing a custom configuration can be triggered by typing

```Shell
python ./eyeBERT.py -h
```

which yields the following results:

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
  -s, --silent          no printed output
  -j, --json            output as json
```

Based on the above, the user can force the test to run at a specified datarate, pattern, and by defining a polling frequency for getting the test statistics. For example, if want to run at 10.2Gbps (on a support SFP+) using `PRBS31` as BERT pattern and by polling stats at 5Hz, the following call is used:

```Shell
python ./eyeBERT.py "/dev/tty.usbmodem142101" -r10.2 -pPRBS31 -f5
```

If no custom `datarate` or `wavelen` is provided, the tester will figure out the correct value based on the DDM information inside the SFP's eeprom connected to the tetser.
The default BERT pattern is `PRBS7`.

When providing the `-j` or `--json` flag, the test will run outputting any data in JSON format. For example, if we want to store all test results in josn format to file, the test can be launched as follow:

```Shell
python ./eyeBERT.py "/dev/tty.usbmodem142101" -j 2>&1 | tee BERTresult.txt
```

##### Interaction while test is running

Once the test is running, the user can interact with it by typing any of the following and press the `Enter` key:

- `q` or `quit` to quit the test
- `r` or `reset` to trigger clearing all current stats
- `sfp` to get info
- `tx on` or `tx off` to turn SFP tx power on or off respectively. Setting `tx off` will obviously prevent the test from running since no data will be output
- `rate` followed by "space" and the value in Gbps (e.g. `rate 1.25` to set 1.25 Gpbs). This will automatically trigger a test reset
- `p` followed by "space" and the pattern that is wanted (e.g. `p PRBS31`). This will automatically trigger a test reset

#### Run as an Imported Package

The package can be imported with

```Python
import eyeBERT
```

An object of the `EyeBERT_MicroX` class can be created by providing the serial port and optionally

- `wavelen=` wavelength in nm (default `None` - tester will use value from SFP)
- `datarate=` datarate in Gbps (default `None` - tester will use value from SFP)
- `pattern=` datarate as string (see [Supported-BERT-patterns](#Supported-BERT-patterns)) (default `PRBS7`)
- `noOutput=` boolean flag to prevent any print to stdout (default `True`)
- `useJson=` boolean flag to force any output formatted as JSON (default `False`). Meaningless if `noOutput=True`

When using all default values, the tester object can be created as follow

```Python
tester = eyeBERT.EyeBERT_MicroX("/dev/tty.usbmodem142101")
```

which will also reset all test stats.
To fetch test statistics, the following command can be used

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
- `status`: showing any warnings or `ok` is all is good

The user is welcome to make use or print this values as he/she pleases, thus the reason to use the `noOutput=True` flag to prevent any unwelcome prints.
Similarly, can get dictionaries for the following other commands:

- `getBERTinfo()`
- `getSFPinfo()`

The settings for the test can be changed with

- `setBERTwaveLength(wl)`: sets wavelength `wl` in nm
- `setBERTdataRate(dr)`: sets datarate `dr` in Gbps
- `setBERTpattern(pattern)`: sets pattern (see [Supported-BERT-patterns](#Supported-BERT-patterns))
- `setSFPtxEnable(on)`: turns SFP tx On or Off using boolean `on`
- `BERTrestartTest()`: resets all test stats, effectively re-starting the BER test

The EyeBERT tester also support a quick test mode triggered by `BERTrunQuickTest()`, which will run through the min and max datarates supported by the inserted SFP, and some additional datarates in between those.

#### Supported BERT patterns

The eyeBERT tester supports the following pattern:

- PRBS 2^7 -1 :arrow_right: `eyeBERT.pattPRBS7` or simply `"PRBS7"`
- PRBS 2^9 -1 :arrow_right: `eyeBERT.pattPRBS9` or simply `"PRBS9"`
- PRBS 2^11 -1 :arrow_right: `eyeBERT.pattPRBS11` or simply `"PRBS11"`
- PRBS 2^15 -1 :arrow_right: `eyeBERT.pattPRBS15` or simply `"PRBS15"`
- PRBS 2^23 -1 :arrow_right: `eyeBERT.pattPRBS23` or simply `"PRBS23"`
- PRBS 2^31 -1 :arrow_right: `eyeBERT.pattPRBS31` or simply `"PRBS31"`
- PRBS 2^58 -1 :arrow_right: `eyeBERT.pattPRBS58` or simply `"PRBS58"`
- PRBS 2^63 -1 :arrow_right: `eyeBERT.pattPRBS63` or simply `"PRBS63"`

and additionally, can be used to set the following test modes:

- low power :arrow_right: `eyeBERT.pattLowPow` or simply `"LOW_POW"`
- loop back :arrow_right: `eyeBERT.pattLoopBk` or simply `"LOOPBACK"`

## License

This project is licensed under the [MIT License :copyright: allefr](LICENSE)
