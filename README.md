# Pipeline Module

A lightweight, generic Go pipeline library with support for:

- **Type-safe pipelined step functions** (`StepFunc[T]`).
- **Middleware-style interceptors** (`Use`).
- **Conditional routing** (`Conditional`).
- **Parallel branch execution** (`Parallel`).
- **Error short-circuiting**.

---

## Installation

```bash
go get github.com/TheOrchestraX/pipeline@latest
```

## Quick Start
```go
import (
  "fmt"
  "github.com/TheOrchestraX/pipeline"
)

func main() {
  // Create a new pipeline for int values
  p := pipeline.New[int]().
    Use(func(next pipeline.StepFunc[int]) pipeline.StepFunc[int] {
      return func(x int) (int, error) {
        fmt.Println("Input:", x)
        out, err := next(x)
        fmt.Println("Output:", out)
        return out, err
      }
    }).
    Then(pipeline.Wrap(func(x int) int { return x + 5 })).
    Then(func(x int) (int, error) {
      if x%2 != 0 {
        return x, fmt.Errorf("odd result: %d", x)
      }
      return x * 2, nil
    })

  result, err := p.Execute(3)
  if err != nil {
    panic(err)
  }
  fmt.Println("Final result:", result)
}
```

### API Reference

Creates a new, empty pipeline for type T.
```go
func New[T any]() *Pipeline[T]
```

Registers a middleware interceptor that wraps all subsequently added steps.
```go
func (p *Pipeline[T]) Use(middleware Middleware[T]) *Pipeline[T]
``` 

Appends a step to the pipeline. Steps run in the order they’re added, after middleware wrapping.
```go
func (p *Pipeline[T]) Then(step StepFunc[T]) *Pipeline[T]
``` 

Executes the pipeline on the given input. Returns the final output or the first error encountered.
```go
func (p *Pipeline[T]) Execute(input T) (T, error)
``` 

Converts a pure function f(T) T into a StepFunc[T] that never errors.
```go
func Wrap[T any](f func(T) T) StepFunc[T]
``` 

Creates a step that chooses between thenStep and elseStep based on the boolean result of predicate.
```go
func Conditional[T any](predicate func(T) bool, thenStep, elseStep StepFunc[T]) StepFunc[T]
``` 

Runs multiple steps in parallel on the same input, then calls combiner on the results. If any step errors, the first error is returned.
```go
func Parallel[T any](combiner func([]T) (T, error), steps ...StepFunc[T]) StepFunc[T]
``` 

## Examples

####  Conditional routing:
```go
inc := pipeline.Wrap(func(x int) int { return x + 1 })
dec := pipeline.Wrap(func(x int) int { return x - 1 })
condStep := pipeline.Conditional(func(x int) bool { return x > 0 }, inc, dec)
result, _ := pipeline.New[int]().Then(condStep).Execute(-5)
// result == -6
``` 


#### Parallel execution:
```go
f1 := pipeline.Wrap(func(x int) int { return x + 1 })
f2 := pipeline.Wrap(func(x int) int { return x * 2 })
sumCombiner := func(results []int) (int, error) {
  sum := 0
  for _, r := range results {
    sum += r
  }
  return sum, nil
}
result, _ := pipeline.New[int]().Then(pipeline.Parallel(sumCombiner, f1, f2)).Execute(10)
// f1: 11, f2: 20, sum = 31
```

## Future Work

Adding Javascript bindings for this Go library is planned, allowing users to leverage the pipeline functionality in those languages.

Using GOJA in a pipeline
```go
import (
  "github.com/dop251/goja"
  "github.com/TheOrchestraX/pipeline"
)

func JSStep(code string) pipeline.StepFunc[int] {
  return func(input int) (int, error) {
    vm := goja.New()
    vm.Set("input", input)
    val, err := vm.RunString(code)       // similar API to Otto
    if err != nil {
      return input, err
    }
    num, err := val.ToInteger()
    return int(num), err
  }
}
```

