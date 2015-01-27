package main

import (
	"os"
	"runtime/pprof"

	paza "../"
)

func main() {
	set := paza.NewSet()
	set.AddRec("expr", set.OrdChoice(
		set.Concat(
			"expr",
			set.ByteIn([]byte("+-*/")),
			set.OneOrMore(set.ByteRange('a', 'z')),
		),
		set.OneOrMore(set.ByteRange('a', 'z')),
	))
	n := 100000
	f, err := os.Create("profile")
	if err != nil {
		panic(err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}
	for i := 0; i < n; i++ {
		input := paza.NewInput([]byte("foo+bar-baz*qux/quux"))
		ok, l := set.Call("expr", input, 0)
		if !ok || l != 20 {
			panic("fixme")
		}
	}
	pprof.StopCPUProfile()
}
