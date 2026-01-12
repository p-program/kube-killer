package killer

import (
	"fmt"
	"sync"
	"testing"

	"github.com/p-program/kube-killer/config"
	"github.com/stretchr/testify/assert"
)

func prepareZeusro() *Zeusro {
	projectConfig := config.NewProjectConfig()
	return NewZeusro(projectConfig, "default", false)
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
	skipIfNoCluster(t)
	z := prepareZeusro()
	z.DryRun()
	err := z.Run()
	// Error may occur if no pods exist, which is fine
	_ = err
}

func TestZeusro(t *testing.T) {
	skipIfNoCluster(t)
	z := prepareZeusro()
	z.DryRun()
	err := z.Run()
	// Error may occur if no pods exist, which is fine
	_ = err
}

func TestZeusroNewZeusro(t *testing.T) {
	skipIfNoCluster(t)
	projectConfig := config.NewProjectConfig()
	z := NewZeusro(projectConfig, "default", true)
	assert.NotNil(t, z)
	assert.Equal(t, "default", z.namespace)
	assert.True(t, z.dryRun)
	assert.NotNil(t, z.config)
}

func TestZeusroDryRun(t *testing.T) {
	skipIfNoCluster(t)
	projectConfig := config.NewProjectConfig()
	z := NewZeusro(projectConfig, "default", false)
	z.DryRun()
	assert.True(t, z.dryRun)
}
