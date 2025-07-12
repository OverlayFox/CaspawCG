package types

import (
	"fmt"
	"strings"
	"unicode"
)

// ParseCommandLine splits a command line string into arguments, handling quoted sections
// and escape characters. It supports both single and double quotes.
func ParseCommandLine(line string) ([]string, error) {
	if line == "" {
		return []string{}, nil
	}

	var parts []string
	currentPart := &strings.Builder{}
	inQuote := rune(0) // Stores the quote character (', ", or 0 if not in a quote)
	escaped := false   // True if the last character was an escape character

	for _, r := range line {
		if escaped {
			currentPart.WriteRune(r)
			escaped = false
			continue
		}

		switch {
		case r == '\\':
			if err := handleEscapeCharacter(currentPart, inQuote); err != nil {
				return nil, err
			}
			escaped = true
		case inQuote != 0:
			if err := handleQuotedCharacter(currentPart, r, &inQuote); err != nil {
				return nil, err
			}
		case unicode.IsSpace(r):
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		case r == '"' || r == '\'':
			inQuote = r
		default:
			currentPart.WriteRune(r)
		}
	}

	// Add the last part if it exists
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	//Validate final state
	if inQuote != 0 {
		return nil, fmt.Errorf("unmatched quote: %q", inQuote)
	}
	if escaped {
		return nil, fmt.Errorf("trailing escape character")
	}

	return parts, nil
}

// handleEscapeCharacter processes escape characters based on quote context
func handleEscapeCharacter(currentPart *strings.Builder, inQuote rune) error {
	if inQuote == 0 {
		currentPart.WriteRune('\\')
	}
	return nil
}

// handleQuotedCharacter processes characters within quoted sections
func handleQuotedCharacter(currentPart *strings.Builder, r rune, inQuote *rune) error {
	if r == *inQuote {
		*inQuote = 0 // End of quoted section
	} else {
		currentPart.WriteRune(r)
	}
	return nil
}
