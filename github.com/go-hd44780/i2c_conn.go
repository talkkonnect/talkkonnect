package hd44780

import (
	"strings"
	//	"sync"

	//"github.com/kidoman/embd"
	//"github.com/kidoman/embd/controller/hd44780"
	"github.com/zlowred/embd"
	"github.com/zlowred/embd/controller/hd44780"

	// load only rpi
	//_ "github.com/kidoman/embd/host/rpi"
	_ "github.com/zlowred/embd/host/rpi"
)

// I2C4bit allow communicate wit HD44780 via I2C in 4bit mode
type I2C4bit struct {
	//	sync.Mutex

	// max lines
	Lines int
	// LCD width (number of character in line)
	Width int

	// i2c address
	addr      byte
	lastLines []string
	active    bool

	hd        *hd44780.HD44780
	backlight bool
}

// NewI2C4bit create new I2C4bit structure with some defaults
func NewI2C4bit(addr byte) (h *I2C4bit) {
	h = &I2C4bit{
		Lines: 2,
		addr:  addr,
		Width: lcdWidth,
	}
	return
}

// Open / initialize LCD interface
func (h *I2C4bit) Open() (err error) {
	//	h.Lock()
	//	defer h.Unlock()

	if h.active {
		return
	}

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}

	bus := embd.NewI2CBus(1)

	h.hd, err = hd44780.NewI2C(
		bus,
		h.addr,
		hd44780.PCF8574PinMap,
		hd44780.RowAddress16Col,
		hd44780.TwoLine,
		//		hd44780.BlinkOff,
		hd44780.CursorOff,
		hd44780.EntryIncrement,
	)
	if err != nil {
		return err
	}

	h.lastLines = make([]string, h.Lines, h.Lines)
	h.reset()

	h.hd.BacklightOn()
	h.backlight = true

	h.active = true

	return
}

// Active return true when interface is working ok
func (h *I2C4bit) Active() bool {
	return h.active
}

// Reset interface
func (h *I2C4bit) Reset() {
	//	h.Lock()
	//	defer h.Unlock()
	h.hd.Clear()
}

// Clear screen
func (h *I2C4bit) Clear() {
	//	h.Lock()
	//	defer h.Unlock()
	h.hd.Clear()
}

func (h *I2C4bit) reset() {
	// clear the display
	h.hd.Clear()
}

// Close interface, clear display.
func (h *I2C4bit) Close() {
	//	h.Lock()
	//	defer h.Unlock()

	if !h.active {
		return
	}
	h.hd.BacklightOff()
	h.backlight = false
	h.hd.Clear()
	embd.CloseI2C()
	h.active = false
}

// DisplayLines sends one or more lines separated by \n to lcd
func (h *I2C4bit) DisplayLines(msg string) {
	for line, text := range strings.Split(msg, "\n") {
		h.Display(line, text)
	}
}

// Display only one line
func (h *I2C4bit) Display(line int, text string) {
	//	h.Lock()
	//	defer h.Unlock()

	if !h.active {
		return
	}

	if line >= h.Lines {
		return
	}

	if !h.backlight {
		return
	}

	// skip not changed lines
	if h.lastLines[line] == text {
		return
	}

	h.lastLines[line] = text

	textLen := len(text)
	if textLen < lcdWidth {
		text = text + strings.Repeat(" ", h.Width-textLen)
	} else if textLen > h.Width {
		text = text[:h.Width]
	}

	h.hd.SetCursor(0, line)
	for _, c := range text {
		h.hd.WriteChar(byte(c))
	}
}

func (h *I2C4bit) ToggleBacklight() {
	if !h.active {
		return
	}
	if h.backlight {
		h.hd.BacklightOff()
		h.hd.Clear()
		h.hd.Home()
		h.backlight = false
	} else {
		h.hd.BacklightOn()
		h.backlight = true
		for l, line := range h.lastLines {
			h.lastLines[l] = ""
			h.Display(l, line)
		}
	}
}

func (h *I2C4bit) SetChar(pos byte, def []byte) {
	if len(def) != 8 {
		panic("invalid def - req 8 bytes")
	}
	h.hd.WriteInstruction(0x40 + pos*8)
	for _, d := range def {
		h.hd.WriteChar(d)
	}
}
