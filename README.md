# Spectronix Eye-BERT MicroX - CLI

Python and Go (golang) CLI to interface with Spectronix Eye-BERT MicroX bit-error-rate (BER) testers.

## Getting Started

Look at each implementation to see what is needed:

- [python](./python/README.md)
- [go](./go/README.md)

## Usage

Both implementations can either be used standalone, or imported.

The user _must_ provide the serial port connected to the tester as argument.

On Linux or MacOs the correct port can be found by typing

```Shell
ls /dev/tty*
```

before and after connecting the USB port of the tester to the computer; after connecting the tester, a new port is displayed. On MacOs, it is usually in the form `/dev/tty.usbmodem142101`.

On Windows, after making sure the correct driver is installed (check Spectronix website), the serial port is detected as a `COM` port, e.g. `COM3`. The Device Manager can be used to see what number is assigned.

Additional settings can be provided, depending on the implementation:

- [python](./python/README.md)
- [go](./go/README.md)

## License

This project is licensed under the [MIT License :copyright: allefr](LICENSE)
