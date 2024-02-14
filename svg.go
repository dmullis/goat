// All output is buffered into the object SVG, then written to the output stream.
package goat

import (
	"fmt"
	"io"
	"log"
)

type SVG struct {
	SVGConfig

	Width,
	Height int
	anchorSet
}

func className(aS anchorSelector) string {
	return fmt.Sprintf("_%02d", aS)
}

// See: https://drafts.csswg.org/mediaqueries-5/#prefers-color-scheme
func (s SVG) Headers() string {
	config := s.SVGConfig
	svgElem := fmt.Sprintf(
		"<svg xmlns='%s' version='%s' height='%d' width='%d'" +
			" fill='%s' font-family='%s' font-size='%s' text-anchor='middle' >\n",
		"http://www.w3.org/2000/svg",
		"1.1",   /*version*/
		s.Height, s.Width,
		"currentColor",
		config.FontNames,
		config.FontSize,
	)

	// Make order of CSS class definitions in SVG text
	// match that of source TXT -- not possible with map iteration:
	//    1. regression testing requires deterministic output
	//    2. easier debugging
	var anchorClasses, anchorDarkClasses string
	for _, anchorSelector := range s.anchorSet.Selectors {
		name := className(anchorSelector)
		anchorClasses += fmt.Sprintf(".%s {%s}\n",
			name,
			s.anchorSet.payload[anchorSelector].Class)
		anchorDarkClasses += fmt.Sprintf("    .%s {%s}\n",
			name,
			s.anchorSet.payload[anchorSelector].DarkClass)
	}

	// XX  Adding 'color-scheme: dark' fixes display of file://.../examples/*.svg in local
	//     Firefox -- not needed on Github
	style := fmt.Sprintf(
		`<style type="text/css">
svg {
    color: %s;
    stroke: currentColor;
}
text {
    stroke: none;
}
path {
    fill: none;
}
%s@media (prefers-color-scheme: dark) {
    svg {
      color-scheme: dark;
      color: %s;
    }
%s}
</style>
`,
		config.SvgColorLightScheme,
		anchorClasses,
		config.SvgColorDarkScheme,
		anchorDarkClasses)

	return svgElem + style
}

func writeBytes(out io.Writer, format string, args ...interface{}) {
	bytesOut := fmt.Sprintf(format, args...)

	_, err := out.Write([]byte(bytesOut))
	if err != nil {
		panic(err)
	}
}

