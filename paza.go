package paza

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
)

var (
	pt = fmt.Printf
)

type Set struct {
	parsers map[string]Parser
	serial  uint64
}

type Node struct {
	Name  string
	Start int
	Len   int
	Subs  []*Node
}

type Parser func(input *Input, start int) (ok bool, n int, node *Node)

type stackEntry struct {
	parser string
	start  int
	ok     bool
	length int
	node   *Node
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
		parsers: make(map[string]Parser),
	}
}

func (s *Set) Add(name string, parser Parser) {
	s.parsers[name] = parser
}

func (s *Set) Call(name string, input *Input, start int) (retOk bool, retLen int, retNode *Node) {
	if start >= len(input.Text) {
		return false, 0, nil
	}
	parser, ok := s.parsers[name]
	if !ok {
		panic("parser not found: " + name)
	}

	defer func() {
		if retNode != nil {
			retNode.Name = name
		}
	}()

	// search stack
	for i := len(input.stack) - 1; i >= 0; i-- {
		mem := input.stack[i]
		if mem.parser == name && mem.start == start { // found
			return mem.ok, mem.length, mem.node
		}
	}
	// not found, append a new entry
	input.stack = append(input.stack, stackEntry{
		parser: name,
		start:  start,
		ok:     false,
		length: 0,
		node:   nil,
	})
	// find the right bound
	lastOk := false
	lastLen := 0
	var lastNode *Node
	stackSize := len(input.stack) // save stack size
	for {
		ok, l, node := parser(input, start)
		input.stack = input.stack[:stackSize] // unwind stack
		if !ok {
			return false, 0, nil
		}
		if l < lastLen { // over bound
			return lastOk, lastLen, lastNode
		} else if l == lastLen { // not extending
			return ok, l, node
		}
		lastOk = ok
		lastLen = l
		lastNode = node
		// update stack
		for i := len(input.stack) - 1; i >= 0; i-- {
			e := input.stack[i]
			if e.parser == name && e.start == start {
				input.stack[i].ok = ok
				input.stack[i].length = l
				input.stack[i].node = node
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

func (n *Node) dump(writer io.Writer, input *Input, level int) {
	start := n.Start
	end := n.Start + n.Len
	fmt.Fprintf(writer, "%s%q %s %d-%d\n", strings.Repeat("  ", level), input.Text[start:end], n.Name, start, end)
	for _, sub := range n.Subs {
		sub.dump(writer, input, level+1)
	}
}

func (n *Node) Dump(writer io.Writer, input *Input) {
	n.dump(writer, input, 0)
}

func (n *Node) Equal(n2 *Node) bool {
	if n.Name != n2.Name {
		return false
	}
	if n.Start != n2.Start {
		return false
	}
	if n.Len != n2.Len {
		return false
	}
	if len(n.Subs) != len(n2.Subs) {
		return false
	}
	for i, sub := range n.Subs {
		if !sub.Equal(n2.Subs[i]) {
			return false
		}
	}
	return true
}
