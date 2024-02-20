package goat

import (
	"bufio"
	"io"
)

type (
	exists struct{}
	runeSet map[rune]exists
)

// Canvas represents the current state of parsing of the UTF-8 source text.
type Canvas struct {
	// units of cells
	Width, Height int

	data   map[Index]rune

	text   map[Index]rune

	anchorSet
}

// Characters where more than one line segment can come together.
var jointRunes = []rune{
	'.',
	'\'',
	'+',
	'*',
	'o',
}
var jointRunesSet = makeSet(jointRunes)

// XX  "reserved" is not quite the right term, since some of these are accepted as part of text strings.
var reservedSet = makeSet(
	append(
		jointRunes,
		'-',
		'_',
		'|',
		'v',
		'^',
		'>',
		'<',
		'/',
		'\\',
		')',
		'(',
		' ',   // X SPACE is reserved
	))

// The SVG graphic that each of these runes codes for is too wide, or otherwise
// problematic if laid side-by-side with another of the set.
// So, if seen in the input, Goat assumes the user intends
// them to appear as part of a text string in the output SVG.
var wideSVGSet = makeSet([]rune{
	'o',   // too wide
	'*',   // too wide
	'v',   // X  Input containing " over " needs to be considered text.
	//	'>',   // Uncommenting would get 'o<' and '>o' wrong.  But o> and >o -- never desired to be text?
	//	'<',   // ibid.
	'^',
	')',
	'(',
	'.',   // Dropping this would cause " v. " to be considered graphics.
})

// NewCanvas creates a fully-populated Canvas according to GoAT-formatted text read from
// an io.Reader, consuming all bytes available.
func NewCanvas(in io.Reader) (c Canvas) {
	width := 0
	height := 0

	scanner := bufio.NewScanner(in)

	c = Canvas{
		data:	make(map[Index]rune),
		text:	nil,
		anchorSet: NewAnchorSet(),
	}

	// first step is a text-line oriented scan of full input
	for line := 0; scanner.Scan(); line++ {
		lineRunes := []rune(scanner.Text())
		// treat blank line as a special case
		if len(lineRunes) == 0 {
			height++
			continue
		}

		// If a comment line, handle here
		if lineRunes[0] == '#' {
			// For each comment line with validly-formatted anchor keys in columns 2-6
			// the remainder of the line is assumed to be HTML attribute specifications
			// that have meaning either to <text> elements, or to an anchor <a> element
			// that will enclose a contiguous run of the <text> elements.
			if isAnchorSpecifier(lineRunes) {
				c.parseAnchorSpecifier(lineRunes, anchorSelector(line+1))
			}
			continue
		}

		// common case: Insert to the 'data' map.
		w := 0
		// X  Type of second value assigned from "for ... range" operator over a string is "rune".
		//               https://go.dev/ref/spec#For_statements
		//    But yet, counterintuitively, type of anyString[_index_] is 'byte'.
		//               https://go.dev/ref/spec#String_types
		for _, r := range lineRunes {
			//if r > 255 {
			//	fmt.Printf("r == 0x%x\n", r)
			//}
			if r == '	' {
				panic("TAB character found on input")
			}
			i := Index{w, height}
			c.data[i] = r
			w++
		}
		width = max(width,w)
		height++
	}

	// XX  Push init below up to caller; rename to "NewDataIndex()
	c.Width = width
	c.Height = height

	c.text = make(map[Index]rune)
	// Cell-wise scan of c.data to fill the 'c.text' map, with runes removed from c.data.
	//  Why not done in the course of the line-oriented scan, in the loop above?
	//    => because "looking down" will be required, as done in canvas.MoveToText().
	c.MoveToText()
	return
}

// Move contents of every cell that appears, according to a tricky set of rules,
// to be "text", into a separate map: from data[] to text[].
// So data[] and text[] are an exact partitioning of the
// incoming grid-aligned runes.
func (c *Canvas) MoveToText() {
	for i := range leftRightMinor(c.Width, c.Height) {
		if c.shouldMoveToText(i) {
			c.text[i] = c.runeAt(i)	// c.runeAt() Reads from c.data[]
		}
	}
	for i := range c.text {
		delete(c.data, i)
	}
}

