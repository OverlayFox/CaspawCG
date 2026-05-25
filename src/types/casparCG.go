package types

import (
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
	return casparTypes.MixerParamsFill{
		X:      float32(s.PosX) / float32(baseRes.Width),  // Convert to 0-1 range // x=100 on a 1920 width should be 100/1920 = 0.052
		Y:      float32(s.PosY) / float32(baseRes.Height), // Convert to 0-1 range // y=50 on a 1080 height should be 50/1080 = 0.046
		XScale: float32(s.SizeX) / 100,                    // Convert percentage to 0-1 range // sizeX=50 should be 0.5
		YScale: float32(s.SizeY) / 100,                    // Convert percentage to 0-1 range // sizeY=50 should be 0.5
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
	PushCGData(template string, layer, channel int, data map[string]any, sizing Sizing) error
	StopCGData(template string, layer, channel int) error
}
