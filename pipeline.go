package pipeline

import (
	"sync"
)

// StepFunc is a pipeline step that transforms an input of type T, optionally returning an error.
type StepFunc[T any] func(T) (T, error)

// Middleware is a function that wraps a StepFunc to provide cross-cutting behavior.
type Middleware[T any] func(next StepFunc[T]) StepFunc[T]

// Pipeline chains a series of StepFuncs to process data in sequence.
type Pipeline[T any] struct {
	steps       []StepFunc[T]
	middlewares []Middleware[T]
}

// New creates a new, empty Pipeline for type T.
func New[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		steps:       make([]StepFunc[T], 0),
		middlewares: make([]Middleware[T], 0),
	}
}

// Use appends a Middleware to be applied to all subsequent steps.
func (p *Pipeline[T]) Use(mw Middleware[T]) *Pipeline[T] {
	p.middlewares = append(p.middlewares, mw)
	return p
}

// Then appends a StepFunc to the pipeline, applying any registered Middleware.
func (p *Pipeline[T]) Then(step StepFunc[T]) *Pipeline[T] {
	// Apply middlewares in reverse registration order
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		step = p.middlewares[i](step)
	}
	p.steps = append(p.steps, step)
	return p
}

// Execute runs the pipeline on the given input, passing the output of each step to the next.
// If any step returns an error, execution stops and that error is returned.
func (p *Pipeline[T]) Execute(input T) (T, error) {
	curr := input
	var err error
	for _, s := range p.steps {
		curr, err = s(curr)
		if err != nil {
			return curr, err
		}
	}
	return curr, nil
}

// Wrap converts a pure function f(T) T into a StepFunc[T], capturing no errors.
func Wrap[T any](f func(T) T) StepFunc[T] {
	return func(input T) (T, error) {
		return f(input), nil
	}
}

// Conditional creates a StepFunc that chooses between thenStep and elseStep based on predicate.
func Conditional[T any](predicate func(T) bool, thenStep, elseStep StepFunc[T]) StepFunc[T] {
	return func(input T) (T, error) {
		if predicate(input) {
			return thenStep(input)
		}
		return elseStep(input)
	}
}

// Parallel runs multiple StepFuncs on the same input concurrently, then combines their outputs.
func Parallel[T any](combiner func([]T) (T, error), steps ...StepFunc[T]) StepFunc[T] {
	return func(input T) (T, error) {
		var (
			wg      sync.WaitGroup
			results = make([]T, len(steps))
			errs    = make([]error, len(steps))
		)
		wg.Add(len(steps))
		for i, step := range steps {
			go func(idx int, s StepFunc[T]) {
				defer wg.Done()
				results[idx], errs[idx] = s(input)
			}(i, step)
		}
		wg.Wait()
		// Return first error if any
		for _, err := range errs {
			if err != nil {
				return results[0], err
			}
		}
		// Combine results
		return combiner(results)
	}
}
