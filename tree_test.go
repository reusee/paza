package paza

import "testing"

type treeTestCase struct {
	Text string
	Name string
	Node *Node
}

func testTree(t *testing.T, set *Set, cases []treeTestCase) {
	for _, c := range cases {
		input := NewInput([]byte(c.Text))
		ok, l, node := set.Call(c.Name, input, 0)
		if !ok || l != len(c.Text) {
			t.Fatalf("match fail: %v %d %v", ok, l, c)
		}
		if !node.Equal(c.Node) {
			pt("== expected ==\n")
			c.Node.Dump(input)
			pt("== return ==\n")
			node.Dump(input)
			t.Fatalf("tree not match")
		}
	}
}

func TestParseTree(t *testing.T) {
	/*
		expr = expr (+ | -) term | term
		term = term (* / /) factor | factor
		factor = [0-9]+ | '(' expr ')'
	*/
	set := NewSet()
	set.AddRec("expr", set.OrdChoice(
		set.NamedConcat("plus-expr", "expr", set.NamedRune("plus-op", '+'), "term"),
		set.NamedConcat("minus-expr", "expr", set.NamedRune("minus-op", '-'), "term"),
		"term",
	))
	set.AddRec("term", set.OrdChoice(
		set.NamedConcat("mul-expr", "term", set.NamedRune("mul-op", '*'), "factor"),
		set.NamedConcat("div-expr", "term", set.NamedRune("div-op", '/'), "factor"),
		"factor",
	))
	set.Add("factor", set.OrdChoice(
		set.NamedRegex("digit", `[0-9]+`),
		set.NamedConcat("quoted", set.NamedRune("left-quote", '('), "expr", set.NamedRune("right-quote", ')')),
	))

	cases := []treeTestCase{
		{"1", "expr", &Node{"expr", 0, 1, []*Node{
			{"term", 0, 1, []*Node{
				{"factor", 0, 1, []*Node{
					{"digit", 0, 1, nil},
				}}}}}}},
		{"1+2", "expr", &Node{"expr", 0, 3, []*Node{
			{"plus-expr", 0, 3, []*Node{
				{"expr", 0, 1, []*Node{
					{"term", 0, 1, []*Node{
						{"factor", 0, 1, []*Node{
							{"digit", 0, 1, nil}}}}}}},
				{"plus-op", 1, 1, nil},
				{"term", 2, 1, []*Node{
					{"factor", 2, 1, []*Node{
						{"digit", 2, 1, nil}}}}}}}}}},
		{"1-2", "expr", &Node{"expr", 0, 3, []*Node{
			{"minus-expr", 0, 3, []*Node{
				{"expr", 0, 1, []*Node{
					{"term", 0, 1, []*Node{
						{"factor", 0, 1, []*Node{
							{"digit", 0, 1, nil}}}}}}},
				{"minus-op", 1, 1, nil},
				{"term", 2, 1, []*Node{
					{"factor", 2, 1, []*Node{
						{"digit", 2, 1, nil}}}}}}}}}},
		{"1*2", "expr", &Node{"expr", 0, 3, []*Node{
			{"term", 0, 3, []*Node{
				{"mul-expr", 0, 3, []*Node{
					{"term", 0, 1, []*Node{
						{"factor", 0, 1, []*Node{
							{"digit", 0, 1, nil}}}}},
					{"mul-op", 1, 1, nil},
					{"factor", 2, 1, []*Node{
						{"digit", 2, 1, nil}}}}}}}}}},
		{"1/2", "expr", &Node{"expr", 0, 3, []*Node{
			{"term", 0, 3, []*Node{
				{"div-expr", 0, 3, []*Node{
					{"term", 0, 1, []*Node{
						{"factor", 0, 1, []*Node{
							{"digit", 0, 1, nil}}}}},
					{"div-op", 1, 1, nil},
					{"factor", 2, 1, []*Node{
						{"digit", 2, 1, nil}}}}}}}}}},
		{"(1)", "expr", &Node{"expr", 0, 3, []*Node{
			{"term", 0, 3, []*Node{
				{"factor", 0, 3, []*Node{
					{"quoted", 0, 3, []*Node{
						{"left-quote", 0, 1, nil},
						{"expr", 1, 1, []*Node{
							{"term", 1, 1, []*Node{
								{"factor", 1, 1, []*Node{
									{"digit", 1, 1, nil}}}}}}},
						{"right-quote", 2, 1, nil}}}}}}}}}},
	}
	testTree(t, set, cases)
}
