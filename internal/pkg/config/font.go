package config

import (
	"fmt"
	"os"

	"github.com/golang/freetype/truetype"
)

func LoadFont(filename string) (*truetype.Font, error) {
	ttf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	font, err := truetype.Parse(ttf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	return font, nil
}
