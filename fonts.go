package main 

import (
	"os"
	"bytes"
	"github.com/benoitkugler/textlayout/fonts/truetype"
)

type Fonts struct {
	font *[]*truetype.Font
}

func LoadFont(path string) *Fonts {
	f := Fonts {
		font: &[]*truetype.Font{},
	}

	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(data)
	ttf, err := truetype.Parse(reader)

	*f.font = append(*f.font, ttf)

	return &f
}
