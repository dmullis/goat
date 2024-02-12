package goat

import (
	"log"
	"strings"
	"unicode"
)

type (
	anchorSelector int  // UTF-8 line number within source TXT

	anchorSet struct {
		// record number of line where found in TXT source
		Selectors []anchorSelector

		// for parsing text found in the diagram
		Opens,
		Closes map[rune]anchorSelector

		// for generating output SVG
		payload map[anchorSelector]svgAnchorPayload
	}

	svgAnchorPayload struct {
		Replacements [2]rune
		Class,
		DarkClass,
		Attributes string
	}
)

func findAnchorKey(wantedAI anchorSelector, runeMap map[rune]anchorSelector) (_ rune) {
	for r, aI := range runeMap {
		if aI == wantedAI {
			return r
		}
	}
	log.Panicln("internal error")
	return
}

func NewAnchorSet() anchorSet {
	return anchorSet{
		Selectors: make([]anchorSelector,0),

		Opens: make(map[rune]anchorSelector),
		Closes:   make(map[rune]anchorSelector),

		payload: make(map[anchorSelector]svgAnchorPayload),
	}
}

func isAnchorSpecifier(lineRunes []rune) bool {
	if len(lineRunes) < 6 {
		return false
	}
	return lineRunes[0] == '#' &&
		unicode.IsPrint(lineRunes[1]) &&
		unicode.IsPrint(lineRunes[2]) &&
		unicode.IsPrint(lineRunes[3]) &&
		unicode.IsPrint(lineRunes[4]) &&
		unicode.IsSpace(lineRunes[5])
}

func (c *Canvas) parseAnchorSpecifier(lineRunes []rune, newSelector anchorSelector) {
	aSet := &c.anchorSet
	aSet.Selectors = append(aSet.Selectors, newSelector)
	openRune, closeRune := lineRunes[1], lineRunes[2]

	aSet.Opens[openRune] = newSelector
	aSet.Closes[closeRune] = newSelector

	valueRunes := lineRunes[6:]

	logPanic := func(f string) {
		log.Panicf(`
	Field
		'%s'
	in line
		'%s'
	%s
	%s
	%s`,
			f,
			string(lineRunes),
			"unexpected.",
			"Recall that fields are divided by space characters -- quoting of",
			"embedded space is not supported.")
	}

	var classStr, darkClassStr, attrStr string
	fields := strings.Fields(string(valueRunes))
	darkStart := -1
	for iF, f := range fields {
		if f[0] == '#' {   // X  Compare zero-extended 'byte' to 'rune'
			// early exit -- trailing sh-style comment found
			break
		}
		// Is the field the dark-scheme switch?
		if f == "--" {
			darkStart = iF+1
			break
		}
		// Is the field an HTML attribute, to be added to the <a> element?
		//  X  Must test for "=" first, because value may contain ":"
		if _, _, found := strings.Cut(f, "="); found {
			attrStr += " " + f
			continue
		}
		// Is the field a CSS property to be added to the class definition?
		if _, _, found := strings.Cut(f, ":"); found {
			classStr += " " + f + ";"
			continue
		}
		logPanic(f)
	}
	if darkStart > 0 {
		for _, f := range fields[darkStart:] {
			if f[0] == '#' {
				break
			}
			// Is the field a CSS property to be added to the class definition?
			if _, _, found := strings.Cut(f, ":"); found {
				darkClassStr += " " + f + ";"
				continue
			}
			logPanic(f)
		}
	}
	aSet.payload[newSelector] = svgAnchorPayload{
		Replacements: [2]rune(lineRunes[3:5]),

		Class: classStr,
		DarkClass: darkClassStr,

		Attributes: attrStr,
	}
}
