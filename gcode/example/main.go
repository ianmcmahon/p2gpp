package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ianmcmahon/p2gpp/gcode"
)

type splice struct {
	tool     int
	position float64
	length   float64
}

func (s splice) end() float64 {
	return s.position + s.length
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <filename.gcode>\n", os.Args[0])
		os.Exit(1)
	}
	filename := os.Args[1]

	stmts, err := parseFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	splices := createSplices(stmts)
	for _, splice := range splices {
		fmt.Printf("%#v\n", splice)
	}
}

func parseFile(filename string) ([]*gcode.Statement, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	statements := make([]*gcode.Statement, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		stmts, err := gcode.ParseStatement(line)
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmts)
	}

	return statements, nil
}

func createSplices(stmts []*gcode.Statement) []splice {
	var splices []splice

	E := 0.0
	curTool := -1

	for _, stmt := range stmts {
		if stmt.IsToolchange() {
			if E == 0.0 {
				curTool = stmt.Tool()
				// if we see a toolchange before any E movement, no need to try to make a splice from it, just set curTool
				continue
			}

			// splice!
			if splices == nil { // then this is the first one!
				// this syntax is a little goofy, but it means "anonymous literal slice of splice structs, initialized with one
				// anonymous splice struct, with current tool, starting at position 0.0 (because it's the first splice), with length E."
				splices = []splice{splice{curTool, 0.0, E}}
			} else {
				lastSplice := splices[len(splices)-1]
				splices = append(splices, splice{curTool, lastSplice.end(), E})
			}

			curTool = stmt.Tool()
		}

		E += stmt.Moved("E")
	}

	return splices
}
