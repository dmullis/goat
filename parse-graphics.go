package goat

import (
	"io"
)


// Drawable represents anything that can Draw itself.
type Drawable interface {
	draw(out io.Writer)
}

// Line represents a straight segment between two points 'start' and 'stop', where
// 'start' is either lesser in X (north-east, east, south-east), or
// equal in X and lesser in Y (south).
type Line struct {
	start Index
	stop  Index

	startRune rune
	stopRune rune

	// dashed	    bool
	needsNudgingDown      bool
	needsNudgingLeft      bool
	needsNudgingRight     bool
	needsTinyNudgingLeft  bool
	needsTinyNudgingRight bool

	// This is a line segment all by itself. This centers the segment around
	// the midline.
	lonely bool
	// N or S. Only useful for half steps - chops of this half of the line.
	chop Orientation

	// X-major, Y-minor.  Therefore, always one of the compass points NE, E, SE, S.
	orientation Orientation

	state lineState
}

type lineState int

const (
	_Unstarted lineState = iota
	_Started
)

func (l *Line) started() bool {
	return l.state == _Started
}

func (c *Canvas) setStart(l *Line, i Index) {
	if l.state == _Unstarted {
		l.start = i
		l.startRune = c.runeAt(i)
		l.stop = i
		l.stopRune = c.runeAt(i)
		l.state = _Started
	}
}

func (c *Canvas) setStop(l *Line, i Index) {
	if l.state == _Started {
		l.stop = i
		l.stopRune = c.runeAt(i)
	}
}

func (l *Line) goesSomewhere() bool {
	return l.start != l.stop
}

func (l *Line) horizontal() bool {
	return l.orientation == E || l.orientation == W
}

func (l *Line) vertical() bool {
	return l.orientation == N || l.orientation == S
}

func (l *Line) diagonal() bool {
	return l.orientation == NE || l.orientation == SE || l.orientation == SW || l.orientation == NW
}

// Triangle corresponds to "^", "v", "<" and ">" runes in the absence of
// surrounding alphanumerics.
type Triangle struct {
	start	     Index
	orientation  Orientation
	needsNudging bool
}

// Circle corresponds to "o" or "*" runes in the absence of surrounding
// alphanumerics.
type Circle struct {
	start Index
	bold  bool
}

// RoundedCorner corresponds to combinations of "-." or "-'".
type RoundedCorner struct {
	start	    Index
	orientation Orientation
}

// Text corresponds to any runes not reserved for diagrams, or reserved runes
// surrounded by alphanumerics.
type Text struct {
	// One of 'attributeKeySet' to start an anchor '<a>',
	// 'closeAnchorKey' to emit '</a>'; 0x0 otherwise.
	//anchorKey rune

	start	 Index

	// Possibly multiple bytes, from Unicode source of type 'rune'.
	// XX  Why changed from the original (locally-allocated) rune type?
	str	 string
}

// Bridge corresponds to combinations of "-)-" or "-(-" and is displayed as
// the vertical line "hopping over" the horizontal.
type Bridge struct {
	start	    Index
	orientation Orientation
}

// Orientation represents the primary direction that a Drawable is facing.
type Orientation int

const (
	NONE Orientation = iota // No orientation; no structure present.
	N			// North
	NE			// Northeast
	NW			// Northwest
	S			// South
	SE			// Southeast
	SW			// Southwest
	E			// East
	W			// West
)

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

		for _, textObj := range c.Text() {
			textObj.draw(dst, c)
		}
	}
	writeBytes(dst, "</g>\n")
}

