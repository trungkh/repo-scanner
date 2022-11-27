package utstring

import (
	"math"
	"strings"
)

// Color type
type Color int

// ColorType type
type ColorType int

const (
	// FOREGROUND color type
	FOREGROUND ColorType = 1 + iota

	// BACKGROUND color type
	BACKGROUND
)

const (
	// DEFAULT color
	DEFAULT Color = 1 + iota

	// BLACK color
	BLACK

	// RED color
	RED

	// GREEN color
	GREEN

	// GREEN color
	YELLOW

	// BLUE color
	BLUE

	// MAGENTA color
	MAGENTA

	// CYAN color
	CYAN

	// LIGHT_GRAY color
	LIGHT_GRAY

	// DARK_GRAY color
	DARK_GRAY

	// LIGHT_RED color
	LIGHT_RED

	// LIGHT_GREEN color
	LIGHT_GREEN

	// LIGHT_YELLOW color
	LIGHT_YELLOW

	// LIGHT_BLUE color
	LIGHT_BLUE

	// LIGHT_MAGENTA color
	LIGHT_MAGENTA

	// LIGHT_CYAN color
	LIGHT_CYAN

	// WHITE color
	WHITE
)

// EscChar constant
var EscChar = "\x1B"

// ResetChar constant
var ResetChar = EscChar + "[0m"

// ProgressBarOption struct
type ProgressBarOption struct {
	Size       float64 `json:"size"`
	Max        float64 `json:"max"`
	Value      float64 `json:"value"`
	FullColor  bool    `json:"full_color"`
	ValueColor Color
	BackColor  Color
}

/**
 * Public function
 **/

// GetColorCode function
//
// **@Params:** [ `c`: color; `t`: type ]
//
// **@Return:** [ `$1`: color code; `$2`: status ]
func GetColorCode(c Color, t ColorType) (string, bool) {
	fcs := map[Color][]string{
		DEFAULT:       []string{"39", "49"},
		BLACK:         []string{"30", "40"},
		RED:           []string{"31", "41"},
		GREEN:         []string{"32", "42"},
		YELLOW:        []string{"33", "43"},
		BLUE:          []string{"34", "44"},
		MAGENTA:       []string{"35", "45"},
		CYAN:          []string{"36", "46"},
		LIGHT_GRAY:    []string{"37", "47"},
		DARK_GRAY:     []string{"90", "100"},
		LIGHT_RED:     []string{"91", "101"},
		LIGHT_GREEN:   []string{"92", "102"},
		LIGHT_YELLOW:  []string{"93", "103"},
		LIGHT_BLUE:    []string{"94", "104"},
		LIGHT_MAGENTA: []string{"95", "105"},
		LIGHT_CYAN:    []string{"96", "106"},
		WHITE:         []string{"97", "107"},
	}

	cc, ok := fcs[c]
	if ok {
		if t == FOREGROUND {
			return cc[0], true
		} else {
			return cc[1], true
		}
	}
	return "", false
}

// ApplyColor function
//
// **@Params:** [ `s`: string; `f`: fore color; `b`: back color ]
//
// **@Return:** [ `$1`: colored string ]
func ApplyColor(s string, f Color, b Color) string {
	cf, of := GetColorCode(f, FOREGROUND)
	cb, ob := GetColorCode(b, BACKGROUND)
	if !of && !ob {
		return s
	}

	val := EscChar + "["
	if of {
		val += cf
	}
	if ob {
		if of {
			val += ";"
		}
		val += cb
	}
	val += "m" + s + ResetChar

	return val
}

// ApplyForeColor function
//
// **@Params:** [ `s`: string; `c`: color ]
//
// **@Return:** [ `$1`: colored string ]
func ApplyForeColor(s string, c Color) string {
	col, ok := GetColorCode(c, FOREGROUND)
	if ok {
		return EscChar + "[" + col + "m" + s + ResetChar
	}
	return s
}

// ApplyBackColor function
//
// **@Params:** [ `s`: string; `c`: color ]
//
// **@Return:** [ `$1`: colored string ]
func ApplyBackColor(s string, c Color) string {
	col, ok := GetColorCode(c, BACKGROUND)
	if ok {
		return EscChar + "[" + col + "m" + s + ResetChar
	}
	return s
}

// RenderProgressBar function
//
// **@Params:** [ `o`: option ]
//
// **@Return:** [ `$1`: colored progress string ]
func RenderProgressBar(o ProgressBarOption) string {
	val := ""

	if o.Size <= 0 || o.Max <= 0 {
		return val
	}

	backChr, valChr, curChr := " ", " ", " "
	if !o.FullColor {
		val += ApplyForeColor("[", o.BackColor)
		o.Size -= 2
		backChr = " "
		valChr = "="
		curChr = ">"
	}

	p := o.Value / o.Max
	c := math.Floor(float64(o.Size) * float64(p))

	colorFnc := func(s string, c Color) string {
		if o.FullColor {
			return ApplyBackColor(s, c)
		} else {
			return ApplyForeColor(s, c)
		}
	}

	for i := float64(0); i < o.Size; i++ {
		if i < (c-1) || c >= o.Size {
			val += colorFnc(valChr, o.ValueColor)
		} else if c < o.Size && i == (c-1) {
			val += colorFnc(curChr, o.ValueColor)
		} else {
			val += colorFnc(backChr, o.BackColor)
		}
	}

	if !o.FullColor {
		val += ApplyForeColor("]", o.BackColor)
	}

	return val
}

// RenderCLICommand function
//
// **@Params:** [ `cmd`: command; `a`: arguments ]
//
// **@Return:** [ `$1`: formated string ]
func RenderCLICommand(cmd string, a string) string {
	a = strings.Trim(a, " ")
	if a != "" {
		a = " " + a
	}
	return WrapSingleQuote(ApplyForeColor(cmd+a, YELLOW))
}
