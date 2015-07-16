// Basic benchmark
// go test -test.run="^bench$" -bench=.
//
// BenchmarkCall	  500000	        2890 ns/op
// BenchmarkPlain	100000000	        13.6 ns/op
package breaker

import (
	"testing"
)

var (
	breaker *Breaker
)

func init() {
	breaker, _ = NewBreaker(OptionsDefaults())
}

func ftest() error {
	for index := 0; index < 10; index++ {
		_ = "nothing"
	}
	return nil
}

func benchmarkCallN(b *testing.B) {
	for n := 0; n < b.N; n++ {
		breaker.Call(ftest)
	}
}

func benchmarkPlainN(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ftest()
	}
}

func BenchmarkCall(b *testing.B)  { benchmarkCallN(b) }
func BenchmarkPlain(b *testing.B) { benchmarkPlainN(b) }
