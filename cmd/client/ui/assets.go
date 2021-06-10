package ui

import (
	"embed"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed graphics/*
var assets embed.FS

func imageByFilename(filename string) (*ebiten.Image, error) {
	file, err := assets.Open(filename)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	ebitenImg := ebiten.NewImageFromImage(img)
	return ebitenImg, nil
}
