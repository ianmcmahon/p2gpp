package gcode

import (
	"fmt"
	"strings"
)

type Statement struct {
	command string
	params  map[string]float64
	comment string
}

// any time a Statement is treated like a string, this will render it as output-ready g-code
func (s *Statement) String() string {
	words := []string{}

	if s.command != "" {
		// this kind of special casing is why I like the idea of having a special ToolchangeStatement object that renders itself differently
		if s.IsToolchange() && s.command == "T" {
			words = append(words, fmt.Sprintf("T%d", int(s.params["T"])))
		} else {
			words = append(words, s.command)

			for k, v := range s.params {
				words = append(words, fmt.Sprintf("%s%.4f", k, v))
			}
		}
	}

	if s.comment != "" {
		words = append(words, fmt.Sprintf("; %s", s.comment))
	}
	return strings.Join(words, " ")
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

func (s *Statement) Tool() int {
	if !s.IsToolchange() {
		return -1
	}

	return int(s.params["T"])
}