// Lines returns a slice of all Line drawables that we can detect -- in all
// possible orientations.
func (c *Canvas) Lines() (lines []Line) {
	horizontalMidlines := c.getLinesForSegment('-')

	diagUpLines := c.getLinesForSegment('/')
	for i, l := range diagUpLines {
		// /_
		if c.runeAt(l.start.east()) == '_' {
			diagUpLines[i].needsTinyNudgingLeft = true
		}

		// _
		// /
		if c.runeAt(l.stop.north()) == '_' {
			diagUpLines[i].needsTinyNudgingRight = true
		}

		//  _
		// /
		if !l.lonely && c.runeAt(l.stop.nEast()) == '_' {
			diagUpLines[i].needsTinyNudgingRight = true
		}

		// _/
		if !l.lonely && c.runeAt(l.start.west()) == '_' {
			diagUpLines[i].needsTinyNudgingLeft = true
		}

		// \
		// /
		if !l.lonely && c.runeAt(l.stop.north()) == '\\' {
			diagUpLines[i].needsTinyNudgingRight = true
		}

		// /
		// \
		if !l.lonely && c.runeAt(l.start.south()) == '\\' {
			diagUpLines[i].needsTinyNudgingLeft = true
		}
	}

	diagDownLines := c.getLinesForSegment('\\')
	for i, l := range diagDownLines {
		// _\
		if c.runeAt(l.stop.west()) == '_' {
			diagDownLines[i].needsTinyNudgingRight = true
		}

		// _
		// \
		if c.runeAt(l.start.north()) == '_' {
			diagDownLines[i].needsTinyNudgingLeft = true
		}

		//  _
		//   \
		if !l.lonely && c.runeAt(l.start.nWest()) == '_' {
			diagDownLines[i].needsTinyNudgingLeft = true
		}

		// \_
		if !l.lonely && c.runeAt(l.stop.east()) == '_' {
			diagDownLines[i].needsTinyNudgingRight = true
		}

		// \
		// /
		if !l.lonely && c.runeAt(l.stop.south()) == '/' {
			diagDownLines[i].needsTinyNudgingRight = true
		}

		// /
		// \
		if !l.lonely && c.runeAt(l.start.north()) == '/' {
			diagDownLines[i].needsTinyNudgingLeft = true
		}
	}

	horizontalBaselines := c.getLinesForSegment('_')
	for i, l := range horizontalBaselines {
		// TODO: make this nudge an orientation
		horizontalBaselines[i].needsNudgingDown = true

		//     _
		// _| |
		if c.runeAt(l.stop.sEast()) == '|' || c.runeAt(l.stop.nEast()) == '|' {
			horizontalBaselines[i].needsNudgingRight = true
		}

		// _
		//  |  _|
		if c.runeAt(l.start.sWest()) == '|' || c.runeAt(l.start.nWest()) == '|' {
			horizontalBaselines[i].needsNudgingLeft = true
		}

		//     _
		// _/	\
		if c.runeAt(l.stop.east()) == '/' || c.runeAt(l.stop.sEast()) == '\\' {
			horizontalBaselines[i].needsTinyNudgingRight = true
		}

		//	 _
		// \_	/
		if c.runeAt(l.start.west()) == '\\' || c.runeAt(l.start.sWest()) == '/' {
			horizontalBaselines[i].needsTinyNudgingLeft = true
		}

		// _\
		if c.runeAt(l.stop.east()) == '\\' {
			horizontalBaselines[i].needsNudgingRight = true
			horizontalBaselines[i].needsTinyNudgingRight = true
		}

		//
		// /_
		if c.runeAt(l.start.west()) == '/' {
			horizontalBaselines[i].needsNudgingLeft = true
			horizontalBaselines[i].needsTinyNudgingLeft = true
		}
		//  _
		//  /
		if c.runeAt(l.stop.south()) == '/' {
			horizontalBaselines[i].needsTinyNudgingRight = true
		}

		//  _
		//  \
		if c.runeAt(l.start.south()) == '\\' {
			horizontalBaselines[i].needsTinyNudgingLeft = true
		}

		//  _
		// '
		if c.runeAt(l.start.sWest()) == '\'' {
			horizontalBaselines[i].needsNudgingLeft = true
		}

		// _
		//  '
		if c.runeAt(l.stop.sEast()) == '\'' {
			horizontalBaselines[i].needsNudgingRight = true
		}
	}

	verticalLines := c.getLinesForSegment('|')

	lines = append(lines, horizontalMidlines...)
	lines = append(lines, horizontalBaselines...)
	lines = append(lines, verticalLines...)
	lines = append(lines, diagUpLines...)
	lines = append(lines, diagDownLines...)
	lines = append(lines, c.HalfSteps()...)  // vertical, only

	return
}

