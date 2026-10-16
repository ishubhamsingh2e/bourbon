package app

import (
	"errors"
	"sync"
	"testing"
)

type Database interface {
	Query() string
}

type MockDB struct {
	dsn string
}

func (m *MockDB) Query() string {
	return "mock:" + m.dsn
}

type UserService struct {
	DB Database
}

type CircularA struct {
	B *CircularB
}

type CircularB struct {
	A *CircularA
}

func TestContainerSingleton(t *testing.T) {
	c := NewContainer()
	db := &MockDB{dsn: "postgres://localhost"}
	c.Singleton(db)

	var resolved *MockDB
	if err := c.Resolve(&resolved); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if resolved != db {
		t.Errorf("resolved pointer %p != original %p", resolved, db)
	}
}

func TestContainerBindConstructor(t *testing.T) {
	c := NewContainer()
	c.Singleton(&MockDB{dsn: "memory"})
	c.Bind(func(db *MockDB) *UserService {
		return &UserService{DB: db}
	})

	var svc1 *UserService
	var svc2 *UserService

	if err := c.Resolve(&svc1); err != nil {
		t.Fatalf("Resolve svc1 failed: %v", err)
	}
	if err := c.Resolve(&svc2); err != nil {
		t.Fatalf("Resolve svc2 failed: %v", err)
	}

	if svc1 == svc2 {
		t.Errorf("transient bindings should produce distinct instances")
	}
	if svc1.DB.Query() != "mock:memory" {
		t.Errorf("unexpected DB query output: %v", svc1.DB.Query())
	}
}

func TestContainerConstructorError(t *testing.T) {
	c := NewContainer()
	expectedErr := errors.New("cannot connect")
	c.Bind(func() (*MockDB, error) {
		return nil, expectedErr
	})

	var db *MockDB
	err := c.Resolve(&db)
	if err == nil {
		t.Fatalf("expected error from constructor, got nil")
	}
}

func TestContainerCircularDependency(t *testing.T) {
	c := NewContainer()
	c.Bind(func(b *CircularB) *CircularA {
		return &CircularA{B: b}
	})
	c.Bind(func(a *CircularA) *CircularB {
		return &CircularB{A: a}
	})

	var a *CircularA
	err := c.Resolve(&a)
	if err == nil {
		t.Fatalf("expected circular dependency error, got nil")
	}
	if !errors.Is(err, ErrCircularDependency) {
		t.Errorf("expected ErrCircularDependency, got: %v", err)
	}
}

func TestContainerNonPointerTarget(t *testing.T) {
	c := NewContainer()
	c.Singleton(&MockDB{dsn: "test"})

	var db MockDB
	err := c.Resolve(db)
	if !errors.Is(err, ErrTargetNotPointer) {
		t.Errorf("expected ErrTargetNotPointer, got %v", err)
	}
}

func TestContainerConcurrentResolve(t *testing.T) {
	c := NewContainer()
	c.Singleton(&MockDB{dsn: "concurrent"})
	c.Bind(func(db *MockDB) *UserService {
		return &UserService{DB: db}
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var svc *UserService
			if err := c.Resolve(&svc); err != nil {
				t.Errorf("concurrent Resolve failed: %v", err)
			}
		}()
	}
	wg.Wait()
}
