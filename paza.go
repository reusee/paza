package paza

import (
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
	"unicode/utf8"
)

var (
	pt = fmt.Printf
)

type Set struct {
	parsers   map[string]Parser
	recursive map[string]bool
	serial    uint64
}

type Parser func(input *Input, start int) (ok bool, n int)

type stackEntry struct {
	parser string
	start  int
	ok     bool
	length int
}

type Input struct {
	text  []byte
	stack []stackEntry
}

func NewInput(text []byte) *Input {
	return &Input{
		text: text,
	}
}

func NewSet() *Set {
	return &Set{
		parsers:   make(map[string]Parser),
		recursive: make(map[string]bool),
	}
}

func (s *Set) Add(name string, parser Parser) {
	s.parsers[name] = parser
}

func (s *Set) AddRec(name string, parser Parser) {
	s.parsers[name] = parser
	s.recursive[name] = true
}

func (s *Set) Regex(re string) Parser {
	regex := regexp.MustCompile(re)
	return func(input *Input, start int) (bool, int) {
		if loc := regex.FindIndex(input.text[start:]); loc != nil && loc[0] == 0 {
			return true, loc[1]
		}
		return false, 0
	}
}

func (s *Set) Rune(r rune) Parser {
	return func(input *Input, start int) (bool, int) {
		ru, l := utf8.DecodeRune(input.text[start:])
		if ru == utf8.RuneError {
			return false, 0
		}
		if ru != r {
			return false, 0
		}
		return true, l
	}
}

func (s *Set) Call(name string, input *Input, start int) (bool, int) {
	fn, ok := s.parsers[name]
	if !ok {
		panic("parser not found " + name)
	}

	// non recursive parser
	if _, ok := s.recursive[name]; !ok {
		return fn(input, start)
	}

	// search stack
	for i := len(input.stack) - 1; i >= 0; i-- {
		mem := input.stack[i]
		if mem.parser == name && mem.start == start { // found
			return mem.ok, mem.length
		}
	}
	// not found, append a new entry
	entry := stackEntry{
		parser: name,
		start:  start,
		ok:     false,
		length: 0,
	}
	input.stack = append(input.stack, entry)
	for { // find the right bound
		stackSize := len(input.stack) // save stack size
		ok, l := fn(input, start)
		input.stack = input.stack[:stackSize] // unwind stack
		if !ok {
			return false, 0
		}
		if l < entry.length { // over bound
			return entry.ok, entry.length
		} else if l == entry.length { // not extending
			return ok, l
		}
		entry.ok = ok
		entry.length = l
		// update stack
		for i := len(input.stack) - 1; i >= 0; i-- {
			e := input.stack[i]
			if e.parser == name && e.start == start {
				input.stack[i].ok = ok
				input.stack[i].length = l
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
			panic(fmt.Sprintf("unknown parser type %T", parser))
		}
	}
	return
}

func (s *Set) Concat(parsers ...interface{}) Parser {
	names := s.getNames(parsers)
	return func(input *Input, start int) (bool, int) {
		index := start
		for _, name := range names {
			if ok, l := s.Call(name, input, index); !ok {
				return false, 0
			} else {
				index += l
			}
		}
		return true, index - start
	}
}

func (s *Set) OrdChoice(parsers ...interface{}) Parser {
	names := s.getNames(parsers)
	return func(input *Input, start int) (bool, int) {
		for _, name := range names {
			if ok, l := s.Call(name, input, start); ok {
				return ok, l
			}
		}
		return false, 0
	}
}
