package paza

import (
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
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

type memoryKey struct {
	parser string
	start  int
}

type memoryValue struct {
	ok     bool
	length int
}

type Input struct {
	text   []byte
	memory map[memoryKey]memoryValue
}

func NewInput(text []byte) *Input {
	return &Input{
		text:   text,
		memory: make(map[memoryKey]memoryValue),
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

func (s *Set) AddRegex(name, re string) {
	regex := regexp.MustCompile(re)
	s.Add(name, func(input *Input, start int) (bool, int) {
		if loc := regex.FindIndex(input.text[start:]); loc != nil && loc[0] == 0 {
			return true, loc[1]
		}
		return false, 0
	})
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

	// recursive parser
	key := memoryKey{name, start}
	memo, ok := input.memory[key]
	if !ok { // first call
		mem := memoryValue{ // 0-bound, always fail
			ok:     false,
			length: 0,
		}
		input.memory[key] = mem
		for {
			ok, l := fn(input, start) // try to increase bound
			if !ok {
				return false, 0
			}
			if l < mem.length { // use last bound
				return mem.ok, mem.length
			} else if l == mem.length { // not extending
				return ok, l
			}
			mem = memoryValue{ // update
				ok:     ok,
				length: l,
			}
			input.memory[key] = mem
		}
	} else { // not first call, return memory
		return memo.ok, memo.length
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
