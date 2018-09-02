package hd44780

import "time"

const (
	// Timing constants
	//ePulse = 1 * time.Microsecond
	//eDelay = 70 * time.Microsecond
	ePulse = 2 * time.Microsecond
	eDelay = 140 * time.Microsecond

	// Some defaults
	lcdWidth = 20 // Maximum characters per line
	lcdChr   = true
	lcdCmd   = false

	lcdLine1 = 0x80 // LCD RAM address for the 1st line
	lcdLine2 = 0xC0 // LCD RAM address for the 2nd line
        lcdLine3 = 0x94 // LCD RAM address for the 3rd line
        lcdLine4 = 0xD4 // LCD RAM address for the 4th line
)
