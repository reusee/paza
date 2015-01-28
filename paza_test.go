package paza

import (
	"bytes"
	"testing"
)

type testCase struct {
	text   []byte
	parser string
	ok     bool
	length int
}

func test(t *testing.T, set *Set, cases []testCase) {
	for _, c := range cases {
		input := NewInput(c.text)
		ok, l, _ := set.Call(c.parser, input, 0)
		if c.ok != ok || c.length != l {
			pt("=== expected ===\n")
			pt("%s %s %v %d\n", c.text, c.parser, c.ok, c.length)
			pt("=== result ===\n")
			pt("%v %d\n", ok, l)
			t.Fatalf("%v", c)
		}
	}
}

func TestAll(t *testing.T) {
	set := NewSet()
	set.Add("a", set.Regex(`a`))
	set.Add("+", set.Regex(`\+`))
	set.Add("expr", set.OrdChoice(set.Concat("expr", "+", "a"), "a"))

	cases := []testCase{
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
	test(t, set, cases)
}

func TestCalc(t *testing.T) {
	/*
		expr = expr (+ | -) term | term
		term = term (* / /) factor | factor
		factor = [0-9]+ | '(' expr ')'
	*/
	set := NewSet()
	set.Add("expr", set.OrdChoice(
		set.Concat("expr", set.Rune('+'), "term"),
		set.Concat("expr", set.Rune('-'), "term"),
		"term",
	))
	set.Add("term", set.OrdChoice(
		set.Concat("term", set.Rune('*'), "factor"),
		set.Concat("term", set.Rune('/'), "factor"),
		"factor",
	))
	set.Add("factor", set.OrdChoice(
		set.Regex(`[0-9]+`),
		set.Concat(set.Rune('('), "expr", set.Rune(')')),
	))

	cases := []testCase{
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
	test(t, set, cases)
}

func TestRegex(t *testing.T) {
	/*
		<RE>	::=	<RE> "|" <simple-RE> | <simple-RE>
		<simple-RE>	::=	<simple-RE> <basic-RE> | <basic-RE>
		<basic-RE>	::=	<elementary-RE> "*" | <elementary-RE> "+" | <elementary-RE>
		<elementary-RE>	::=	"(" <RE> ")" | "." | "$" | [a-zA-Z0-9]
	*/
	set := NewSet()
	set.Add("re", set.OrdChoice(
		set.Concat("re", set.Rune('|'), "simple-re"),
		"simple-re",
	))
	set.Add("simple-re", set.OrdChoice(
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

	cases := []testCase{
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
	test(t, set, cases)
}

func TestIndirect(t *testing.T) {
	/*
		L = P.x | x
		P = P(n) | L
	*/
	set := NewSet()
	set.Add("L", set.OrdChoice(
		set.Concat("P", set.Rune('.'), set.Rune('x')),
		set.Rune('x'),
	))
	set.Add("P", set.OrdChoice(
		set.Concat("P", set.Rune('('), set.Rune('n'), set.Rune(')')),
		"L",
	))

	cases := []testCase{
		{[]byte("x"), "L", true, 1},
		{[]byte("x(n)(n).x(n).x"), "L", true, 14},
	}
	test(t, set, cases)
}

func TestIndirect2(t *testing.T) {
	/*
		A = Ba | d
		B = Cb | e
		C = Ac | f
	*/
	set := NewSet()
	set.Add("A", set.OrdChoice(
		set.Concat("B", set.Rune('a')),
		set.Rune('d'),
	))
	set.Add("B", set.OrdChoice(
		set.Concat("C", set.Rune('b')),
		set.Rune('e'),
	))
	set.Add("C", set.OrdChoice(
		set.Concat("A", set.Rune('c')),
		set.Rune('f'),
	))

	cases := []testCase{
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
	test(t, set, cases)
}

func TestPanic(t *testing.T) {
	set := NewSet()
	func() {
		defer func() {
			if p := recover(); p == nil || p.(string) != "parser not found: foo" {
				t.Fatal("should panic")
			}
		}()
		set.Call("foo", NewInput([]byte("FOO")), 0)
	}()

	func() {
		defer func() {
			if p := recover(); p == nil || p.(string) != "unknown parser type: int" {
				t.Fatal("should panic")
			}
		}()
		set.OrdChoice(42)
	}()

	set.Add("rune", set.Rune('a'))
	func() {
		defer func() {
			if p := recover(); p == nil || p.(string) != "utf8 decode error" {
				t.Fatal("should panic")
			}
		}()
		set.Call("rune", NewInput([]byte("白")[1:]), 0)
	}()
}

func TestByteIn(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.ByteIn([]byte("qwerty")))
	if ok, l, _ := set.Call("foo", NewInput([]byte("a")), 0); ok || l != 0 {
		t.Fatal("fail")
	}
	if ok, l, _ := set.Call("foo", NewInput([]byte("q")), 0); !ok || l != 1 {
		t.Fatal("fail")
	}
	if ok, l, _ := set.Call("foo", NewInput([]byte("qa")), 0); !ok || l != 1 {
		t.Fatal("fail")
	}
}

func TestByteRange(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.ByteRange('a', 'z'))
	if ok, l, _ := set.Call("foo", NewInput([]byte("A")), 0); ok || l != 0 {
		t.Fatal("fail")
	}
	if ok, l, _ := set.Call("foo", NewInput([]byte("a")), 0); !ok || l != 1 {
		t.Fatal("fail")
	}
	if ok, l, _ := set.Call("foo", NewInput([]byte("aA")), 0); !ok || l != 1 {
		t.Fatal("fail")
	}
}

func TestOneOrMore(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.OneOrMore(set.Rune('a')))
	cases := []testCase{
		{[]byte(""), "foo", false, 0},
		{[]byte("b"), "foo", false, 0},
		{[]byte("bb"), "foo", false, 0},
		{[]byte("a"), "foo", true, 1},
		{[]byte("aa"), "foo", true, 2},
		{[]byte("aaa"), "foo", true, 3},
		{[]byte("aaab"), "foo", true, 3},
		{[]byte("aaabb"), "foo", true, 3},
	}
	test(t, set, cases)
}

func TestZeroOrMore(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.ZeroOrMore(set.Rune('a')))
	cases := []testCase{
		{[]byte(""), "foo", true, 0},
		{[]byte("b"), "foo", true, 0},
		{[]byte("bb"), "foo", true, 0},
		{[]byte("a"), "foo", true, 1},
		{[]byte("aa"), "foo", true, 2},
		{[]byte("aaa"), "foo", true, 3},
		{[]byte("aaab"), "foo", true, 3},
		{[]byte("aaabb"), "foo", true, 3},
	}
	test(t, set, cases)
}

func TestDump(t *testing.T) {
	buf := new(bytes.Buffer)
	input := NewInput([]byte("foo"))
	node := &Node{"name", 0, 3, []*Node{
		{"sub1", 0, 1, nil},
		{"sub2", 1, 1, nil},
		{"sub3", 2, 1, nil},
	}}
	node.Dump(buf, input)
	if !bytes.Equal(buf.Bytes(), []byte(`"foo" name 0-3
  "f" sub1 0-1
  "o" sub2 1-2
  "o" sub3 2-3
`)) {
		t.Fatal("not equal")
	}
}

func TestEqual(t *testing.T) {
	node := &Node{"name", 0, 3, []*Node{
		{"sub1", 0, 1, nil},
		{"sub2", 1, 1, nil},
		{"sub3", 2, 1, nil},
	}}
	if node.Equal(&Node{"foo", 0, 3, nil}) {
		t.Fatal("name")
	}
	if node.Equal(&Node{"name", 1, 3, nil}) {
		t.Fatal("start")
	}
	if node.Equal(&Node{"name", 0, 2, nil}) {
		t.Fatal("len")
	}
	if node.Equal(&Node{"name", 0, 3, []*Node{
		{"sub1", 2, 1, nil},
	}}) {
		t.Fatal("sub len")
	}
	if node.Equal(&Node{"name", 0, 3, []*Node{
		{"sub1", 0, 1, nil},
		{"sub2", 1, 1, nil},
		{"sub8", 2, 1, nil},
	}}) {
		t.Fatal("sub")
	}
}
