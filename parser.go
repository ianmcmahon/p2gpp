package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Parse the source file into a sort of tree of relocatable blocks of gcode.
// We are expecting layer markers, and use them as boundaries to split the file
// into start code, an array of layers, and end code.
// This function handles splitting on layer markers, and delegates finer parsing
// to parseLayer() and parseBlock().
func parseFile(filename string) (*UnifiedGCode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gCode := &UnifiedGCode{}

	layerRE := regexp.MustCompile(`^;\s*LAYER\s+(\d+)$`)

	buf := new(bytes.Buffer)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// check for ;LAYER markers
		match := layerRE.FindStringSubmatch(line)
		if len(match) > 1 {
			layerNum, err := strconv.ParseInt(match[1], 10, 32)
			if err != nil {
				return nil, err
			}

			if layerNum == 0 { // then we're finishing up the start block
				gCode.startCode, err = gCode.parseBlock(buf)
				if err != nil {
					return nil, err
				}
			} else { // we're finishing the previous layer
				layer, err := gCode.parseLayer(buf)
				if err != nil {
					return nil, err
				}
				gCode.layers = append(gCode.layers, layer)
			}
			buf.Reset()
		}
		// how am I gonna catch the end of the last layer? there's no marker for it

		fmt.Fprintln(buf, line)
	}

	return gCode, nil
}

// parse a raw line into a Command object
func (g *UnifiedGCode) parseCommand(line string) Command {
	operation := strings.TrimSpace(line)
	comment := ""
	if i := strings.Index(line, ";"); i >= 0 {
		operation = strings.TrimSpace(line[0:i])
		comment = strings.TrimSpace(line[i+1:])
	}

	if operation == "" {
		return &commentCommand{comment}
	}

	cmd := &basicCommand{operation, comment}

	return cmd
}

// parse a bufferful of gcode into a Block, which
// is an ordered list of Commands
func (g *UnifiedGCode) parseBlock(in io.Reader) (Block, error) {
	block := &metaBlock{
		lines: make([]Command, 0),
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		command := g.parseCommand(scanner.Text())
		block.lines = append(block.lines, command)
	}

	return block, nil
}

// we expect a layer to contain model code and purge tower code.
// we split this apart and parse each block separately.
func (g *UnifiedGCode) parseLayer(in io.Reader) (*Layer, error) {
	// there are a few possible ways a layer can be laid out

	// FIRST LAYER
	// keychain has a toolchange on first layer
	// layer0 starts with ; CP WIPE TOWER FIRST LAYER BRIM
	// layer0 contains a CP TOOLCHANGE
	//
	// head has no toolchange on first layer
	// layer0 starts with ; CP WIPE TOWER FIRST LAYER BRIM
	// layer0 has immediately after a ; CP EMPTY GRID
	// not all layers start with purge
	//
	// there can be multiple model and purge blocks in a layer, and they need to happen in order

	layer := &Layer{
		blocks: make([]Block, 0),
	}

	brimStart := -1
	brimEnd := -1
	toolchangeStart := -1
	toolchangeEnd := -1
	emptyGridStart := -1
	emptyGridEnd := -1

	lines := make([]string, 0)
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		i := len(lines) - 1

		// I subtract or add 1 to each index to include the hbars around the block
		switch line {

		case "; CP WIPE TOWER FIRST LAYER BRIM START":
			brimStart = i - 1
			// walk back an additional 6 lines to catch retract and reposition
			// TODO: this needs to be done more reliably
			brimStart -= 6
			// everything up until now is a model block
			layer.blocks = append(layer.blocks, g.parseModelBlock(lines[0:brimStart]))
			lines = lines[brimStart:]

		case "; CP WIPE TOWER FIRST LAYER BRIM END":
			brimEnd = i + 1
			layer.blocks = append(layer.blocks, g.parsePurgeBlock(lines[0:brimEnd]))
			lines = lines[brimEnd:]

		case "; CP EMPTY GRID START":
			// this is a tough one, I have cases where there's a retract/move before,
			// and a case where there's a single move
			// for now I'll fail to catch anything before the hbar TODO
			emptyGridStart = i - 1
			// everything up to now is model
			layer.blocks = append(layer.blocks, g.parseModelBlock(lines[0:emptyGridStart]))
			lines = lines[emptyGridStart:]

		case "; CP EMPTY GRID END":
			emptyGridEnd = i + 1
			layer.blocks = append(layer.blocks, g.parsePurgeBlock(lines[0:emptyGridEnd]))
			lines = lines[emptyGridEnd:]

		case "; CP TOOLCHANGE START":
			// this needs some lookback too
			toolchangeStart = i - 1
			layer.blocks = append(layer.blocks, g.parseModelBlock(lines[0:toolchangeStart]))
			lines = lines[toolchangeStart:]

		case "; CP TOOLCHANGE END":
			toolchangeEnd = i + 1
			layer.blocks = append(layer.blocks, g.parsePurgeBlock(lines[0:toolchangeEnd]))
			lines = lines[toolchangeEnd:]

		}
	}

	// whatever's left over is model
	layer.blocks = append(layer.blocks, g.parseModelBlock(lines))

	return layer, nil
}

func (g *UnifiedGCode) parseModelBlock(lines []string) *modelBlock {
	b := &modelBlock{
		lines: make([]Command, 0),
	}
	for _, line := range lines {
		b.lines = append(b.lines, g.parseCommand(line))
	}

	return b
}

func (g *UnifiedGCode) parsePurgeBlock(lines []string) *purgeBlock {
	b := &purgeBlock{
		lines: make([]Command, 0),
	}
	for _, line := range lines {
		b.lines = append(b.lines, g.parseCommand(line))
	}

	return b
}
