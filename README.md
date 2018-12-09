# growatt-rs232-reader

This reader allows some Growatt Inverters to publish data on a REST endpoint.

Example output of ```http://127.0..0.1:5701/status```:
```json
{
  "VoltagePV1": 335.8,
  "VoltagePV2": 0,
  "VoltageBus": 380.6,
  "VoltageGrid": 225.7,
  "TotalProduction": 3822.6,
  "DayProduction": 0.7,
  "Frequency": 49.99,
  "Power": 798.4,
  "Temperature": 26.9,
  "OperationHours": 1891.338,
  "Status": "Normal",
  "FaultCode": 0,
  "Timestamp": "2018-12-09T13:15:54.363021599+01:00"
}
```

## How to build and run:

It is written in GoLang. For compilation, see the build.sh script.

Run as ```./growatt /dev/ttyUSB0 9600``` or without any arguments to use the default shown.
Currently, web port, stop bits, parity, etc. are fixed.

## Required

You need a 'USB to serial' converter. Remove the little plate to expose the RS2323 port and connect the cable. Connect the USB-side to a Raspberry Pi or other device.

## Disclaimer

No, this isn't production ready quality code. See License.

The init string for the converter is fixed. 

## License:

*MIT*

Copyright 2018 lemval@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
