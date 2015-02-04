package paza

import "testing"

func BenchmarkRecursive(b *testing.B) {
	set := NewSet()
	set.Add("expr", set.OrdChoice(
		set.Concat(
			"expr",
			set.SliceRegex(`[\+\-\*/]`),
			set.SliceRegex(`[a-z]+`),
		),
		set.SliceRegex(`[a-z]+`),
	))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := NewBytesInput([]byte("foo+bar-baz*qux/quux"))
		ok, _, _ := set.Call("expr", input, 0)
		if !ok {
			b.Fatal("fail")
		}
	}
}

func BenchmarkNonRecursive(b *testing.B) {
	set := NewSet()
	set.Add("foo", set.SliceRegex(`foo`))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := NewBytesInput([]byte(`foofoofoo`))
		ok, _, _ := set.Call("foo", input, 0)
		if !ok {
			b.Fatal("fail")
		}
	}
}
