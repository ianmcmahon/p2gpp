package main

import "fmt"

func main() {
	gcode, err := parseFile("keychain-unprocessed.gcode")
	if err != nil {
		panic(err)
	}

	fmt.Printf("startCode:\n")
	for _, stmt := range gcode.startCode.Lines() {
		fmt.Printf("\t%#v\n", stmt)
	}

	for i, layer := range gcode.layers {
		fmt.Printf("layer %d:\n", i)
		for b, block := range layer.blocks {
			fmt.Printf("\tblock %d:\n", b)
			for _, stmt := range block.Lines() {
				fmt.Printf("\t\t%#v\n", stmt)
			}
		}
	}
	fmt.Printf("endCode: %#v\n", gcode.endCode)
}
