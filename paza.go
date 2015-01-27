package paza

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

var (
	pt = fmt.Printf
)

type parserInfo struct {
	parser    Parser
	recursive bool
}

type Set struct {
	parsers map[string]parserInfo
	serial  uint64
	enter   func(name string, input *Input, start int)
	leave   func(name string, input *Input, start int, ok bool, length int)
}

type Parser func(input *Input, start int) (ok bool, n int)

type stackEntry struct {
	parser string
	start  int
	ok     bool
	length int
}

type Input struct {
	Text  []byte
	stack []stackEntry
}

func NewInput(text []byte) *Input {
	return &Input{
		Text: text,
	}
}

func NewSet() *Set {
	return &Set{
		parsers: make(map[string]parserInfo),
	}
}

func (s *Set) SetEnter(fn func(string, *Input, int)) {
	s.enter = fn
}

func (s *Set) SetLeave(fn func(string, *Input, int, bool, int)) {
	s.leave = fn
}

func (s *Set) Add(name string, parser Parser) {
	s.parsers[name] = parserInfo{parser, false}
}

func (s *Set) AddRec(name string, parser Parser) {
	s.parsers[name] = parserInfo{parser, true}
}

func (s *Set) Call(name string, input *Input, start int) (retOk bool, retLen int) {
	/* TODO
	pt("=> call %s %d\n", name, start)
	defer func() {
		pt("<- result %s %d %v %d\n", name, start, retOk, retLen)
	}()
	*/

	if start >= len(input.Text) {
		return false, 0
	}
	info, ok := s.parsers[name]
	if !ok {
		panic("parser not found: " + name)
	}

	// non recursive parser
	if !info.recursive {
		if s.enter != nil {
			s.enter(name, input, start)
		}
		defer func() {
			if s.leave != nil {
				s.leave(name, input, start, retOk, retLen)
			}
		}()
		return info.parser(input, start)
	}

	// search stack
	for i := len(input.stack) - 1; i >= 0; i-- {
		mem := input.stack[i]
		if mem.parser == name && mem.start == start { // found
			return mem.ok, mem.length
		}
	}
	// not found, append a new entry
	input.stack = append(input.stack, stackEntry{
		parser: name,
		start:  start,
		ok:     false,
		length: 0,
	})
	if s.enter != nil {
		s.enter(name, input, start)
	}
	defer func() {
		if s.leave != nil {
			s.leave(name, input, start, retOk, retLen)
		}
	}()
	// find the right bound
	lastOk := false
	lastLen := 0
	stackSize := len(input.stack) // save stack size
	for {
		ok, l := info.parser(input, start)
		input.stack = input.stack[:stackSize] // unwind stack
		if !ok {
			return false, 0
		}
		if l < lastLen { // over bound
			return lastOk, lastLen
		} else if l == lastLen { // not extending
			return ok, l
		}
		lastOk = ok
		lastLen = l
		// update stack
		for i := len(input.stack) - 1; i >= 0; i-- {
			e := input.stack[i]
			if e.parser == name && e.start == start {
				input.stack[i].ok = ok
				input.stack[i].length = l
				break
			}
		}
	}

}

func (s *Set) getNames(parsers []interface{}) (ret []string) {
	for _, parser := range parsers {
		switch parser := parser.(type) {
		case string:
			ret = append(ret, parser)
		case Parser:
			name := "__parser__" + strconv.Itoa(int(atomic.AddUint64(&s.serial, 1)))
			s.Add(name, parser)
			ret = append(ret, name)
		default:
			panic(fmt.Sprintf("unknown parser type: %T", parser))
		}
	}
	return
}
