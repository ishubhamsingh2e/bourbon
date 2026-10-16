// Package cron implements schedule.Scheduler using standard library time.Ticker.
package cron

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/schedule"
)

type taskEntry struct {
	id       string
	interval time.Duration
	cronExpr string
	fn       schedule.TaskFunc
}

// Scheduler manages recurring tasks and interval runners.
type Scheduler struct {
	mu     sync.RWMutex
	tasks  map[string]*taskEntry
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New constructs an empty recurring task Scheduler.
func New() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*taskEntry),
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Every registers a task to run periodically at the specified duration interval.
func (s *Scheduler) Every(interval time.Duration, task schedule.TaskFunc) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	s.tasks[id] = &taskEntry{
		id:       id,
		interval: interval,
		fn:       task,
	}
	return id
}

// Cron registers a task using standard 5-field cron syntax (min, hour, dom, month, dow).
func (s *Scheduler) Cron(expression string, task schedule.TaskFunc) (string, error) {
	fields := strings.Fields(expression)
	if len(fields) != 5 {
		return "", fmt.Errorf("cron: invalid expression '%s', expected 5 fields", expression)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	s.tasks[id] = &taskEntry{
		id:       id,
		cronExpr: expression,
		fn:       task,
	}
	return id, nil
}

// Cancel removes a task by ID.
func (s *Scheduler) Cancel(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, taskID)
}

// Start initiates the scheduling background loop.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	tasks := make([]*taskEntry, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	s.mu.Unlock()

	for _, t := range tasks {
		s.wg.Add(1)
		go s.runTask(t)
	}
}

func (s *Scheduler) runTask(t *taskEntry) {
	defer s.wg.Done()

	if t.interval > 0 {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				_ = t.fn(s.ctx)
			}
		}
	}

	// For cron expressions, poll every minute on the minute boundary
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastRunMinute int = -1

	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-ticker.C:
			if now.Second() == 0 && now.Minute() != lastRunMinute {
				if matchCron(t.cronExpr, now) {
					lastRunMinute = now.Minute()
					_ = t.fn(s.ctx)
				}
			}
		}
	}
}

func matchCron(expr string, t time.Time) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}

	minMatch := matchField(fields[0], t.Minute())
	hourMatch := matchField(fields[1], t.Hour())
	domMatch := matchField(fields[2], t.Day())
	monthMatch := matchField(fields[3], int(t.Month()))
	dowMatch := matchField(fields[4], int(t.Weekday()))

	return minMatch && hourMatch && domMatch && monthMatch && dowMatch
}

func matchField(field string, currentVal int) bool {
	if field == "*" {
		return true
	}
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err == nil && step > 0 {
			return currentVal%step == 0
		}
	}
	val, err := strconv.Atoi(field)
	if err == nil {
		return val == currentVal
	}
	return false
}

// Stop terminates all scheduled task goroutines.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	s.wg.Wait()
}
