package paza

import "testing"

func TestAll(t *testing.T) {
	set := NewSet()
	set.Add("a", set.Regex(`a`))
	set.Add("+", set.Regex(`\+`))
	// direct recursive
	set.AddRec("expr", set.OrdChoice(set.Concat("expr", "+", "a"), "a"))
	// TODO indirect recursive

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
