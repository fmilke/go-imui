package main

import (
	"image/png"
	"image"
	"os"
)

func loadPng(path string) (i image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return image.Black, err
	}

	i, err = png.Decode(file)
	if err != nil {
		return image.Black, err
	}

	return i, nil
}

func writePng(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	err = png.Encode(file, img)

	return err
}