// Draw a straight line as an SVG path.
func (l Line) draw(out io.Writer) {
	start := l.start.asPixel()
	stop := l.stop.asPixel()

	// For cases when a vertical line hits a perpendicular like this:
	//
	//   |		|
	//   |	  or	v
	//  ---	       ---
	//
	// We need to nudge the vertical line half a vertical cell in the
	// appropriate direction in order to meet up cleanly with the midline of
	// the cell next to it.

	// A diagonal segment all by itself needs to be shifted slightly to line
	// up with _ baselines:
	//     _
	//	\_
	//
	// TODO make this a method on Line to return accurate pixel
	if l.lonely {
		switch l.orientation {
		case NE:
			start.X -= 4
			stop.X -= 4
			start.Y += 8
			stop.Y += 8
		case SE:
			start.X -= 4
			stop.X -= 4
			start.Y -= 8
			stop.Y -= 8
		case S:
			start.Y -= 8
			stop.Y -= 8
		}

		// Half steps
		switch l.chop {
		case NONE:
		case N:
			stop.Y -= 8
		case S:
			start.Y += 8
		default:
			panic("impossible 'chop' orientation")
		}
	}

	if l.needsNudgingDown {
		stop.Y += 8
		if l.horizontal() {
			start.Y += 8
		}
	}

	if l.needsNudgingLeft {
		start.X -= 8
	}

	if l.needsNudgingRight {
		stop.X += 8
	}

	if l.needsTinyNudgingLeft {
		start.X -= 4
		if l.orientation == NE {
			start.Y += 8
		} else if l.orientation == SE {
			start.Y -= 8
		}
	}

	if l.needsTinyNudgingRight {
		stop.X += 4
		if l.orientation == NE {
			stop.Y -= 8
		} else if l.orientation == SE {
			stop.Y += 8
		}
	}

	// If either end is a hollow circle, back off drawing to the edge of the circle,
	// rather extending as usual to center of the cell.
	const (
		ORTHO = 6
		DIAG_X = 3  // XX  By eye, '3' is a bit too much'; '2' is not enough.
		DIAG_Y = 5
	)
	if (l.startRune == 'o') {
		switch l.orientation {
		case NE:
			start.X += DIAG_X
			start.Y -= DIAG_Y
		case E:
			start.X += ORTHO
		case SE:
			start.X += DIAG_X
			start.Y += DIAG_Y
		case S:
			start.Y += ORTHO
		default:
			panic("impossible orientation")
		}
	}
	// X  'stopRune' case differs from 'startRune' only by inversion of the arithmetic signs.
	if (l.stopRune == 'o') {
		switch l.orientation {
		case NE:
			stop.X -= DIAG_X
			stop.Y += DIAG_Y
		case E:
			stop.X -= ORTHO
		case SE:
			stop.X -= DIAG_X
			stop.Y -= DIAG_Y
		case S:
			stop.Y -= ORTHO
		default:
			panic("impossible orientation")
		}
	}

	writeBytes(out,
		"<path d='M %d,%d L %d,%d'></path>\n",
		start.X, start.Y,
		stop.X, stop.Y,
	)
}

// Draw a solid triangle as an SVG polygon element.
func (t Triangle) draw(out io.Writer) {
	// https://www.w3.org/TR/SVG/shapes.html#PolygonElement

	/*
		  +-----+-----+
		  |    /|\    |
		  |   / | \   |
		x +- / -+- \ -+
		  | /	|   \ |
		  |/	|    \|
		  +-----+-----+
			y
	*/

	x, y := float32(t.start.asPixel().X), float32(t.start.asPixel().Y)
	r := 0.0

	x0 := x + 8
	y0 := y
	x1 := x - 4
	y1 := y - 0.35*16
	x2 := x - 4
	y2 := y + 0.35*16

	switch t.orientation {
	case N:
		r = 270
		if t.needsNudging {
			x0 += 8
			x1 += 8
			x2 += 8
		}
	case NE:
		r = 300
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case NW:
		r = 240
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case W:
		r = 180
		if t.needsNudging {
			x0 -= 8
			x1 -= 8
			x2 -= 8
		}
	case E:
		r = 0
		if t.needsNudging {
			x0 -= 8
			x1 -= 8
			x2 -= 8
		}
	case S:
		r = 90
		if t.needsNudging {
			x0 += 8
			x1 += 8
			x2 += 8
		}
	case SW:
		r = 120
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	case SE:
		r = 60
		x0 += 4
		x1 += 4
		x2 += 4
		if t.needsNudging {
			x0 += 6
			x1 += 6
			x2 += 6
		}
	}

	writeBytes(out,
		"<polygon points='%f,%f %f,%f %f,%f' fill='currentColor' transform='rotate(%f, %f, %f)'></polygon>\n",
		x0, y0,
		x1, y1,
		x2, y2,
		r,
		x, y,
	)
}

// Draw a solid circle as an SVG circle element.
func (c *Circle) draw(out io.Writer) {
	var fill string
	if c.bold {
		fill = "currentColor"
	} else {
		fill = "none"
	}
	pixel := c.start.asPixel()
	const circleRadius = 6
	writeBytes(out,
		"<circle cx='%d' cy='%d' r='%d' fill='%s'></circle>\n",
		pixel.X,
		pixel.Y,
		circleRadius,
		fill,
	)
}

