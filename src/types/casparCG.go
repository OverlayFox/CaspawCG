package types

import (
	"math"
	"time"

	casparTypes "github.com/overlayfox/casparcg-amcp-go/types"
)

type Sizing struct {
	PosX  int     `json:"posX"`
	PosY  int     `json:"posY"`
	SizeX float64 `json:"sizeX"`
	SizeY float64 `json:"sizeY"`
}

func (s Sizing) IsDefault() bool {
	if s.PosX == 0 && s.PosY == 0 && s.SizeX == 100 && s.SizeY == 100 {
		return true
	}
	return false
}

func (s Sizing) GetCasparMixerParams(baseRes Resolution) casparTypes.MixerParamsFill {
	round5 := func(v float64) float32 {
		return float32(math.Round(v*1e5) / 1e5)
	}

	return casparTypes.MixerParamsFill{
		X:      round5(float64(s.PosX) / float64(baseRes.Width)),  // Convert to 0-1 range // x=100 on a 1920 width should be 100/1920 = 0.052
		Y:      round5(float64(s.PosY) / float64(baseRes.Height)), // Convert to 0-1 range // y=50 on a 1080 height should be 50/1080 = 0.046
		XScale: round5(s.SizeX / 100),                             // Convert percentage to 0-1 range // sizeX=50 should be 0.5
		YScale: round5(s.SizeY / 100),                             // Convert percentage to 0-1 range // sizeY=50 should be 0.5
	}
}

type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (r Resolution) AspectRatio() float64 {
	if r.Height == 0 {
		return 0
	}
	return float64(r.Width) / float64(r.Height)
}

type CasparCGClient interface {
	Connect() error
	GetTemplates() ([]string, error)

	// Control functions for CG templates
	PushCGData(template string, layer, channel int, data map[string]any, sizing Sizing, delay time.Duration) error
	StopCGData(template string, layer, channel int, delay time.Duration) error
	ClearAll()
}
