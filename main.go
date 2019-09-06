package main

import "fmt"

func main() {
	gcode, err := parseFile("keychain-unprocessed.gcode")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", gcode)
}
