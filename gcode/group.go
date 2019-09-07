package gcode

import "sync"

/*
	This is some syntactic sugar for classifying commands into established groups.
	This comes from https://github.com/synthetos/g2/wiki/GCode-Parsing, who purports
	to have taken it from the NIST standard.

	How this works is simple, I'm creating a type called Group which is just an alias
	for int.  The const section declares all the Group constants, and numbers them
	from zero up.  groupMemberships maps each group to a list of commands in the group.

	The Group() method is bound to Statement, and it runs a function which initializes
	a reverse map based on the groupMemberships map.  This is run via a sync.Once object,
	which ensures that the initialization only happens the first time it's called on any
	Statement object.  This is thread safe, although I don't expect concurrency in this.

	This initialization could have also been done in func init(), which would happen at
	startup.  I used sync.Once because I like the lazy loading aspect.
*/

type Group int

const (
	UNKNOWN Group = iota
	NON_MODAL
	MOTION
	PLANE_SELECTION
	DISTANCE_MODE
	FEED_RATE_MODE
	UNITS
	CUTTER_RAD_COMP
	TOOL_LENGTH_OFFSET
	RETURN_MODE
	CS_SEL
	PATH_CONTROL_MODE
	STOPPING
	TOOLCHANGE
	SPINDLE_TURNING
	COOLANT
	FRO_ENABLE
)

var groupMemberships = map[Group][]string{
	// non-modals
	NON_MODAL: []string{"G4", "G10", "G28", "G30", "G53", "G92", "G92.1", "G92.2", "G92.3"},
	// G groups
	MOTION:             []string{"G0", "G1", "G2", "G3", "G38.2", "G80", "G81", "G82", "G83", "G84", "G85", "G86", "G87", "G88", "G89"},
	PLANE_SELECTION:    []string{"G17", "G18", "G19"},
	DISTANCE_MODE:      []string{"G90", "G91"},
	FEED_RATE_MODE:     []string{"G93", "G94"},
	UNITS:              []string{"G20", "G21"},
	CUTTER_RAD_COMP:    []string{"G40", "G41", "G42"},
	TOOL_LENGTH_OFFSET: []string{"G43", "G49"},
	RETURN_MODE:        []string{"G98", "G99"},
	CS_SEL:             []string{"G54", "G55", "G56", "G57", "G58", "G59", "G59.1", "G59.2", "G59.3"},
	PATH_CONTROL_MODE:  []string{"G61", "G61.1", "G64"},

	// M group
	STOPPING:        []string{"M0", "M1", "M2", "M30", "M60"},
	TOOLCHANGE:      []string{"M6", "T"},
	SPINDLE_TURNING: []string{"M3", "M4", "M5"},
	COOLANT:         []string{"M7", "M8", "M9"}, // special case: M7 and M8 may be active at the same time
	FRO_ENABLE:      []string{"M48", "M49"},
}

var once sync.Once
var commandToGroup map[string]Group

func (s *Statement) Group() Group {
	once.Do(func() {
		commandToGroup = make(map[string]Group, 0)
		for group, commands := range groupMemberships {
			for _, command := range commands {
				commandToGroup[command] = group
			}
		}
	})

	if group, ok := commandToGroup[s.command]; ok {
		return group
	}
	return UNKNOWN
}

func (s *Statement) IsMotion() bool {
	return s.Group() == MOTION
}

func (s *Statement) IsToolchange() bool {
	return s.Group() == TOOLCHANGE
}
