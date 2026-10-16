package benchmarks

import (
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
)

type BenchmarkService struct {
	Name string
}

func BenchmarkContainerSingletonResolution(b *testing.B) {
	c := app.NewContainer()
	c.Singleton(func() (*BenchmarkService, error) {
		return &BenchmarkService{Name: "singleton"}, nil
	})

	b.ReportAllocs()

	for b.Loop() {
		var svc *BenchmarkService
		if err := c.Resolve(&svc); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContainerFactoryResolution(b *testing.B) {
	c := app.NewContainer()
	c.Bind(func() (*BenchmarkService, error) {
		return &BenchmarkService{Name: "transient"}, nil
	})

	b.ReportAllocs()

	for b.Loop() {
		var svc *BenchmarkService
		if err := c.Resolve(&svc); err != nil {
			b.Fatal(err)
		}
	}
}
