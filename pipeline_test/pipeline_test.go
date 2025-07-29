// =====================
// pipeline_test.go
// =====================
package pipeline_test_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/TheOrchestraX/pipeline"
)

func TestPipeline_Success(t *testing.T) {
	p := pipeline.New[int]().
		Then(pipeline.Wrap(func(x int) int { return x + 1 })).
		Then(pipeline.Wrap(func(x int) int { return x * 2 }))

	out, err := p.Execute(3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if out != 8 {
		t.Errorf("Expected 8, got %d", out)
	}
}

func TestPipeline_Error(t *testing.T) {
	errFail := errors.New("failure")
	p := pipeline.New[int]().
		Then(func(x int) (int, error) { return x + 1, nil }).
		Then(func(x int) (int, error) { return x, errFail }).
		Then(func(x int) (int, error) { return x * 2, nil })

	out, err := p.Execute(3)
	if err != errFail {
		t.Fatalf("Expected error %v, got %v", errFail, err)
	}
	if out != 4 {
		t.Errorf("Expected out to be 4, got %d", out)
	}
}

func TestPipeline_Middleware(t *testing.T) {
	var logs []string
	mw := func(next pipeline.StepFunc[int]) pipeline.StepFunc[int] {
		return func(x int) (int, error) {
			logs = append(logs, fmt.Sprintf("before %d", x))
			res, err := next(x)
			logs = append(logs, fmt.Sprintf("after %d", res))
			return res, err
		}
	}
	p := pipeline.New[int]().Use(mw).
		Then(pipeline.Wrap(func(x int) int { return x + 2 }))
	out, err := p.Execute(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if out != 3 {
		t.Errorf("Expected 3, got %d", out)
	}
	expected := []string{"before 1", "after 3"}
	if len(logs) != len(expected) {
		t.Fatalf("Expected logs %v, got %v", expected, logs)
	}
}

func TestPipeline_Conditional(t *testing.T) {
	inc := pipeline.Wrap(func(x int) int { return x + 1 })
	dec := pipeline.Wrap(func(x int) int { return x - 1 })
	cond := pipeline.Conditional(func(x int) bool { return x%2 == 0 }, inc, dec)
	p := pipeline.New[int]().Then(cond)

	even, _ := p.Execute(4)
	odd, _ := p.Execute(3)
	if even != 5 || odd != 2 {
		t.Errorf("Conditional failed, got %d and %d", even, odd)
	}
}

func TestPipeline_Parallel(t *testing.T) {
	f1 := pipeline.Wrap(func(x int) int { return x + 1 })
	f2 := pipeline.Wrap(func(x int) int { return x * 2 })
	combiner := func(results []int) (int, error) {
		sum := 0
		for _, v := range results {
			sum += v
		}
		return sum, nil
	}
	p := pipeline.New[int]().Then(pipeline.Parallel(combiner, f1, f2))
	out, err := p.Execute(3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// f1: 4, f2: 6, sum = 10
	if out != 10 {
		t.Errorf("Expected 10, got %d", out)
	}
}
