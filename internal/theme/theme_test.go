package theme

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mihai-valentin/ccs/internal/db"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestGenerateSeedColor_Deterministic(t *testing.T) {
	a := GenerateSeedColor()
	b := GenerateSeedColor()
	if a != b {
		t.Errorf("GenerateSeedColor should be deterministic, got %q != %q", a, b)
	}
}

func TestGenerateSeedColor_ValidHex(t *testing.T) {
	got := string(GenerateSeedColor())
	matched, _ := regexp.MatchString(`^#[0-9A-F]{6}$`, got)
	if !matched {
		t.Errorf("GenerateSeedColor returned %q, not a #RRGGBB hex", got)
	}
}

func TestThemeColor_Persistence(t *testing.T) {
	d := setupTestDB(t)

	// Initially unset → falls back to seed color (deterministic, non-empty).
	first := GetThemeColor(d)
	if first == "" {
		t.Fatal("expected non-empty seed color")
	}

	// Set explicit → Get returns it.
	if err := SetThemeColor(d, "#7C3AED"); err != nil {
		t.Fatalf("SetThemeColor: %v", err)
	}
	got := GetThemeColor(d)
	if string(got) != "#7C3AED" {
		t.Errorf("GetThemeColor after set = %q, want %q", got, "#7C3AED")
	}

	// Reset → falls back to seed color again.
	if err := ResetThemeColor(d); err != nil {
		t.Fatalf("ResetThemeColor: %v", err)
	}
	after := GetThemeColor(d)
	if after != first {
		t.Errorf("after reset got %q, want seed %q", after, first)
	}
}

func TestHSLToRGB_EdgeCases(t *testing.T) {
	// Zero saturation → all channels equal to lightness.
	r, g, b := hslToRGB(0.5, 0, 0.4)
	if r != 0.4 || g != 0.4 || b != 0.4 {
		t.Errorf("s=0 should give grayscale=l, got %v/%v/%v", r, g, b)
	}
	// Red hue, full saturation, mid lightness → red-dominant RGB.
	r, g, b = hslToRGB(0, 1, 0.5)
	if r <= g || r <= b {
		t.Errorf("hue=0 should produce red-dominant, got r=%v g=%v b=%v", r, g, b)
	}
}
