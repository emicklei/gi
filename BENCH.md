```
func main() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if i == j {
				print("a")
			} else if i < j {
				print("b")
			} else {
				print("c")
			}
		}
	}
}
```
23-10-20025
```
BenchmarkIfElseIfElse/native-12                 1000000000               0.0000491 ns/op               0 B/op          0 allocs/op
BenchmarkIfElseIfElse/run-12                          45          25560395 ns/op         2582401 B/op      75783 allocs/op
BenchmarkIfElseIfElse/walk-12                          1        2714144625 ns/op         5869624 B/op     130140 allocs/op
```
24-10-2025
```
BenchmarkIfElseIfElse/native-12                 1000000000               0.0000535 ns/op               0 B/op          0 allocs/op
BenchmarkIfElseIfElse/run-12                         204           5726964 ns/op         2571951 B/op      75580 allocs/op
BenchmarkIfElseIfElse/walk-12                        207           5695296 ns/op         2070261 B/op      70127 allocs/op
```