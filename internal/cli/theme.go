package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mihai/ccs/internal/theme"
	"github.com/spf13/cobra"
)

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func newThemeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "theme",
		Short: "View or manage the UI accent color",
		Long:  "Show the current theme color, set a custom hex color, or reset to the auto-generated seed color.",
		RunE:  runThemeShow,
	}

	cmd.AddCommand(
		newThemeSetCmd(),
		newThemeResetCmd(),
	)

	return cmd
}

// runThemeShow displays the current theme color hex value and a colored sample block.
func runThemeShow(_ *cobra.Command, _ []string) error {
	d, err := openDB()
	if err != nil {
		return err
	}
	defer d.Close()

	color := theme.GetThemeColor(d)
	hex := string(color)

	swatch := lipgloss.NewStyle().
		Background(color).
		Padding(0, 2).
		Render("      ")

	fmt.Printf("Current theme color: %s  %s\n", hex, swatch)
	return nil
}

func newThemeSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <hex>",
		Short: "Set a custom theme color (e.g. #7C3AED)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			hex := args[0]
			if !strings.HasPrefix(hex, "#") {
				hex = "#" + hex
			}
			hex = strings.ToUpper(hex)

			if !hexColorRe.MatchString(hex) {
				return fmt.Errorf("invalid hex color %q — expected format #RRGGBB", args[0])
			}

			d, err := openDB()
			if err != nil {
				return err
			}
			defer d.Close()

			if err := theme.SetThemeColor(d, hex); err != nil {
				return fmt.Errorf("saving theme color: %w", err)
			}

			swatch := lipgloss.NewStyle().
				Background(lipgloss.Color(hex)).
				Padding(0, 2).
				Render("      ")

			fmt.Printf("Theme color set to %s  %s\n", hex, swatch)
			return nil
		},
	}
}

func newThemeResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset to the auto-generated seed color",
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := openDB()
			if err != nil {
				return err
			}
			defer d.Close()

			if err := theme.ResetThemeColor(d); err != nil {
				return fmt.Errorf("resetting theme color: %w", err)
			}

			color := theme.GenerateSeedColor()
			swatch := lipgloss.NewStyle().
				Background(color).
				Padding(0, 2).
				Render("      ")

			fmt.Printf("Theme color reset to seed %s  %s\n", string(color), swatch)
			return nil
		},
	}
}
