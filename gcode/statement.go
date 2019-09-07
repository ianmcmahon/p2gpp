package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Statement struct {
	command string
	params  map[string]float64
	comment string
}

func (s *Statement) IsMotion() bool {
	return s.Group() == MOTION
}

// distance moved in one axis by this statement.
// Useful for E distance, less useful for XYZ because
// you probably want the cartesian distance.
func (s *Statement) Moved(axis string) float64 {
	// probably unnecessary, but if there are non-motion commands that have XYZ or E params,
	// make sure we don't return their value
	if !s.IsMotion() {
		return 0.0
	}

	if v, ok := s.params[axis]; ok {
		return v
	}
	return 0.0
}

// any time a Statement is treated like a string, this will render it as output-ready g-code
func (s *Statement) String() string {
	words := []string{}

	if s.command != "" {
		words = append(words, s.command)
	}

	for k, v := range s.params {
		words = append(words, fmt.Sprintf("%s%.4f", k, v))
	}

	if s.comment != "" {
		words = append(words, fmt.Sprintf("; %s", s.comment))
	}
	return strings.Join(words, " ")
}

// parse a raw line into a Statement object
func parseStatement(line string) *Statement {
	code := strings.TrimSpace(line)
	comment := ""
	if i := strings.Index(line, ";"); i >= 0 {
		code = strings.TrimSpace(line[0:i])
		comment = strings.TrimSpace(line[i+1:])
	}

	stmt := &Statement{
		params:  make(map[string]float64, 0),
		comment: comment,
	}

	if code == "" {
		return stmt
	}

	// a statement is a sequence of words, not necessarily space separated
	// each word is a single letter, case insensitive, followed by a single number
	// the first word is a command word, and its number must be a positive integer
	// or real.
	// Real beause some command words exist, like G92.2.  I'm not sure if they're
	// relevant to slicer output, but I'm supporting them.
	// All subsequent words are parameters, and can be positive or negative reals,
	// with or without decimal point.

	// this regex says "line starts with a group which is G, M, or T followed by a
	// positive real,followed by zero or more whitespace,
	// optionally followed by a group that starts with A-Z.
	commandRE := regexp.MustCompile(`^([GMTgmt][\d.]+)\s*([[:alpha:]].*)?`)

	// this regex says "find any group which starts with A-Z followed by zero or more whitespace
	// followed by an optional - followed by any number of digits or decimal.
	// it captures the alpha and the number into separate groups.
	// note that this regex will accept invalid code such as X0.1.2, but subsequent strconv won't.
	paramRE := regexp.MustCompile(`([[:alpha:]])\s*(-?[\d.]+)`)

	// first, pull the command word off the front
	matches := commandRE.FindStringSubmatch(code)
	if len(matches) != 3 { // [ match, group 1 (command), group 2 (params) ]
		panic(fmt.Errorf("bad match! '%s': %v", line, matches))
	}

	// TODO: this panics on bad parse right now, eventually this will need to return an error

	stmt.command = matches[1]

	// now, find all the param groups in what's left
	paramMatches := paramRE.FindAllStringSubmatch(matches[2], -1)
	for _, m := range paramMatches {
		letter := m[1]
		number := m[2]
		value, err := strconv.ParseFloat(number, 64)
		if err != nil {
			panic(fmt.Errorf("bad param! '%s': %s: %v\n", line, m[0], err))
		}
		stmt.params[letter] = value
	}

	return stmt
}
