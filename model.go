package main

import "fmt"

type Command interface {
	String() string
}

type Block interface {
	Lines() []Command
}

type StartCode struct{}
type EndCode struct{}

type Layer struct {
	blocks []Block
}

type UnifiedGCode struct {
	startCode Block
	layers    []*Layer
	endCode   Block
}

//

type basicCommand struct {
	operation string
	comment   string
}

func (c *basicCommand) String() string {
	if c.comment == "" {
		return fmt.Sprintf("%s\n", c.operation)
	}
	return fmt.Sprintf("%s ; %s\n", c.operation, c.comment)
}

type commentCommand struct {
	comment string
}

func (c *commentCommand) String() string {
	return fmt.Sprintf("; %s\n", c.comment)
}

type movementCommand struct {
	operation string
	comment   string
}

func (c *movementCommand) String() string {
	if c.comment == "" {
		return fmt.Sprintf("%s\n", c.operation)
	}
	return fmt.Sprintf("%s ; %s\n", c.operation, c.comment)
}

type metaBlock struct {
	lines []Command
}

func (b *metaBlock) Lines() []Command {
	return b.lines
}

type purgeBlock struct {
	lines []Command
}

func (b *purgeBlock) Lines() []Command {
	return b.lines
}

type modelBlock struct {
	lines []Command
}

func (b *modelBlock) Lines() []Command {
	return b.lines
}