func newHalfStep(i Index, chop Orientation) Line {
	return Line{
		start:	     i,
		stop:	     i.south(),
		lonely:	     true,
		chop:	     chop,
		orientation: S,
	}
}

func (c *Canvas) HalfSteps() (lines []Line) {
	for idx := range upDownMinor(c.Width, c.Height) {
		if o := c.partOfHalfStep(idx); o != NONE {
			lines = append(
				lines,
				newHalfStep(idx, o),
			)
		}
	}
	return
}

func (c *Canvas) getLinesForSegment(segment rune) []Line {
	var iter canvasIterator
	var orientation Orientation
	var passThroughs []rune

	switch segment {
	case '-':
		iter = leftRightMinor
		orientation = E
		passThroughs = append(jointRunes, '<', '>', '(', ')')
	case '_':
		iter = leftRightMinor
		orientation = E
		passThroughs = append(jointRunes, '|')
	case '|':
		iter = upDownMinor
		orientation = S
		passThroughs = append(jointRunes, '^', 'v')
	case '/':
		iter = diagUp
		orientation = NE
		passThroughs = append(jointRunes, 'o', '*', '<', '>', '^', 'v', '|')
	case '\\':
		iter = diagDown
		orientation = SE
		passThroughs = append(jointRunes, 'o', '*', '<', '>', '^', 'v', '|')
	default:
		return nil
	}

	return c.getLines(iter, segment, passThroughs, orientation)
}

// ci: the order that we traverse locations on the canvas.
// segment: the primary character we're tracking for this line.
// passThroughs: characters the line segment is allowed to be drawn underneath
// (without terminating the line).
// orientation: the orientation for this line.
func (c *Canvas) getLines(
	ci canvasIterator,
	segment rune,
	passThroughs []rune,
	o Orientation,
) (lines []Line) {
	// Helper to throw the current line we're tracking on to the slice and
	// start a new one.
	snip := func(cl Line) Line {
		// Only collect lines that actually go somewhere or are isolated
		// segments; otherwise, discard what's been collected so far within 'cl'.
		if cl.goesSomewhere() {
			lines = append(lines, cl)
		}

		return Line{orientation: o}
	}

	currentLine := Line{orientation: o}
	lastSeenRune := ' '

	// XX  linear search of slice -- alternative to a map test
	contains := func(in []rune, r rune) bool {
		for _, v := range in {
			if r == v {
				return true
			}
		}
		return false
	}

	for idx := range ci(c.Width+1, c.Height+1) {
		r := c.runeAt(idx)

		isSegment := r == segment
		isPassThrough := contains(passThroughs, r)
		isRoundedCorner := c.isRoundedCorner(idx)
		isDot := isDot(r)
		isTriangle := isTriangle(r)

		justPassedThrough := contains(passThroughs, lastSeenRune)

		shouldKeep := (isSegment || isPassThrough) && isRoundedCorner == NONE

		// This is an edge case where we have a rounded corner... that's also a
		// joint... attached to orthogonal line, e.g.:
		//
		//  '+--
		//   |
		//
		// TODO: This also depends on the orientation of the corner and our
		// line.
		// NW / NE line can't go with EW/NS lines, vertical is OK though.
		if isRoundedCorner != NONE && o != E && (c.partOfVerticalLine(idx) || c.partOfDiagonalLine(idx)) {
			shouldKeep = true
		}

		// Don't connect | to > for diagonal lines or )) for horizontal lines.
		if isPassThrough && justPassedThrough && o != S {
			currentLine = snip(currentLine)
		}

		// Don't connect o to o, + to o, etc. This character is a new pass-through
		// so we still want to respect shouldKeep; we just don't want to draw
		// the existing line through this cell.
		if justPassedThrough && (isDot || isTriangle) {
			currentLine = snip(currentLine)
		}

		switch currentLine.state {
		case _Unstarted:
			if shouldKeep {
				c.setStart(&currentLine, idx)
			}
		case _Started:
			if !shouldKeep {
				// Snip the existing line, don't add the current cell to it
				// *unless* its a line segment all by itself. If it is, keep a
				// record that it's an individual segment because we need to
				// adjust later in the / and \ cases.
				if !currentLine.goesSomewhere() && lastSeenRune == segment {
					if !c.partOfRoundedCorner(currentLine.start) {
						c.setStop(&currentLine, idx)
						currentLine.lonely = true
					}
				}
				currentLine = snip(currentLine)
			} else if isPassThrough {
				// Snip the existing line but include the current pass-through
				// character because we may be continuing the line.
				c.setStop(&currentLine, idx)
				currentLine = snip(currentLine)
				c.setStart(&currentLine, idx)
			} else if shouldKeep {
				// Keep the line going and extend it by this character.
				c.setStop(&currentLine, idx)
			}
		}

		lastSeenRune = r
	}
	return
}

