go-hd44780
==========

Simple library for Hitachi HD44780 LCD Display and Golang.

Dedicated for use in Raspberry Pi and lcd connected to GPIO-ports / I2C in 4-bit mode.

## Dependency

* github.com/stianeikeland/go-rpio
* github.com/zlowred/embd (github.com/kidoman/embd)

## Example

```go
lcd := hd44780.NewGPIO4bit()
if err := l.lcd.Open(); err != nil {
	panic("Can't open lcd: " + err.Error())
}
lcd.DisplayLines("line1\nline2")
lcd.Close()
```

For I2C use NewI2C4bit(addr).


## License 
Copyright (c) 2015 Karol Będkowski.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
