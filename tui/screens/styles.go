package screens

import (
	"charm.land/lipgloss/v2"
	"github.com/CheeziCrew/curd"
)

// Gruyere uses the cyan/blue palette.
var (
	palette = curd.GruyerePalette
	st      = palette.Styles()
)

// Re-export colors that screen files reference directly.
var (
	colorFg      = curd.ColorFg
	colorGray    = curd.ColorGray
	colorGreen   = curd.ColorGreen
	colorRed     = curd.ColorRed
	colorCyan    = curd.ColorCyan
	colorBrCyan  = curd.ColorBrCyan
	colorBlue    = curd.ColorBlue
	colorBrBlue  = curd.ColorBrBlue
)

// Re-export common styles.
var (
	titleStyle    = st.Title
	subtitleStyle = st.Subtitle
	inputBox      = st.InputBox
	helpStyle     = st.HelpMargin
	accentStyle   = st.AccentStyle
	dimStyle      = lipgloss.NewStyle().Foreground(colorGray)
)