// Triangles detects intended triangles -- typically at the end of an intended line --
// and returns a representational slice composed of types Triangle and Line.
func (c *Canvas) Triangles() (triangles []Drawable) {
	o := NONE

	for idx := range upDownMinor(c.Width, c.Height) {
		needsNudging := false
		start := idx

		r := c.runeAt(idx)

		if !isTriangle(r) {
			continue
		}

		// Identify orientation and nudge the triangle to touch any
		// adjacent walls.
		switch r {
		case '^':
			o = N
			//  ^  and ^
			// /	    \
			if c.runeAt(start.sWest()) == '/' {
				o = NE
			} else if c.runeAt(start.sEast()) == '\\' {
				o = NW
			}
		case 'v':
			if c.runeAt(start.north()) == '|' {
				// |
				// v
				o = S
			} else if c.runeAt(start.nEast()) == '/' {
				//  /
				// v
				o = SW
			} else if c.runeAt(start.nWest()) == '\\' {
				//  \
				//   v
				o = SE
			} else {
				// Conclusion: Meant as a text string 'v', not a triangle
				//panic("Not sufficient to fix all 'v' troubles.")
				// continue   XX Already committed to non-text output for this string?
				o = S
			}
		case '<':
			o = W
		case '>':
			o = E
		}

		// Determine if we need to snap the triangle to something and, if so,
		// draw a tail if we need to.
		switch o {
		case N:
			r := c.runeAt(start.north())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(triangles, newHalfStep(start, N))
			}
		case NW:
			r := c.runeAt(start.nWest())
			// Need to draw a tail.
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					Line{
						start:	     start.nWest(),
						stop:	     start,
						orientation: SE,
					},
				)
			}
		case NE:
			r := c.runeAt(start.nEast())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					Line{
						start:	     start,
						stop:	     start.nEast(),
						orientation: NE,
					},
				)
			}
		case S:
			r := c.runeAt(start.south())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(triangles, newHalfStep(start, S))
			}
		case SE:
			r := c.runeAt(start.sEast())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					Line{
						start:	     start,
						stop:	     start.sEast(),
						orientation: SE,
					},
				)
			}
		case SW:
			r := c.runeAt(start.sWest())
			if r == '-' || isJoint(r) && !isDot(r) {
				needsNudging = true
				triangles = append(
					triangles,
					Line{
						start:	     start.sWest(),
						stop:	     start,
						orientation: NE,
					},
				)
			}
		case W:
			r := c.runeAt(start.west())
			if isDot(r) {
				needsNudging = true
			}
		case E:
			r := c.runeAt(start.east())
			if isDot(r) {
				needsNudging = true
			}
		}

		triangles = append(
			triangles,
			Triangle{
				start:	      start,
				orientation:  o,
				needsNudging: needsNudging,
			},
		)
	}
	return
}

