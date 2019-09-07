package gcode

import (
	"fmt"
	"testing"
)

type testLine struct {
	line          string
	expectedGroup Group
}

// set up a list of test lines to parse, along with metadata about what to assert in the tests
var testLines = []testLine{
	{"M0", STOPPING},
	{"T0", TOOLCHANGE},
	{"G4 S0 ; Dwell", NON_MODAL},
	{"G1  Y142.400", MOTION},
	{"G1 Z0.40 F10800", MOTION},
	{"; --- P2PP Set wipe speed to 2000.0mm/s", UNKNOWN},
	{"G1 F4000.0", MOTION},
	{"G0 X230.082 Y142.4", MOTION},
	{"; CP TOOLCHANGE WIPE", UNKNOWN},
	{"G1  X181.000 E1.8654 F1600", MOTION},
	{"G1  X180.250 Y142.900  E-0.0343", MOTION},
	{"G1  X239.000 E2.2329 F1800", MOTION},
}

func TestParseStatement(t *testing.T) {
	E := 0.0
	for _, line := range testLines {
		stmt, err := ParseStatement(line.line)
		if err != nil {
			t.Error(err)
			continue
		}

		fmt.Printf("%s\n", stmt)
		E += stmt.Moved("E")
	}

	fmt.Printf("Total E travel: %.4f\n", E)
	if E != 4.064 {
		t.Errorf("test code E travel summed to %.4f, expected %.4f", E, 4.064)
	}
}

func TestGroupClassification(t *testing.T) {
	for _, line := range testLines {
		stmt, err := ParseStatement(line.line)
		if err != nil {
			t.Error(err)
			continue
		}

		if stmt.Group() != line.expectedGroup {
			t.Errorf("%s should classify as group %v, but was group %v instead", stmt, stmt.Group(), line.expectedGroup)
		}
	}
}
