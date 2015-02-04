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
		input := NewBytesInput(c.text)
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
	set.Add("a", set.SliceRegex(`a`))
	set.Add("+", set.SliceRegex(`\+`))
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
		set.Concat("expr", set.SliceRune('+'), "term"),
		set.Concat("expr", set.SliceRune('-'), "term"),
		"term",
	))
	set.Add("term", set.OrdChoice(
		set.Concat("term", set.SliceRune('*'), "factor"),
		set.Concat("term", set.SliceRune('/'), "factor"),
		"factor",
	))
	set.Add("factor", set.OrdChoice(
		set.SliceRegex(`[0-9]+`),
		set.Concat(set.SliceRune('('), "expr", set.SliceRune(')')),
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
		set.Concat("re", set.SliceRune('|'), "simple-re"),
		"simple-re",
	))
	set.Add("simple-re", set.OrdChoice(
		set.Concat("simple-re", "basic-re"),
		"basic-re",
	))
	set.Add("basic-re", set.OrdChoice(
		set.Concat("elementary-re", set.SliceRune('*')),
		set.Concat("elementary-re", set.SliceRune('+')),
		"elementary-re",
	))
	set.Add("elementary-re", set.OrdChoice(
		set.Concat(set.SliceRune('('), "re", set.SliceRune(')')),
		set.SliceRune('.'),
		set.SliceRune('$'),
		set.SliceRegex(`[a-zA-Z0-9]`),
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
		set.Concat("P", set.SliceRune('.'), set.SliceRune('x')),
		set.SliceRune('x'),
	))
	set.Add("P", set.OrdChoice(
		set.Concat("P", set.SliceRune('('), set.SliceRune('n'), set.SliceRune(')')),
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
		set.Concat("B", set.SliceRune('a')),
		set.SliceRune('d'),
	))
	set.Add("B", set.OrdChoice(
		set.Concat("C", set.SliceRune('b')),
		set.SliceRune('e'),
	))
	set.Add("C", set.OrdChoice(
		set.Concat("A", set.SliceRune('c')),
		set.SliceRune('f'),
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
		set.Call("foo", NewBytesInput([]byte("FOO")), 0)
	}()

	func() {
		defer func() {
			if p := recover(); p == nil || p.(string) != "unknown parser type: int" {
				t.Fatal("should panic")
			}
		}()
		set.OrdChoice(42)
	}()

	set.Add("rune", set.SliceRune('a'))
	func() {
		defer func() {
			if p := recover(); p == nil || p.(string) != "utf8 decode error" {
				t.Fatal("should panic")
			}
		}()
		set.Call("rune", NewBytesInput([]byte("白")[1:]), 0)
	}()
}

func TestByteIn(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.ByteIn([]byte("qwerty")))
	cases := []testCase{
		{[]byte(""), "foo", false, 0},
		{[]byte("a"), "foo", false, 0},
		{[]byte("q"), "foo", true, 1},
		{[]byte("qa"), "foo", true, 1},
	}
	test(t, set, cases)
}

func TestByteRange(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.ByteRange('a', 'z'))
	cases := []testCase{
		{[]byte(""), "foo", false, 0},
		{[]byte("A"), "foo", false, 0},
		{[]byte("a"), "foo", true, 1},
		{[]byte("aA"), "foo", true, 1},
	}
	test(t, set, cases)
}

func TestOneOrMore(t *testing.T) {
	set := NewSet()
	set.Add("foo", set.OneOrMore(set.SliceRune('a')))
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
	set.Add("foo", set.ZeroOrMore(set.SliceRune('a')))
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
	input := NewBytesInput([]byte("foo"))
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

func TestRepeat(t *testing.T) {
	set := NewSet()
	set.Add("0+", set.Repeat(0, -1, set.SliceRune('a')))
	set.Add("0-1", set.Repeat(0, 1, set.SliceRune('a')))
	set.Add("0-2", set.Repeat(0, 2, set.SliceRune('a')))
	set.Add("1+", set.Repeat(1, -1, set.SliceRune('a')))
	set.Add("1-1", set.Repeat(1, 1, set.SliceRune('a')))
	set.Add("1-2", set.Repeat(1, 2, set.SliceRune('a')))
	set.Add("2+", set.Repeat(2, -1, set.SliceRune('a')))
	set.Add("2-2", set.Repeat(2, 2, set.SliceRune('a')))
	set.Add("2-3", set.Repeat(2, 3, set.SliceRune('a')))
	cases := []testCase{
		{[]byte(""), "0+", true, 0},
		{[]byte("a"), "0+", true, 1},
		{[]byte("aa"), "0+", true, 2},
		{[]byte("b"), "0+", true, 0},
		{[]byte("bb"), "0+", true, 0},

		{[]byte(""), "0-1", true, 0},
		{[]byte("a"), "0-1", true, 1},
		{[]byte("aa"), "0-1", true, 1},
		{[]byte("ab"), "0-1", true, 1},

		{[]byte(""), "0-2", true, 0},
		{[]byte("a"), "0-2", true, 1},
		{[]byte("aa"), "0-2", true, 2},
		{[]byte("aaa"), "0-2", true, 2},
		{[]byte("baa"), "0-2", true, 0},

		{[]byte(""), "1+", false, 0},
		{[]byte("b"), "1+", false, 0},
		{[]byte("a"), "1+", true, 1},
		{[]byte("aa"), "1+", true, 2},
		{[]byte("aaa"), "1+", true, 3},

		{[]byte(""), "1-1", false, 0},
		{[]byte("b"), "1-1", false, 0},
		{[]byte("a"), "1-1", true, 1},
		{[]byte("aa"), "1-1", true, 1},
		{[]byte("ab"), "1-1", true, 1},

		{[]byte(""), "1-2", false, 0},
		{[]byte("b"), "1-2", false, 0},
		{[]byte("a"), "1-2", true, 1},
		{[]byte("aa"), "1-2", true, 2},
		{[]byte("aaa"), "1-2", true, 2},

		{[]byte(""), "2+", false, 0},
		{[]byte("b"), "2+", false, 0},
		{[]byte("a"), "2+", false, 0},
		{[]byte("aa"), "2+", true, 2},
		{[]byte("aaa"), "2+", true, 3},

		{[]byte(""), "2-2", false, 0},
		{[]byte("b"), "2-2", false, 0},
		{[]byte("a"), "2-2", false, 0},
		{[]byte("aa"), "2-2", true, 2},
		{[]byte("aaa"), "2-2", true, 2},

		{[]byte(""), "2-3", false, 0},
		{[]byte("b"), "2-3", false, 0},
		{[]byte("a"), "2-3", false, 0},
		{[]byte("aa"), "2-3", true, 2},
		{[]byte("aaa"), "2-3", true, 3},
		{[]byte("aaaa"), "2-3", true, 3},
	}
	test(t, set, cases)
}

func TestPredicate(t *testing.T) {
	set := NewSet()
	set.Add("p", set.Concat(
		set.Predicate(set.SliceRune('a')),
		set.SliceRegex(`.*bar`)))
	set.Add("np", set.Concat(
		set.NotPredicate(set.SliceRune('a')),
		set.SliceRegex(`.*bar`)))
	cases := []testCase{
		{[]byte("foobar"), "p", false, 0},
		{[]byte("afoobar"), "p", true, 7},
		{[]byte("aafoobar"), "p", true, 8},

		{[]byte("afoobar"), "np", false, 0},
		{[]byte("aafoobar"), "np", false, 0},
		{[]byte("foobar"), "np", true, 6},
		{[]byte("bazfoobar"), "np", true, 9},
	}
	test(t, set, cases)
}
