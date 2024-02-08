/*
Package goat formats "ASCII-art" drawings into Github-flavored Markdown.

 <goat>
 porcelain API
                            BuildAndWriteSVG()
                               .----------.
     ASCII-art                |            |                      Markdown
      ----------------------->|            +------------------------->
                              |            |
                               '----------'
   · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·
 plumbing API

                                Canvas{}
               NewCanvas() .-------------------.  WriteSVGBody()
                           |                   |    .-------.
     ASCII-art    .--.     | data map[x,y]rune |   |  SVG{}  |    Markdown
      ---------->|    +--->| text map[x,y]rune +-->|         +------->
                  '--'     |                   |   |         |
                           '-------------------'    '-------'
 </goat>
*/
package goat

import (
	"bytes"
	"io"
	"log"
)

// BuildAndWriteSVG reads in a newline-delimited ASCII diagram from src and writes a
// corresponding SVG diagram to dst.
func BuildAndWriteSVG(src io.Reader, dst io.Writer,
	svgColorLightScheme, svgColorDarkScheme string) {

	canvas := NewCanvas(src)
	var buff bytes.Buffer
	canvas.WriteSVGBody(&buff)

	svg := SVG{
		Body:	buff.String(),
		Width:	canvas.widthScreen(),
		Height: canvas.heightScreen(),
	}

	writeBytes(dst, svg.String(svgColorLightScheme, svgColorDarkScheme))
}


// WriteSVGBody writes the entire content of a Canvas out to a stream in SVG format.
func (c *Canvas) WriteSVGBody(dst io.Writer) {
	// We desire that pixel coordinate {0,0} should lie at center of the 8x16
	// "cell" at top-left corner of the enclosing SVG element, and that a
	// visually-pleasing margin separate that cell from the visible top-left
	// corner; the 'translate(8,16)' below accomplishes that.
	writeBytes(dst, "<g transform='translate(8,16)'>\n")
	{
		for _, l := range c.Lines() {
			l.draw(dst)
		}

		for _, tI := range c.Triangles() {
			tI.draw(dst)
		}

		for _, c := range c.RoundedCorners() {
			c.draw(dst)
		}

		for _, c := range c.Circles() {
			c.draw(dst)
		}

		for _, bI := range c.Bridges() {
			bI.draw(dst)
		}

		// Text drawing executes state-correctness checks based on recent history
		tD := &textDrawer{
			canvas: c,
			stack: []anchorIndex{},
		}
		for _, textObj := range c.Text() {
			tD.draw(dst, textObj)
		}
		if unpoppedAnchorIndexes := len(tD.stack); unpoppedAnchorIndexes > 0 {
			lastUnpoppedIndex := tD.stack[unpoppedAnchorIndexes-1]
			beginKeyRune := findAnchorKey(lastUnpoppedIndex, c.anchorStarts)
			log.Panicf(
				"End of input reached, but %d unclosed anchor-begin keys remain" +
					", last is '%c'",
				unpoppedAnchorIndexes, beginKeyRune)
		}
	}
	writeBytes(dst, "</g>\n")
}
