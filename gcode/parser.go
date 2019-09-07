package gcode

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// parse a raw line into a Statement object
func ParseStatement(line string) (*Statement, error) {
	code := strings.TrimSpace(line)
	comment := ""
	if i := strings.Index(line, ";"); i >= 0 {
		code = strings.TrimSpace(line[0:i])
		comment = strings.TrimSpace(line[i+1:])
	}

	// upcasing the code here makes parsing simpler, and all downstream code can use exclusively upcase
	code = strings.ToUpper(code)

	stmt := &Statement{
		params:  make(map[string]float64, 0),
		comment: comment,
	}

	if code == "" {
		return stmt, nil
	}

	// a statement is a sequence of words, not necessarily space separated
	// each word is a single letter, followed by a single number
	// the first word is a command word, and its number must be a positive integer
	// or real.
	// Real beause some command words exist, like G92.2.  I'm not sure if they're
	// relevant to slicer output, but I'm supporting them.
	// All subsequent words are parameters, and can be positive or negative reals,
	// with or without decimal point.

	// catching toolchanges with a special regex
	toolchangeRE := regexp.MustCompile(`^(T)(\d+)`)

	// this regex says "line starts with a group which is G, M, or T followed by a
	// positive real,followed by zero or more whitespace,
	// optionally followed by a group that starts with A-Z.
	// adding support for capturing O codes
	commandRE := regexp.MustCompile(`^([GMO][\d.]+)\s*([A-Z].*)?`)

	// this regex says "find any group which starts with A-Z followed by zero or more whitespace
	// followed by an optional - followed by any number of digits or decimal.
	// it captures the alpha and the number into separate groups.
	// note that this regex will accept invalid code such as X0.1.2, but subsequent strconv won't.
	paramRE := regexp.MustCompile(`([A-Z])\s*(-?[\d.]+)`)

	tool := toolchangeRE.FindStringSubmatch(code)
	if len(tool) > 0 {
		// have to do a special thing for Tx toolchange commands.
		// Turn them into a command of T with a param of T = x
		stmt.command = tool[1] // this should be "T"
		toolNum, err := strconv.ParseInt(tool[2], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("unparsable tool number: '%s' %v", tool[2], err)
		}
		stmt.params[tool[1]] = float64(toolNum)
		return stmt, nil
	}

	// first, pull the command word off the front
	matches := commandRE.FindStringSubmatch(code)
	if len(matches) != 3 { // [ match, group 1 (command), group 2 (params) ]
		return nil, fmt.Errorf("bad match! '%s': %v", line, matches)
	}

	stmt.command = matches[1]

	// now, find all the param groups in what's left
	paramMatches := paramRE.FindAllStringSubmatch(matches[2], -1)
	for _, m := range paramMatches {
		letter := m[1]
		number := m[2]
		value, err := strconv.ParseFloat(number, 64)
		if err != nil {
			return nil, fmt.Errorf("bad param! '%s': %s: %v", line, m[0], err)
		}
		stmt.params[letter] = value
	}

	return stmt, nil
}
