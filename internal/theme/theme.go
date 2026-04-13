package theme

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/user"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/mihai-valentin/ccs/internal/db"
)

const configKeyThemeColor = "theme_color"

// GenerateSeedColor produces a deterministic lipgloss.Color by hashing the
// current hostname, username, and OS. The hash selects an HSL hue while
// keeping saturation and lightness fixed for readability on dark terminals.
func GenerateSeedColor() lipgloss.Color {
	h := sha256.New()

	hostname, _ := os.Hostname()
	h.Write([]byte(hostname))

	if u, err := user.Current(); err == nil {
		h.Write([]byte(u.Username))
	}

	h.Write([]byte(runtime.GOOS))

	sum := h.Sum(nil)
	// Use first 2 bytes to pick a hue in [0, 360).
	hue := float64(binary.BigEndian.Uint16(sum[:2])%360) / 360.0
	return lipgloss.Color(hslToHex(hue, 0.65, 0.55))
}

// GetThemeColor returns the user's configured theme color from the database.
// If no color has been set, it falls back to GenerateSeedColor.
func GetThemeColor(d *db.DB) lipgloss.Color {
	hex, err := d.GetConfig(configKeyThemeColor)
	if err != nil || hex == "" {
		return GenerateSeedColor()
	}
	return lipgloss.Color(hex)
}

// SetThemeColor persists the given hex color string (e.g. "#7C3AED") as the
// user's theme color.
func SetThemeColor(d *db.DB, hex string) error {
	return d.SetConfig(configKeyThemeColor, hex)
}

// ResetThemeColor removes the stored theme color so the seed-based default
// is used again.
func ResetThemeColor(d *db.DB) error {
	return d.DeleteConfig(configKeyThemeColor)
}

// hslToHex converts HSL values (h, s, l each in [0,1]) to a "#RRGGBB" string.
func hslToHex(h, s, l float64) string {
	r, g, b := hslToRGB(h, s, l)
	return fmt.Sprintf("#%02X%02X%02X",
		int(math.Round(r*255)),
		int(math.Round(g*255)),
		int(math.Round(b*255)),
	)
}

// hslToRGB converts HSL values (h, s, l each in [0,1]) to RGB in [0,1].
func hslToRGB(h, s, l float64) (float64, float64, float64) {
	if s == 0 {
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r := hueToRGB(p, q, h+1.0/3.0)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-1.0/3.0)
	return r, g, b
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}