// Circles returns a slice of all 'o' and '*' characters not considered text.
func (c *Canvas) Circles() (circles []Circle) {
	for idx := range upDownMinor(c.Width, c.Height) {
		// TODO INCOMING
		if c.runeAt(idx) == 'o' {
			circles = append(circles, Circle{start: idx})
		} else if c.runeAt(idx) == '*' {
			circles = append(circles, Circle{start: idx, bold: true})
		}
	}
	return
}

// RoundedCorners returns a slice of all curvy corners in the diagram.
func (c *Canvas) RoundedCorners() (corners []RoundedCorner) {
	for idx := range leftRightMinor(c.Width, c.Height) {
		if o := c.isRoundedCorner(idx); o != NONE {
			corners = append(
				corners,
				RoundedCorner{start: idx, orientation: o},
			)
		}
	}
	return
}

// For . and ' characters this will return a non-NONE orientation if the
// character encodes a rounded corner.
func (c *Canvas) isRoundedCorner(i Index) Orientation {
	r := c.runeAt(i)

	if !isJoint(r) {
		return NONE
	}

	left := i.west()
	right := i.east()
	lowerLeft := i.sWest()
	lowerRight := i.sEast()
	upperLeft := i.nWest()
	upperRight := i.nEast()

	opensUp := r == '\'' || r == '+'
	opensDown := r == '.' || r == '+'

	dashRight := c.runeAt(right) == '-' || c.runeAt(right) == '+' || c.runeAt(right) == '_' || c.runeAt(upperRight) == '_'
	dashLeft := c.runeAt(left) == '-' || c.runeAt(left) == '+' || c.runeAt(left) == '_' || c.runeAt(upperLeft) == '_'

	isVerticalSegment := func(i Index) bool {
		r := c.runeAt(i)
		return r == '|' || r == '+' || r == ')' || r == '(' || isDot(r)
	}

	//  .- or  .-
	// |	  +
	if opensDown && dashRight && isVerticalSegment(lowerLeft) {
		return NW
	}

	// -. or -.  or -.  or _.  or -.
	//   |	   +	  )	 )	o
	if opensDown && dashLeft && isVerticalSegment(lowerRight) {
		return NE
	}

	//   | or   + or   | or	  + or	 + or_ )
	// -'	  -'	 +'	+'     ++     '
	if opensUp && dashLeft && isVerticalSegment(upperRight) {
		return SE
	}

	// |  or +
	//  '-	  '-
	if opensUp && dashRight && isVerticalSegment(upperLeft) {
		return SW
	}

	return NONE
}


// Bridges returns a slice of all bridges, "-)-" or "-(-", composed as a sequence of
// either type Bridge or type Line.
func (c *Canvas) Bridges() (bridges []Drawable) {
	for idx := range leftRightMinor(c.Width, c.Height) {
		if o := c.isBridge(idx); o != NONE {
			bridges = append(
				bridges,
				newHalfStep(idx.north(), S),
				newHalfStep(idx.south(), N),
				Bridge{
					start:	     idx,
					orientation: o,
				},
			)
		}
	}
	return
}

// -)- or -(- or
func (c *Canvas) isBridge(i Index) Orientation {
	r := c.runeAt(i)

	left := c.runeAt(i.west())
	right := c.runeAt(i.east())

	if left != '-' || right != '-' {
		return NONE
	}

	if r == '(' {
		return W
	}

	if r == ')' {
		return E
	}

	return NONE
}