// Draw a rounded corner as an SVG elliptical arc element.
func (c *RoundedCorner) draw(out io.Writer) {
	// https://www.w3.org/TR/SVG/paths.html#PathDataEllipticalArcCommands

	x, y := c.start.asPixelXY()
	startX, startY, endX, endY, sweepFlag := 0, 0, 0, 0, 0

	switch c.orientation {
	case NW:
		startX = x + 8
		startY = y
		endX = x - 8
		endY = y + 16
	case NE:
		sweepFlag = 1
		startX = x - 8
		startY = y
		endX = x + 8
		endY = y + 16
	case SE:
		sweepFlag = 1
		startX = x + 8
		startY = y - 16
		endX = x - 8
		endY = y
	case SW:
		startX = x - 8
		startY = y - 16
		endX = x + 8
		endY = y
	}

	writeBytes(out,
		"<path d='M %d,%d A 16,16 0 0,%d %d,%d'></path>\n",
		startX,
		startY,
		sweepFlag,
		endX,
		endY,
	)
}

// Draw a bridge as an SVG elliptical arc element.
func (b Bridge) draw(out io.Writer) {
	x, y := b.start.asPixelXY()
	sweepFlag := 1

	if b.orientation == W {
		sweepFlag = 0
	}

	writeBytes(out,
		"<path d='M %d,%d A 9,9 0 0,%d %d,%d'></path>\n",
		x, y-8,
		sweepFlag,
		x, y+8,
	)
}

type textDrawer struct {
	canvas *Canvas
	stack []anchorSelector
}

// Draw a single text character as an SVG text element.
func (tD *textDrawer) draw(out io.Writer, t Text) {
	canvas := tD.canvas
	// Detect requested anchor start/end points, emit appropriate element
	aS := canvas.anchorSet
	c_rune := t.rune
	str := string(c_rune)

	if anchorSelector, found := aS.Closes[c_rune]; found {
		if len(tD.stack) == 0 {
			log.Panicf("close key '%c' found at %#v, but no matching open key",
				c_rune, t.start)
		}

		payload := aS.payload[anchorSelector]
		str = string(payload.Replacements[1])
		finalDraw(out, t.start.asPixel(), str)

		if expectedAI := tD.stack[len(tD.stack)-1]; expectedAI != anchorSelector {
			openRune := findAnchorKey(expectedAI, aS.Opens)
			log.Panicf("earlier open key '%c' does not match closing anchor key '%c' at %#v",
				openRune, c_rune, t.start)
		}
		tD.stack = tD.stack[:len(tD.stack)-1]
		writeBytes(out, "</a>\n")
		return
	}

	if anchorSelector, found := aS.Opens[c_rune]; found {
		payload := aS.payload[anchorSelector]
		tD.stack = append(tD.stack, anchorSelector)
		writeBytes(out, "<a class='%s' %s>\n",
			className(anchorSelector),
			payload.Attributes)

		str = string(payload.Replacements[0])
	}
	finalDraw(out, t.start.asPixel(), str)
}

func finalDraw(out io.Writer, p pixel, str string) {
	// Markdeep special-cases these character and treats them like a
	// checkerboard.
	{
		opacity := 0
		switch str {
		case "▉":
			opacity = -64
		case "▓":
			opacity = 64
		case "▒":
			opacity = 128
		case "░":
			opacity = 191
		}

		fill := "currentColor"

		if opacity != 0 {
			if opacity > 0 {
				fill = fmt.Sprintf("rgb(%d,%d,%d)", opacity, opacity, opacity)
			}
			writeBytes(out,
				"<rect x='%d' y='%d' width='8' height='16' fill='%s'></rect>",
				p.X-4, p.Y-8,
				fill,
			)
			return
		}
	}

	// Escape to allow embedding of output SVG within XML
	switch str {
	case "&":
		str = "&amp;"
	case ">":
		str = "&gt;"
	case "<":
		str = "&lt;"
	}

	// usual case
	writeBytes(out,
		`<text x='%d' y='%d'>%s</text>
`,
		p.X,
		p.Y+4,  // '4' here achieves desired alignment with neighboring graphics
		str)
}
