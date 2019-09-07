package main

import "fmt"

type Statement struct {
	code    string
	comment string

	command string
	params  map[string]float64
}

func (s *Statement) String() string {
	if s.comment == "" {
		return fmt.Sprintf("%s\n", s.code)
	}
	return fmt.Sprintf("%s ; %s\n", s.code, s.comment)
}

func (s *Statement) IsMove() bool {
	return false
}

type Block interface {
	Lines() []*Statement
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

type metaBlock struct {
	lines []*Statement
}

func (b *metaBlock) Lines() []*Statement {
	return b.lines
}

type purgeBlock struct {
	lines []*Statement
}

func (b *purgeBlock) Lines() []*Statement {
	return b.lines
}

type modelBlock struct {
	lines []*Statement
}

func (b *modelBlock) Lines() []*Statement {
	return b.lines
}