// Returns true if it looks like this character belongs to anything besides a
// horizontal line. This is the context we use to determine if a reserved
// character is text or not.
func (c *Canvas) hasLineAboveOrBelow(i Index) bool {
	i_r := c.runeAt(i)

	switch i_r {
	case '*', 'o', '+', 'v', '^':
		return c.partOfDiagonalLine(i) || c.partOfVerticalLine(i)
	case '|':
		return c.partOfVerticalLine(i) || c.partOfRoundedCorner(i)
	case '/', '\\':
		return c.partOfDiagonalLine(i)
	case '-':
		return c.partOfRoundedCorner(i)
	case '(', ')':
		return c.partOfVerticalLine(i)
	}

	return false
}

// Returns true if a "|" segment passes through this index.
func (c *Canvas) partOfVerticalLine(i Index) bool {
	this := c.runeAt(i)
	north := c.runeAt(i.north())
	south := c.runeAt(i.south())

	jointAboveMe := this == '|' && isJoint(north)

	if north == '|' || jointAboveMe {
		return true
	}

	jointBelowMe := this == '|' && isJoint(south)

	if south == '|' || jointBelowMe {
		return true
	}

	return false
}

// Return true if a "--" segment passes through this index.
func (c *Canvas) partOfHorizontalLine(i Index) bool {
	return c.runeAt(i.east()) == '-' || c.runeAt(i.west()) == '-'
}

func (c *Canvas) partOfDiagonalLine(i Index) bool {
	r := c.runeAt(i)

	n := c.runeAt(i.north())
	s := c.runeAt(i.south())
	nw := c.runeAt(i.nWest())
	se := c.runeAt(i.sEast())
	ne := c.runeAt(i.nEast())
	sw := c.runeAt(i.sWest())

	switch r {
	// Diagonal segments can be connected to joint or other segments.
	case '/':
		return ne == r || sw == r || isJoint(ne) || isJoint(sw) || n == '\\' || s == '\\'
	case '\\':
		return nw == r || se == r || isJoint(nw) || isJoint(se) || n == '/' || s == '/'

	// For everything else just check if we have segments next to us.
	default:
		return nw == '\\' || ne == '/' || sw == '/' || se == '\\'
	}
}

// For "-" and "|" characters returns true if they could be part of a rounded
// corner.
func (c *Canvas) partOfRoundedCorner(i Index) bool {
	r := c.runeAt(i)

	switch r {
	case '-':
		dotNext := c.runeAt(i.west()) == '.' || c.runeAt(i.east()) == '.'
		hyphenNext := c.runeAt(i.west()) == '\'' || c.runeAt(i.east()) == '\''
		return dotNext || hyphenNext

	case '|':
		dotAbove := c.runeAt(i.nWest()) == '.' || c.runeAt(i.nEast()) == '.'
		hyphenBelow := c.runeAt(i.sWest()) == '\'' || c.runeAt(i.sEast()) == '\''
		return dotAbove || hyphenBelow
	}

	return false
}

// TODO: Have this take care of all the vertical line nudging.
func (c *Canvas) partOfHalfStep(i Index) Orientation {
	r := c.runeAt(i)
	if r != '\'' && r != '.' && r != '|' {
		return NONE
	}

	if c.isRoundedCorner(i) != NONE {
		return NONE
	}

	w := c.runeAt(i.west())
	e := c.runeAt(i.east())
	n := c.runeAt(i.north())
	s := c.runeAt(i.south())
	nw := c.runeAt(i.nWest())
	ne := c.runeAt(i.nEast())

	switch r {
	case '\'':
		//  _	   _
		//   '-	 -'
		if (nw == '_' && e == '-') || (w == '-' && ne == '_') {
			return N
		}
	case '.':
		// _.-	-._
		if (w == '-' && e == '_') || (w == '_' && e == '-') {
			return S
		}
	case '|':
		//// _	 _
		////  | |
		if n != '|' && (ne == '_' || nw == '_') {
			return N
		}

		if n == '-' {
			return N
		}

		//// _| |_
		if s != '|' && (w == '_' || e == '_') {
			return S
		}

		if s == '-' {
			return S
		}
	}
	return NONE
}
