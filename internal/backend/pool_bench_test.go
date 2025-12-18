package backend

import (
	"fmt"
	"testing"
)

func BenchmarkPool_AddBackend(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p := NewPool()
		for i := 0; i < 1000; i++ {
			_, _ = p.AddBackend(fmt.Sprintf("10.0.0.%d:8080", i), 1)
		}
	}
}

func BenchmarkPool_GetBackends(b *testing.B) {
	p := NewPool()
	for i := 0; i < 1000; i++ {
		_, _ = p.AddBackend(fmt.Sprintf("10.0.0.%d:8080", i), 1)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p.GetBackends()
	}
}

func BenchmarkPool_AliveSnapshot(b *testing.B) {
	p := NewPool()
	for i := 0; i < 1000; i++ {
		_, _ = p.AddBackend(fmt.Sprintf("10.0.0.%d:8080", i), 1)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		p.AliveSnapshot()
	}
}
