package types

import (
	"fmt"
	"strings"
	"unicode"
)

// ParseCommandLine splits a string into arguments, handling quoted parts.
func ParseCommandLine(line string) ([]string, error) {
	var (
		parts       []string
		currentPart strings.Builder
		inQuote     rune // Stores the quote character (', ", or 0 if not in a quote)
		escaped     bool // True if the last character was an escape character
	)

	for _, r := range line {
		if escaped {
			currentPart.WriteRune(r)
			escaped = false
			continue
		}

		switch {
		case r == '\\': // Escape character
			escaped = true
			if inQuote != 0 {
				currentPart.WriteRune(r)
			}

		case inQuote != 0: // Inside a quoted section
			if r == inQuote { // Found the closing quote
				inQuote = 0 // Exit quoted state
			} else {
				currentPart.WriteRune(r) // Add character to the current part
			}

		case unicode.IsSpace(r):
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}

		case r == '"' || r == '\'': // Start of a quoted section
			inQuote = r

		default: // Regular character outside quotes
			currentPart.WriteRune(r)
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	if inQuote != 0 {
		return nil, fmt.Errorf("unmatched quote: %q", inQuote)
	}
	if escaped {
		return nil, fmt.Errorf("trailing escape character")
	}

	return parts, nil
}
