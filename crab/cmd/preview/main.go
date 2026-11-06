// Command preview exports the actual firmware drawing as PNG and animated GIF.
package main

import (
	"example.com/stopwatch-crab/internal/crab"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"os"
)

func frame(left, right, mouth int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, crab.Size, crab.Size))
	line := make([]byte, crab.Size*2)
	for y := 0; y < crab.Size; y++ {
		crab.RenderRow(line, 0, y, left, right, mouth)
		for x := 0; x < crab.Size; x++ {
			v := uint16(line[x*2])<<8 | uint16(line[x*2+1])
			r, g, b := uint8(v>>11), uint8((v>>5)&63), uint8(v&31)
			im.SetRGBA(x, y, color.RGBA{r<<3 | r>>2, g<<2 | g>>4, b<<3 | b>>2, 255})
		}
	}
	return im
}

func main() {
	if err := os.MkdirAll("docs", 0755); err != nil {
		log.Fatal(err)
	}
	im := frame(0, 0, 0)
	f, err := os.Create("docs/crab.png")
	if err != nil {
		log.Fatal(err)
	}
	if err = png.Encode(f, im); err != nil {
		log.Fatal(err)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
	// Palette is derived exactly from the RGB565 renderer, preserving its colors.
	palette := color.Palette{}
	seen := map[color.RGBA]bool{}
	for y := 0; y < crab.Size; y++ {
		for x := 0; x < crab.Size; x++ {
			c := im.RGBAAt(x, y)
			if !seen[c] {
				seen[c] = true
				palette = append(palette, c)
			}
		}
	}
	anim := gif.GIF{LoopCount: 0}
	poses := [][3]int{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {3, 0}, {2, 0}, {1, 0}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 3}, {0, 2}, {0, 1}, {0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {3, 3}, {2, 2}, {1, 1}}
	for _, mouth := range []int{0, 1, 2, 4, 6, 8, 6, 4, 2, 1, 0} {
		poses = append(poses, [3]int{0, 0, mouth})
	}
	for _, p := range poses {
		out := image.NewPaletted(im.Bounds(), palette)
		draw.Draw(out, out.Bounds(), frame(p[0], p[1], p[2]), image.Point{}, draw.Src)
		anim.Image = append(anim.Image, out)
		delay := 8
		if p == [3]int{0, 0, 0} || p[0] == 4 || p[1] == 4 {
			delay = 40
		}
		anim.Delay = append(anim.Delay, delay)
	}
	f, err = os.Create("docs/crab.gif")
	if err != nil {
		log.Fatal(err)
	}
	if err = gif.EncodeAll(f, &anim); err != nil {
		log.Fatal(err)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}
