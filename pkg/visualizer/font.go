package visualizer

import (
	"log"
	"os"
	"os/exec"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var fontFace font.Face

func init() {
	// TODO: support windows maybe?
	filename, err := findFontFile("NotoSans:style=Regular")
	if err != nil {
		panic(err)
	}
	fontByte, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	font, err := truetype.Parse(fontByte)
	if err != nil {
		log.Fatal(err)
	}

	fontFace = truetype.NewFace(font, &truetype.Options{Size: 15})
}

func findFontFile(family string) (string, error) {
	filename, err := exec.Command("fc-match", family, "-f", "%{file}").Output()
	return string(filename), err
}
