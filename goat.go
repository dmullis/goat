/*
Package goat formats "ASCII-art" drawings into Github-flavored Markdown.

 <goat>
 API
                            BuildAndWriteSVG()
                               .----------.
     ASCII-art                |            |                      Markdown
      ----------------------->|            +------------------------->
                              |            |
                               '----------'
   · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·
 internal

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
	"flag"
	"io"
	"log"
)

type SVGConfig struct {
	FontNames,
	FontSize,
	SvgColorLightScheme,
	SvgColorDarkScheme string
}

var DefaultSVGConfig = SVGConfig{
	FontNames: "monospace",
	FontSize: "1.1em",
	SvgColorLightScheme: "#000000",
	SvgColorDarkScheme: "#FFFFFF",
}

func InitFromFlags(svgConfig *SVGConfig) {
	defConfig := DefaultSVGConfig
	flag.StringVar(&svgConfig.FontNames, "fontnames", defConfig.FontNames,
		"Comma-separated list of fonts preferred for rasterization of the SVG")
	flag.StringVar(&svgConfig.FontSize, "fontsize", defConfig.FontSize,
		"attribute 'font-size' requested by the output SVG for text")

	flag.StringVar(&svgConfig.SvgColorLightScheme, "sls",
		defConfig.SvgColorLightScheme, `short for -svg-color-light-scheme`)
	flag.StringVar(&svgConfig.SvgColorLightScheme, "svg-color-light-scheme",
		defConfig.SvgColorLightScheme,
		`See help for -svg-color-dark-scheme`)
	flag.StringVar(&svgConfig.SvgColorDarkScheme, "sds",
		defConfig.SvgColorDarkScheme, `short for -svg-color-dark-scheme`)
	flag.StringVar(&svgConfig.SvgColorDarkScheme, "svg-color-dark-scheme",
		defConfig.SvgColorDarkScheme,
		`Goat's SVG output attempts to learn something about the background being
 drawn on top of by means of a CSS @media query, which returns a string.
 If the string is "dark", Goat draws with the color specified by
 this option; otherwise, Goat draws with the color specified by option
 -svg-color-light-scheme.

 See https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme
`)
}

// BuildAndWriteSVG reads in a newline-delimited ASCII diagram from src and writes a
// corresponding SVG diagram to dst.
func BuildAndWriteSVG(src io.Reader, dst io.Writer, svgConfig SVGConfig) {

	canvas := NewCanvas(src)
	var buff bytes.Buffer
	canvas.WriteSVGBody(&buff)

	svg := SVG{
		Width:	canvas.widthScreen(),
		Height: canvas.heightScreen(),
		SVGConfig: svgConfig,
		anchorSet: canvas.anchorSet,

		Body:	buff.String(),  // body
	}

	writeBytes(dst, svg.String())
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
			stack: []anchorSelector{},
		}
		for _, textObj := range c.Text() {
			tD.draw(dst, textObj)
		}
		if unpoppedAnchorIndexes := len(tD.stack); unpoppedAnchorIndexes > 0 {
			lastUnpoppedIndex := tD.stack[unpoppedAnchorIndexes-1]
			openKeyRune := findAnchorKey(lastUnpoppedIndex, c.anchorSet.Opens)
			log.Panicf(
				"End of input reached, but %d unclosed anchor-open keys remain unclosed" +
					", last is '%c'",
				unpoppedAnchorIndexes, openKeyRune)
		}
	}
	writeBytes(dst, "</g>\n")
}
