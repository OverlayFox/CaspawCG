package googleworkspace

import (
	"os"
	"path/filepath"
	"strings"
)

func getDJLogo(name string) (string, error) {
	// Remove all characters that are not letters, numbers, underscores, or hyphens
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' {
			return r
		}
		return -1
	}, name)
	logoFileName := "dj-logos/" + strings.ToLower(cleanName) + "_Logo_PNG.png"

	absLogoPath, err := filepath.Abs("../../placeholderclub/casparCG/template/images/" + logoFileName)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(absLogoPath); err == nil {
		return logoFileName, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	return "fallback.png", nil
}
