package paza

import (
	"regexp"
	"unicode/utf8"
)

func (s *Set) Regex(re string) Parser {
	regex := regexp.MustCompile(re)
	return func(input *Input, start int) (bool, int) {
		if loc := regex.FindIndex(input.Text[start:]); loc != nil && loc[0] == 0 {
			return true, loc[1]
		}
		return false, 0
	}
}

func (s *Set) NamedRegex(name string, re string) string {
	s.Add(name, s.Regex(re))
	return name
}

func (s *Set) Rune(r rune) Parser {
	return func(input *Input, start int) (bool, int) {
		ru, l := utf8.DecodeRune(input.Text[start:])
		if ru == utf8.RuneError {
			panic("utf8 decode error")
		}
		if ru != r {
			return false, 0
		}
		return true, l
	}
}

func (s *Set) NamedRune(name string, r rune) string {
	s.Add(name, s.Rune(r))
	return name
}

func (s *Set) ByteIn(bs []byte) Parser {
	return func(input *Input, start int) (bool, int) {
		b := input.Text[start]
		for _, bt := range bs {
			if bt == b {
				return true, 1
			}
		}
		return false, 0
	}
}

func (s *Set) NamedByteIn(name string, bs []byte) string {
	s.Add(name, s.ByteIn(bs))
	return name
}

func (s *Set) ByteRange(left, right byte) Parser {
	return func(input *Input, start int) (bool, int) {
		b := input.Text[start]
		if b >= left && b <= right {
			return true, 1
		}
		return false, 0
	}
}

func (s *Set) NamedByteRange(name string, left, right byte) string {
	s.Add(name, s.ByteRange(left, right))
	return name
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

func (s *Set) NamedConcat(name string, parsers ...interface{}) string {
	s.Add(name, s.Concat(parsers...))
	return name
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

func (s *Set) NamedOrdChoice(name string, parsers ...interface{}) string {
	s.Add(name, s.OrdChoice(parsers...))
	return name
}

func (s *Set) OneOrMore(parser interface{}) Parser {
	names := s.getNames([]interface{}{parser})
	name := names[0]
	return func(input *Input, start int) (bool, int) {
		index := start
		ok, l := s.Call(name, input, index)
		if !ok {
			return false, 0
		}
		index += l
		for {
			ok, l = s.Call(name, input, index)
			if ok {
				index += l
			} else {
				break
			}
		}
		return true, index - start
	}
}

func (s *Set) NamedOneOrMore(name string, parser interface{}) string {
	s.Add(name, s.OneOrMore(parser))
	return name
}
