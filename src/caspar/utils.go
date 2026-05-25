package casparcg

import (
	"fmt"
	"strings"

	casparTypes "github.com/overlayfox/casparcg-amcp-go/types"
	"github.com/overlayfox/caspaw-cg/src/types"
)

func VideoModeToResolution(mode casparTypes.VideoMode) (types.Resolution, error) {
	modeStr := string(mode)
	switch {
	case strings.HasPrefix(modeStr, "720"):
		return types.Resolution{Width: 1280, Height: 720}, nil
	case strings.HasPrefix(modeStr, "1080"):
		return types.Resolution{Width: 1920, Height: 1080}, nil
	case strings.HasPrefix(modeStr, "PAL"):
		return types.Resolution{Width: 720, Height: 576}, nil
	default:
		return types.Resolution{}, fmt.Errorf("unsupported video mode: %s", mode)
	}
}
