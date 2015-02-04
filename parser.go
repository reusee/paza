package paza

import (
	"regexp"
	"unicode/utf8"
)

func (s *Set) Regex(re string) Parser {
	regex := regexp.MustCompile(re)
	return func(input *Input, start int) (bool, int, *Node) {
		if start >= input.Len() {
			return false, 0, nil
		}
		if loc := regex.FindIndex(input.BytesSlice(start, -1)); loc != nil && loc[0] == 0 {
			return true, loc[1], &Node{
				Start: start,
				Len:   loc[1],
			}
		}
		return false, 0, nil
	}
}

func (s *Set) NamedRegex(name string, re string) string {
	s.Add(name, s.Regex(re))
	return name
}

func (s *Set) Rune(r rune) Parser {
	return func(input *Input, start int) (bool, int, *Node) {
		if start >= input.Len() {
			return false, 0, nil
		}
		ru, l := utf8.DecodeRune(input.BytesSlice(start, -1))
		if ru == utf8.RuneError {
			panic("utf8 decode error")
		}
		if ru != r {
			return false, 0, nil
		}
		return true, l, &Node{
			Start: start,
			Len:   l,
		}
	}
}

func (s *Set) NamedRune(name string, r rune) string {
	s.Add(name, s.Rune(r))
	return name
}

func (s *Set) ByteIn(bs []byte) Parser {
	return func(input *Input, start int) (bool, int, *Node) {
		if start >= input.Len() {
			return false, 0, nil
		}
		b := input.At(start)[0]
		for _, bt := range bs {
			if bt == b {
				return true, 1, &Node{
					Start: start,
					Len:   1,
				}
			}
		}
		return false, 0, nil
	}
}

func (s *Set) NamedByteIn(name string, bs []byte) string {
	s.Add(name, s.ByteIn(bs))
	return name
}

func (s *Set) ByteRange(left, right byte) Parser {
	return func(input *Input, start int) (bool, int, *Node) {
		if start >= input.Len() {
			return false, 0, nil
		}
		b := input.At(start)[0]
		if b >= left && b <= right {
			return true, 1, &Node{
				Start: start,
				Len:   1,
			}
		}
		return false, 0, nil
	}
}

func (s *Set) NamedByteRange(name string, left, right byte) string {
	s.Add(name, s.ByteRange(left, right))
	return name
}

func (s *Set) Concat(parsers ...interface{}) Parser {
	names := s.getNames(parsers...)
	return func(input *Input, start int) (bool, int, *Node) {
		index := start
		var subs []*Node
		for _, name := range names {
			if ok, l, node := s.Call(name, input, index); !ok {
				return false, 0, nil
			} else {
				index += l
				subs = append(subs, node)
			}
		}
		return true, index - start, &Node{
			Start: start,
			Len:   index - start,
			Subs:  subs,
		}
	}
}

func (s *Set) NamedConcat(name string, parsers ...interface{}) string {
	s.Add(name, s.Concat(parsers...))
	return name
}

func (s *Set) OrdChoice(parsers ...interface{}) Parser {
	names := s.getNames(parsers...)
	return func(input *Input, start int) (bool, int, *Node) {
		for _, name := range names {
			if ok, l, node := s.Call(name, input, start); ok {
				return ok, l, &Node{
					Start: start,
					Len:   l,
					Subs:  []*Node{node},
				}
			}
		}
		return false, 0, nil
	}
}

func (s *Set) NamedOrdChoice(name string, parsers ...interface{}) string {
	s.Add(name, s.OrdChoice(parsers...))
	return name
}

func (s *Set) Repeat(lowerBound, upperBound int, parser interface{}) Parser {
	name := s.getNames(parser)[0]
	return func(input *Input, start int) (bool, int, *Node) {
		index := start
		var subs []*Node
		for {
			ok, l, node := s.Call(name, input, index)
			if ok {
				index += l
				subs = append(subs, node)
				if upperBound > 0 && len(subs) >= upperBound {
					break
				}
			} else {
				break
			}
		}
		if len(subs) < lowerBound {
			return false, 0, nil
		}
		return true, index - start, &Node{
			Start: start,
			Len:   index - start,
			Subs:  subs,
		}
	}
}

func (s *Set) NamedRepeat(name string, lowerBound, upperBound int, parser interface{}) string {
	s.Add(name, s.Repeat(lowerBound, upperBound, parser))
	return name
}

func (s *Set) OneOrMore(parser interface{}) Parser {
	return s.Repeat(1, -1, parser)
}

func (s *Set) NamedOneOrMore(name string, parser interface{}) string {
	s.Add(name, s.OneOrMore(parser))
	return name
}

func (s *Set) ZeroOrMore(parser interface{}) Parser {
	return s.Repeat(0, -1, parser)
}

func (s *Set) NamedZeroOrMore(name string, parser interface{}) string {
	s.Add(name, s.ZeroOrMore(parser))
	return name
}

func (s *Set) Predicate(parser interface{}) Parser {
	name := s.getNames(parser)[0]
	return func(input *Input, start int) (bool, int, *Node) {
		if ok, _, _ := s.Call(name, input, start); ok {
			return true, 0, nil
		}
		return false, 0, nil
	}
}

func (s *Set) NotPredicate(parser interface{}) Parser {
	name := s.getNames(parser)[0]
	return func(input *Input, start int) (bool, int, *Node) {
		if ok, _, _ := s.Call(name, input, start); !ok {
			return true, 0, nil
		}
		return false, 0, nil
	}
}
