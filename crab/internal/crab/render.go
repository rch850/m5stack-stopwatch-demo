// Package crab draws the same procedural illustration on the device and host.
// One RGB565 scanline is sufficient; neither PSRAM nor image assets are required.
package crab

import "math"

const Size = 466

type Color uint16

func RGB(r, g, b uint8) Color { return Color(uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)) }

var (
	background = RGB(7, 22, 32)
	sea        = RGB(11, 38, 48)
	red        = RGB(241, 87, 67)
	darkRed    = RGB(178, 49, 47)
	lightRed   = RGB(255, 147, 105)
	cream      = RGB(255, 238, 200)
	ink        = RGB(24, 29, 38)
)

type row struct {
	y, x   int
	pixels []byte
}

func (r row) span(x0, x1 int, c Color) {
	if x0 < r.x {
		x0 = r.x
	}
	if x1 >= r.x+len(r.pixels)/2 {
		x1 = r.x + len(r.pixels)/2 - 1
	}
	for x := x0; x <= x1; x++ {
		i := (x - r.x) * 2
		r.pixels[i], r.pixels[i+1] = byte(c>>8), byte(c)
	}
}

func (r row) ellipse(cx, cy, rx, ry int, c Color) {
	dy := r.y - cy
	if dy < -ry || dy > ry {
		return
	}
	dx := int(float64(rx) * math.Sqrt(1-float64(dy*dy)/float64(ry*ry)))
	r.span(cx-dx, cx+dx, c)
}

type point struct{ x, y int }

func (r row) polygon(p []point, c Color) {
	var xs [12]int
	n := 0
	for i, a := range p {
		b := p[(i+1)%len(p)]
		if (a.y <= r.y && b.y > r.y) || (b.y <= r.y && a.y > r.y) {
			xs[n] = a.x + (r.y-a.y)*(b.x-a.x)/(b.y-a.y)
			n++
		}
	}
	for i := 1; i < n; i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
	for i := 0; i+1 < n; i += 2 {
		r.span(xs[i], xs[i+1], c)
	}
}

func (r row) limb(ax, ay, bx, by, width int, c Color) {
	dx, dy := float64(bx-ax), float64(by-ay)
	length := math.Sqrt(dx*dx + dy*dy)
	ox, oy := int(-dy*float64(width)/length), int(dx*float64(width)/length)
	r.polygon([]point{{ax + ox, ay + oy}, {bx + ox, by + oy}, {bx - ox, by - oy}, {ax - ox, ay - oy}}, c)
	r.ellipse(ax, ay, width, width, c)
	r.ellipse(bx, by, width, width, c)
}

// RenderRow writes big-endian RGB565 pixels for [x, x+len(dst)/2).
// Claw regions are x=58..177 and x=288..407, y=102..221 (even aligned).
func RenderRow(dst []byte, x, y, left, right int) {
	r := row{y, x, dst}
	r.span(x, x+len(dst)/2-1, background)
	r.ellipse(233, 233, 211, 211, sea)
	// Bubbles, sand and a small coral branch.
	for _, b := range [][3]int{{93, 244, 6}, {365, 251, 9}, {150, 91, 5}, {325, 82, 4}, {344, 352, 5}} {
		r.ellipse(b[0], b[1], b[2], b[2], RGB(39, 83, 91))
		r.ellipse(b[0], b[1], b[2]-2, b[2]-2, sea)
	}
	r.ellipse(233, 381, 143, 30, RGB(18, 51, 59))
	r.ellipse(233, 350, 109, 13, background)
	for s := -1; s <= 1; s += 2 {
		for i := 0; i < 3; i++ {
			r.limb(233+s*63, 278+i*14, 233+s*(106+i*4), 285+i*21, 7, darkRed)
			r.limb(233+s*(106+i*4), 285+i*21, 233+s*(120-i*3), 309+i*18, 5, red)
		}
		r.limb(233+s*67, 269, 233+s*106, 230, 11, darkRed)
		r.limb(233+s*106, 230, 233+s*115, 201, 10, red)
	}
	// Claws close toward the middle of each palm.
	for side, closure := range [2]int{left, right} {
		cx := 118
		if side == 1 {
			cx = 348
		}
		gap := 28 - closure*6
		r.ellipse(cx, 185, 29, 28, darkRed)
		r.ellipse(cx, 179, 27, 27, red)
		r.polygon([]point{{cx - 25, 174}, {cx - 35, 147}, {cx - gap, 113}, {cx - gap + 8, 143}, {cx - 2, 162}, {cx - 2, 183}}, red)
		r.polygon([]point{{cx + 25, 174}, {cx + 35, 147}, {cx + gap, 113}, {cx + gap - 8, 143}, {cx + 2, 162}, {cx + 2, 183}}, red)
		r.limb(cx-18, 177, cx-23, 156, 3, lightRed)
		r.ellipse(cx-9, 182, 5, 7, lightRed)
	}
	// Shell, eye stalks and face.
	r.ellipse(233, 285, 84, 56, darkRed)
	r.ellipse(233, 275, 84, 51, red)
	r.ellipse(224, 254, 55, 21, RGB(250, 110, 79))
	for _, cx := range []int{205, 261} {
		r.limb(cx, 238, cx, 218, 7, darkRed)
		r.ellipse(cx, 211, 18, 21, cream)
		r.ellipse(cx+3, 211, 8, 11, ink)
		r.ellipse(cx+5, 207, 3, 4, RGB(255, 255, 255))
	}
	r.ellipse(192, 278, 11, 6, lightRed)
	r.ellipse(274, 278, 11, 6, lightRed)
	r.ellipse(233, 282, 16, 14, ink)
	r.ellipse(233, 276, 18, 11, red)
	// Button color hints, kept inside the circular screen.
	r.ellipse(180, 377, 7, 7, RGB(255, 208, 68))
	r.ellipse(286, 377, 7, 7, RGB(84, 175, 255))
	if left > 0 {
		r.ellipse(180, 377, 3, 3, cream)
	}
	if right > 0 {
		r.ellipse(286, 377, 3, 3, cream)
	}
}
