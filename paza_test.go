package paza

import "testing"

func TestAll(t *testing.T) {
	set := NewSet()
	set.Add("a", set.Regex(`a`))
	set.Add("+", set.Regex(`\+`))
	set.AddRec("expr", set.OrdChoice(set.Concat("expr", "+", "a"), "a"))

	cases := []struct {
		text   []byte
		parser string
		ok     bool
		length int
	}{
		{[]byte(""), "a", false, 0},
		{[]byte(""), "+", false, 0},
		{[]byte(""), "expr", false, 0},
		{[]byte("a"), "a", true, 1},
		{[]byte("a "), "a", true, 1},
		{[]byte("b"), "a", false, 0},
		{[]byte("+"), "+", true, 1},
		{[]byte("+b"), "+", true, 1},
		{[]byte("b"), "+", false, 0},
		{[]byte("a"), "expr", true, 1},
		{[]byte("a+"), "expr", true, 1},
		{[]byte("a+a"), "expr", true, 3},
		{[]byte("a+a+a+a+a"), "expr", true, 9},
		{[]byte("a+a+a+a+a+"), "expr", true, 9},
		{[]byte("a+a+a+a+a+a"), "expr", true, 11},
	}

	for _, c := range cases {
		input := NewInput(c.text)
		ok, l := set.Call(c.parser, input, 0)
		if c.ok != ok || c.length != l {
			t.Fatalf("%v", c)
		}
	}
}

func TestCalc(t *testing.T) {
	set := NewSet()
	set.AddRec("expr", set.OrdChoice(
		set.Concat("expr", set.Rune('+'), "term"),
		set.Concat("expr", set.Rune('-'), "term"),
		"term",
	))
	set.AddRec("term", set.OrdChoice(
		set.Concat("term", set.Rune('*'), "factor"),
		set.Concat("term", set.Rune('/'), "factor"),
		"factor",
	))
	set.Add("factor", set.OrdChoice(
		set.Regex(`[0-9]+`),
		set.Concat(set.Rune('('), "expr", set.Rune(')')),
	))

	cases := []struct {
		text   []byte
		parser string
		ok     bool
		length int
	}{
		{[]byte("1"), "expr", true, 1},
		{[]byte("1+1"), "expr", true, 3},
		{[]byte("1-1"), "expr", true, 3},
		{[]byte("1*1"), "expr", true, 3},
		{[]byte("1/1"), "expr", true, 3},
		{[]byte("(1/1)"), "expr", true, 5},
		{[]byte("(1)/1"), "expr", true, 5},
		{[]byte("(1)/1*3"), "expr", true, 7},
		{[]byte("(1)/1*(3-2)"), "expr", true, 11},
		{[]byte("(1)/1**(3-2)"), "expr", true, 5},
		{[]byte("*(1)/1**(3-2)"), "expr", false, 0},
		{[]byte(""), "expr", false, 0},
	}

	for _, c := range cases {
		input := NewInput(c.text)
		ok, l := set.Call(c.parser, input, 0)
		if c.ok != ok || c.length != l {
			t.Fatalf("%v", c)
		}
	}
}

func TestRegex(t *testing.T) {
	/*
		<RE>	::=	<RE> "|" <simple-RE> | <simple-RE>
		<simple-RE>	::=	<simple-RE> <basic-RE> | <basic-RE>
		<basic-RE>	::=	<elementary-RE> "*" | <elementary-RE> "+" | <elementary-RE>
		<elementary-RE>	::=	"(" <RE> ")" | "." | "$" | [a-zA-Z0-9]
	*/
	set := NewSet()
	set.AddRec("re", set.OrdChoice(
		set.Concat("re", set.Rune('|'), "simple-re"),
		"simple-re",
	))
	set.AddRec("simple-re", set.OrdChoice(
		set.Concat("simple-re", "basic-re"),
		"basic-re",
	))
	set.Add("basic-re", set.OrdChoice(
		set.Concat("elementary-re", set.Rune('*')),
		set.Concat("elementary-re", set.Rune('+')),
		"elementary-re",
	))
	set.Add("elementary-re", set.OrdChoice(
		set.Concat(set.Rune('('), "re", set.Rune(')')),
		set.Rune('.'),
		set.Rune('$'),
		set.Regex(`[a-zA-Z0-9]`),
	))

	cases := []struct {
		text   []byte
		parser string
		ok     bool
		length int
	}{
		{[]byte(""), "re", false, 0},
		{[]byte("a"), "re", true, 1},
		{[]byte("a*"), "re", true, 2},
		{[]byte("a.*"), "re", true, 3},
		{[]byte("a(.*)"), "re", true, 5},
		{[]byte("a(.*)+"), "re", true, 6},
		{[]byte("a(.*)+$"), "re", true, 7},
		{[]byte("a(.*)+$b+"), "re", true, 9},
		{[]byte("a(.*)+$|b+"), "re", true, 10},
		{[]byte("a(.*)+$|*b+"), "re", true, 7},
	}

	for _, c := range cases {
		input := NewInput(c.text)
		ok, l := set.Call(c.parser, input, 0)
		if c.ok != ok || c.length != l {
			t.Fatalf("%v", c)
		}
	}
}

func TestIndirect(t *testing.T) {
	/*
		L = P.x | x
		P = P(n) | L
	*/
	set := NewSet()
	set.AddRec("L", set.OrdChoice(
		set.Concat("P", set.Rune('.'), set.Rune('x')),
		set.Rune('x'),
	))
	set.AddRec("P", set.OrdChoice(
		set.Concat("P", set.Rune('('), set.Rune('n'), set.Rune(')')),
		"L",
	))

	cases := []struct {
		text   []byte
		parser string
		ok     bool
		length int
	}{
		{[]byte("x"), "L", true, 1},
		{[]byte("x(n)(n).x(n).x"), "L", true, 14},
	}

	for _, c := range cases {
		input := NewInput(c.text)
		ok, l := set.Call(c.parser, input, 0)
		if c.ok != ok || c.length != l {
			t.Fatalf("%v %v %d", c, ok, l)
		}
	}
}

func TestIndirect2(t *testing.T) {
	/*
		A = Ba | d
		B = Cb | e
		C = Ac | f
	*/
	set := NewSet()
	set.AddRec("A", set.OrdChoice(
		set.Concat("B", set.Rune('a')),
		set.Rune('d'),
	))
	set.AddRec("B", set.OrdChoice(
		set.Concat("C", set.Rune('b')),
		set.Rune('e'),
	))
	set.AddRec("C", set.OrdChoice(
		set.Concat("A", set.Rune('c')),
		set.Rune('f'),
	))

	cases := []struct {
		text   []byte
		parser string
		ok     bool
		length int
	}{
		{[]byte("d"), "A", true, 1},
		{[]byte("e"), "B", true, 1},
		{[]byte("f"), "C", true, 1},
		{[]byte("ea"), "A", true, 2},
		{[]byte("fb"), "B", true, 2},
		{[]byte("dc"), "C", true, 2},
		{[]byte("fba"), "A", true, 3},
		{[]byte("dcb"), "B", true, 3},
		{[]byte("eac"), "C", true, 3},
		{[]byte("dcba"), "A", true, 4},
		{[]byte("eacb"), "B", true, 4},
		{[]byte("fbac"), "C", true, 4},
	}

	for _, c := range cases {
		input := NewInput(c.text)
		ok, l := set.Call(c.parser, input, 0)
		if c.ok != ok || c.length != l {
			t.Fatalf("%v %v %d", c, ok, l)
		}
	}
}
