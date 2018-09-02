package hd44780

// Hitachi HD44780U support library

type HD44780 interface {
	Open() (err error)
	Reset()
	Close()
	Clear()
	Display(line int, text string)
	DisplayLines(msg string)
	Active() bool
	SetChar(pos byte, def []byte)
	ToggleBacklight()
}
