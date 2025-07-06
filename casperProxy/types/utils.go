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
			// If inside a quote, the backslash might be part of the literal string.
			// If outside, it means the next character is literal.
			// For simplicity here, we're treating it as escaping the next char regardless.
			// A more robust parser would handle specific escape sequences (e.g., \n, \t).
			if inQuote != 0 {
				currentPart.WriteRune(r) // Keep the backslash if inside a quote for true JSON
			}

		case inQuote != 0: // Inside a quoted section
			if r == inQuote { // Found the closing quote
				inQuote = 0 // Exit quoted state
				// Don't add the quote character itself to the part
			} else {
				currentPart.WriteRune(r) // Add character to the current part
			}

		case unicode.IsSpace(r): // Outside quotes, encountered a space
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}

		case r == '"' || r == '\'': // Start of a quoted section
			inQuote = r // Set the quote type
			// Don't add the quote character itself to the part

		default: // Regular character outside quotes
			currentPart.WriteRune(r)
		}
	}

	// Add the last part if any characters remain
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	// Basic check for unclosed quotes at the end
	if inQuote != 0 {
		return nil, fmt.Errorf("unmatched quote: %q", inQuote)
	}
	if escaped {
		return nil, fmt.Errorf("trailing escape character")
	}

	return parts, nil
}
