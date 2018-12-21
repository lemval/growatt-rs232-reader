# growatt-rs232-reader

This reader allows some Growatt Inverters to publish data on a REST endpoint.

Example output of ```http://127.0.0.1:5701/status```:
```json
{
  "VoltagePV1": 335.8,
  "VoltagePV2": 0,
  "VoltageBus": 380.6,
  "VoltageGrid": 225.7,                   *
  "TotalProduction": 3822.6,              *
  "DayProduction": 0.7,                   *
  "Frequency": 49.99,                     *
  "Power": 798.4,                         *
  "Temperature": 26.9,                    *
  "OperationHours": 1891.338,             *
  "Status": "Normal",
  "FaultCode": 0,
  "Timestamp": "2018-12-09T13:15:54.363021599+01:00"
}
```
Starred fields are optional.

Information can also be retrieved: ```curl http://localhost:5701/info```:

```json
{
  "Reader":"Reading since 14:11:18",
  "Interpreter":"No input since 14:11:18",
  "Publisher":"Normal"
}
```

Above is a perfectly legal state as long as the times are within 10 minutes of the current time. Note that startup takes several seconds.

## Status

Attempting to get it restarted the next morning as on power down (sunset) the interface of the inverter will reset, so the init needs to be resend as soon as the inverter comes back to life. This seems to need polling, as no sign is yet detected which indicates it is powered on again.

## How to build and run:

It is written in GoLang. For compilation, see the build.sh script.

Run as ```./growatt --device /dev/ttyUSB0 --baudrate 9600``` or without any arguments to use the default shown.
Currently, stop bits, parity, etc. are fixed.

If you want to initialise the inverter manually, use ```./growatt --action Init```.
 
## Required

You need a 'USB to serial' converter. Remove the little plate to expose the RS2323 port and connect the cable. Connect the USB-side to a Raspberry Pi or other device. Using a Raspberry the serial output should NOT be activated (raspi-config).

## Openhab

If you would like to use this as Openhab growatt publisher, use this in combination with the HTTP binding: https://www.openhab.org/addons/bindings/http1/

In your ```services/http.cfg```, add the following:
```
growatt.url=http://127.0.0.1:5701/status
growatt.updateInterval=5000
```

Create an ```items/growatt.items``` with the following content:
```
Number Growatt_PvVoltage       "PV voltage [%.1f V]"           { http="<[growatt:5000:JSONPATH($.VoltagePV1)]" }
Number Growatt_GridVoltage     "Net voltage [%.1f V]"          { http="<[growatt:5000:JSONPATH($.VoltageGrid)]" }
Number Growatt_TotalProduction "Production overall [%.1f kWh]" { http="<[growatt:5000:JSONPATH($.TotalProduction)]" }
Number Growatt_DayProduction   "Production today: [%.1f kWh]"  { http="<[growatt:5000:JSONPATH($.DayProduction)]" }
Number Growatt_Frequency       "Net frequency [%.1f Hz]"       { http="<[growatt:5000:JSONPATH($.Frequency)]" }
Number Growatt_Power           "Current power [%.1f W]"        { http="<[growatt:5000:JSONPATH($.Power)]" }
String Growatt_Status          "Status [%s]"                   { http="<[growatt:5000:JSONPATH($.Status)]" }
String Growatt_Update          "Updated: [%s]"                 { http="<[growatt:5000:JSONPATH($.Timestamp)]" }
```

Install the JSsonPath transformation plugin.

Note that on power down of the inverter there will be values missing from the JSON, causing messages in the openhab log. As far as I know there is no way to indicate a field is optional for JSONPATH.


## Disclaimer

No, this isn't production ready quality code. See License.

The init string for the converter is fixed. 

## License:

*MIT*

Copyright 2018 lemval@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