func (c *Canvas) shouldMoveToText(i Index) bool {
	i_r := c.runeAt(i)
	if i_r == ' ' {
		// X  Note that c.runeAt(i) returns ' ' if i lies right of all chars on line i.Y
		return false
	}

	// Returns true if the character at index 'i' of c.data[] is reserved for diagrams.
	// Characters like 'o' and 'v' need more context (e.g., are other text characters
	// nearby) to determine whether they're part of a diagram.
	isReserved := func(i Index) (found bool) {
		i_r, inData := c.data[i]
		if !inData {
			// lies off left or right end of line, treat as reserved
			return true
		}
		_, found = reservedSet[i_r]
		return
	}

	if !isReserved(i) {
		return true
	}

	// This is a reserved character with an incoming line (e.g., "|") above or below it,
	// so call it non-text.
	if c.hasLineAboveOrBelow(i) {
		return false
	}

	w := i.west()
	e := i.east()

	// Reserved characters like "o" or "*" with letters sitting next to them
	// are probably text.
	// TODO: Fix this to count contiguous blocks of text. If we had a bunch of
	// reserved characters previously that were counted as text then this
	// should be as well, e.g., "A----B".

	// 'i' is reserved but surrounded by text and probably part of an existing word.
	// Preserve chains of reserved-but-text characters like "foo----bar".
	if textLeft := !isReserved(w); textLeft {
		return true
	}
	if textRight := !isReserved(e); textRight {
		return true
	}

	wide := func (x Index) bool {
		return inSet(wideSVGSet, c.data, x)
	}
	if wide(i) {
		if wide(e) || wide(w) {
			return true
		}
	}

	// If 'i' has anything other than a space to either left or right, treat as non-text.
	if !(c.runeAt(w) == ' ' && c.runeAt(e) == ' ') {
		return false
	}

	// Circles surrounded by whitespace shouldn't be shown as text.
	if i_r == 'o' || i_r == '*' {
		return false
	}

	// 'i' is surrounded by whitespace or text on one side or the other, at two cell's distance.
	if !isReserved(w.west()) || !isReserved(e.east()) {
		return true
	}

	return false
}

// Text returns a slice of all "text" characters i.e. those not belonging
// to part of the diagram, ordered top-to-bottom, left-to-right.
func (c *Canvas) Text() (text []Text) {
	for idx := range leftRightMinor(c.Width, c.Height) {
		r, found := c.text[idx]
		if !found {
			continue
		}
		text = append(text, Text{
			start: idx,
			rune: r})
	}
	return
}


func makeSet(runeSlice []rune) (rs runeSet) {
	rs = make(runeSet)
	for _, r := range runeSlice {
		rs[r] = exists{}
	}
	return
}

func isJoint(r rune) (in bool) {
	_, in = jointRunesSet[r]
	return
}

// XX  rename 'isCircle()'?
func isDot(r rune) bool {
	return r == 'o' || r == '*'
}

func isTriangle(r rune) bool {
	return r == '^' || r == 'v' || r == '<' || r == '>'
}

func (c *Canvas) heightScreen() int {
	return c.Height*16 + 8 + 1
}

func (c *Canvas) widthScreen() int {
	return (c.Width + 1) * 8
}

// Arg 'canvasMap' is typically either Canvas.data or Canvas.text
func inSet(set runeSet, canvasMap map[Index]rune, i Index) (inset bool) {
	r, inMap := canvasMap[i]
	if !inMap {
		return false 	// r == rune(0)
	}
	_, inset = set[r]
	return
}

// Looks only at c.data[], ignores c.text[].
// Returns the rune for ASCII Space i.e. ' ', in the event that map lookup fails.
//  XX  Name 'dataRuneAt()' would be more descriptive, but maybe too bulky.
func (c *Canvas) runeAt(i Index) rune {
	if val, ok := c.data[i]; ok {
		return val
	}
	return ' '
}
