package output

import (
	"os"
	"regexp"
	"runtime"
)

// ANSI color constants
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"

	// Bright colors
	BrightRed    = "\033[91m"
	BrightGreen  = "\033[92m"
	BrightYellow = "\033[93m"
	BrightBlue   = "\033[94m"
	BrightPurple = "\033[95m"
	BrightCyan   = "\033[96m"
)

// ColorConfig holds color configuration
type ColorConfig struct {
	Enabled bool
}

// Global color configuration
var Colors = &ColorConfig{
	Enabled: isColorSupported(),
}

// isColorSupported checks if the terminal supports color output
func isColorSupported() bool {
	// Disable colors on Windows by default (can be overridden)
	if runtime.GOOS == "windows" {
		return false
	}

	// Check if we're in a TTY
	if !isTerminal(os.Stdout) {
		return false
	}

	// Check TERM environment variable
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return true
}

// isTerminal checks if the file is a terminal
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// SetColorsEnabled allows manual control of color output
func SetColorsEnabled(enabled bool) {
	Colors.Enabled = enabled
}

// Colorize applies color to text if colors are enabled
func Colorize(text, color string) string {
	if !Colors.Enabled {
		return text
	}
	return color + text + Reset
}

// Helper functions for common colors
func RedText(text string) string    { return Colorize(text, Red) }
func GreenText(text string) string  { return Colorize(text, Green) }
func YellowText(text string) string { return Colorize(text, Yellow) }
func BlueText(text string) string   { return Colorize(text, Blue) }
func PurpleText(text string) string { return Colorize(text, Purple) }
func CyanText(text string) string   { return Colorize(text, Cyan) }
func BoldText(text string) string   { return Colorize(text, Bold) }

// Bright color variants
func BrightRedText(text string) string    { return Colorize(text, BrightRed) }
func BrightGreenText(text string) string  { return Colorize(text, BrightGreen) }
func BrightYellowText(text string) string { return Colorize(text, BrightYellow) }
func BrightBlueText(text string) string   { return Colorize(text, BrightBlue) }
func BrightPurpleText(text string) string { return Colorize(text, BrightPurple) }
func BrightCyanText(text string) string   { return Colorize(text, BrightCyan) }

// ColorizeChange colors text based on change type
func ColorizeChange(text string, changeType string) string {
	switch changeType {
	case "added", "add", "+":
		return GreenText("+ " + text)
	case "removed", "remove", "drop", "-":
		return RedText("- " + text)
	case "modified", "modify", "change", "~":
		return YellowText("~ " + text)
	default:
		return text
	}
}

// ColorizeTableName highlights table names
func ColorizeTableName(tableName string) string {
	return CyanText(tableName)
}

// ColorizeColumnName highlights column names
func ColorizeColumnName(columnName string) string {
	return YellowText(columnName)
}

// ColorizeDataType highlights data types
func ColorizeDataType(dataType string) string {
	return PurpleText(dataType)
}

// ColorizeString highlights string values
func ColorizeString(str string) string {
	return BrightGreenText(str)
}

// ColorizeNumber highlights numeric values
func ColorizeNumber(num string) string {
	return BrightBlueText(num)
}

// ColorizeSQLStatement applies syntax highlighting to SQL statements with semantic colors
func ColorizeSQLStatement(statement string) string {
	if !Colors.Enabled {
		return statement
	}

	// Use regex-based approach to highlight only complete SQL keywords/phrases
	// This avoids highlighting partial matches inside strings or identifiers

	result := statement

	// Define keyword patterns with word boundaries
	patterns := []struct {
		pattern string
		color   func(string) string
	}{
		// Table operations
		{`\bALTER\s+TABLE\b`, BrightCyanText},
		{`\bCREATE\s+TABLE\b`, BrightCyanText},
		{`\bDROP\s+TABLE\b`, RedText},

		// Column operations
		{`\bADD\s+COLUMN\b`, GreenText},
		{`\bDROP\s+COLUMN\b`, RedText},
		{`\bMODIFY\s+COLUMN\b`, YellowText},
		{`\bCHANGE\s+COLUMN\b`, YellowText},

		// Index operations
		{`\bADD\s+INDEX\b`, GreenText},
		{`\bADD\s+UNIQUE\s+INDEX\b`, GreenText},
		{`\bADD\s+FULLTEXT\s+INDEX\b`, GreenText},
		{`\bDROP\s+INDEX\b`, RedText},

		// Key operations
		{`\bADD\s+PRIMARY\s+KEY\b`, GreenText},
		{`\bDROP\s+PRIMARY\s+KEY\b`, RedText},
		{`\bADD\s+FOREIGN\s+KEY\b`, GreenText},
		{`\bDROP\s+FOREIGN\s+KEY\b`, RedText},
		{`\bADD\s+CONSTRAINT\b`, GreenText},
		{`\bDROP\s+CONSTRAINT\b`, RedText},

		// Table options
		{`\bENGINE\s*=`, CyanText},
		{`\bDEFAULT\s+CHARSET\s*=`, CyanText},
		{`\bCHARSET\s*=`, CyanText},
		{`\bCOLLATE\s*=`, CyanText},
		{`\bCOMMENT\s*=`, CyanText},
		{`\bAUTO_INCREMENT\s*=`, CyanText},

		// Common SQL keywords (only when standalone)
		{`\bDEFAULT\b`, PurpleText},
		{`\bNULL\b`, PurpleText},
		{`\bNOT\s+NULL\b`, PurpleText},
		{`\bUNSIGNED\b`, PurpleText},
		{`\bZEROFILL\b`, PurpleText},
		{`\bAUTO_INCREMENT\b`, PurpleText},
		{`\bUNIQUE\b`, PurpleText},
	}

	// Apply each pattern
	for _, p := range patterns {
		re := regexp.MustCompile(`(?i)` + p.pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return p.color(match)
		})
	}

	return result
}
