package core

import (
	"fmt"
	"sync"
	"testing"
)

func prepareZeusro() *Zeusro {
	// TODO
	return nil
}

func TestCoin(t *testing.T) {
	z := prepareZeusro()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(w *sync.WaitGroup) {
			fmt.Println(z.Coin())
			w.Done()
		}(&wg)
	}
	wg.Wait()
}

// 4核运行，总共执行了 106582 次；11184 ns/op，表示每次执行耗时 11184 纳秒
// BenchmarkCoin-4   	  106582	     11184 ns/op	       0 B/op	       0 allocs/op
func BenchmarkCoin(b *testing.B) {
	z := prepareZeusro()
	for i := 0; i < b.N; i++ {
		z.Coin()
	}
}

// BenchmarkCoin-4   	   88270	     14926 ns/op	       0 B/op	       0 allocs/op
func BenchmarkFmtCoin(b *testing.B) {
	z := prepareZeusro()
	for i := 0; i < b.N; i++ {
		fmt.Println(z.Coin())
	}
}

func TestFakeZeusro(t *testing.T) {
	z := prepareZeusro()
	z.Run()
}

func TestZeusro(t *testing.T) {
	z := prepareZeusro()
	z.Run()
}
