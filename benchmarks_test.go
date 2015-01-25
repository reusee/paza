package paza

import "testing"

func BenchmarkExpr(b *testing.B) {
	set := NewSet()
	set.AddRec("expr", set.OrdChoice(
		set.Concat(
			"expr",
			set.Regex(`[\+\-\*/]`),
			set.Regex(`[a-z]+`),
		),
		set.Regex(`[a-z]+`),
	))
	input := NewInput([]byte("foo+bar-baz*qux/quux"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok, _ := set.Call("expr", input, 0)
		if !ok {
			b.Fatal("fail")
		}
	}
}
