# System Prompt Additions for Code Quality

## Code Quality Standards

NEVER write production code that contains:

1. **panic() statements in normal operation paths** - always return `(T, error)`
2. **resource leaks** - every resource allocation must have corresponding cleanup
3. **data corruption potential** - all state transitions must preserve data integrity
4. **inconsistent error handling patterns** - establish and follow single pattern
5. **unhandled errors** - every error must be checked and handled appropriately

ALWAYS:

1. **Write comprehensive tests BEFORE implementing features** (TDD red-green-refactor)
2. **Include invariant validation in data structures**
3. **Use proper bounds checking for numeric conversions**
4. **Document known bugs immediately and fix them before continuing**
5. **Implement proper separation of concerns**
6. **Use static analysis tools (go vet, staticcheck, golangci-lint) before considering code complete**
7. **Run race detector for concurrent code** (`go test -race`)

## Development Process Guards

### TESTING REQUIREMENTS:
- Write failing tests first, then implement to make them pass (TDD)
- Never commit code with TODO comments for bugs - fix the bugs
- Include table-driven tests for comprehensive scenario coverage
- Test goroutine safety with race detector (`-race` flag)
- Validate all edge cases and boundary conditions
- Aim for ≥90% test coverage for critical code paths

### ARCHITECTURE REQUIREMENTS:
- Explicit error handling - no silent failures or ignored errors
- Resource safety - all cleanup handled via `defer` or proper lifecycle
- Performance conscious - avoid unnecessary allocations
- API design - consistent patterns across all public interfaces
- Goroutine safety - document and test concurrent access patterns

### REVIEW CHECKPOINTS:

Before marking any code complete, verify:

1. **No compilation warnings**
2. **All tests pass (including race detector: `go test -race`)**
3. **Memory/goroutine leaks checked** (use pprof for long-running code)
4. **No data corruption potential in any code path**
5. **Error handling is comprehensive and consistent**
6. **Code is modular and maintainable**
7. **Documentation matches implementation** (godoc comments)
8. **Static analysis passes** (`go vet`, `staticcheck`, `golangci-lint`)
9. **Test coverage meets target** (≥90% for critical paths)

## Go-Specific Quality Standards

### ERROR HANDLING:
- Use `(T, error)` return pattern for all fallible operations
- Define custom error types with context using `errors.New()` or `fmt.Errorf()`
- Never ignore errors - always check `if err != nil`
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Provide meaningful error messages with sufficient context
- Use sentinel errors (`var ErrNotFound = errors.New(...)`) for expected errors

### RESOURCE MANAGEMENT:
- Use `defer` for cleanup immediately after resource acquisition
- Close all resources (files, connections, channels) explicitly
- Check and handle errors from cleanup operations
- Use context.Context for cancellation and timeouts
- Test for resource leaks in long-running scenarios

### DATA STRUCTURE INVARIANTS:
- Document all invariants in godoc comments
- Implement runtime validation in constructors
- Test invariant preservation across all operations
- Use unexported fields with exported accessor methods to enforce invariants
- Validate state consistency at package boundaries

### PACKAGE ORGANIZATION:
- Single responsibility per package
- Clear public/private API boundaries (exported vs unexported)
- Comprehensive package documentation
- Logical dependency hierarchy (no circular dependencies)
- Internal packages for implementation details

### CONCURRENCY SAFETY:
- Document whether types are goroutine-safe
- Use `sync.Mutex` or `sync.RWMutex` for shared state
- Prefer channels for communication over shared memory
- Always test concurrent code with race detector
- Use `sync.WaitGroup` for goroutine coordination
- Avoid goroutine leaks - ensure all goroutines terminate

## Critical Patterns to Avoid

### DANGEROUS PATTERNS:
```go
// NEVER DO THIS - production panic
panic("this should never happen")

// NEVER DO THIS - unchecked conversion (can overflow)
id := uint32(size) // Can overflow on 64-bit

// NEVER DO THIS - ignoring errors
SomeOperation() // Error ignored

// NEVER DO THIS - leaking resources
file, _ := os.Open("file.txt")
// ... no corresponding file.Close()

// NEVER DO THIS - goroutine leak
go func() {
    for {
        // infinite loop with no exit condition
    }
}()

// NEVER DO THIS - race condition
var counter int
go func() { counter++ }()
go func() { counter++ }()
```

### PREFERRED PATTERNS:
```go
// DO THIS - proper error handling
func Operation() (Result, error) {
    result, err := RiskyOperation()
    if err != nil {
        return Result{}, fmt.Errorf("operation failed: %w", err)
    }
    return process(result), nil
}

// DO THIS - safe conversion with validation
func SafeConvert(size int64) (uint32, error) {
    if size < 0 || size > math.MaxUint32 {
        return 0, fmt.Errorf("size %d out of uint32 range", size)
    }
    return uint32(size), nil
}

// DO THIS - explicit error handling
result, err := SomeOperation()
if err != nil {
    return fmt.Errorf("some operation failed: %w", err)
}

// DO THIS - proper resource management with defer
func ProcessFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close() // Cleanup immediately after acquisition

    // Process file
    return nil
}

// DO THIS - goroutine with proper lifecycle
func Worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case job := <-jobs:
            process(job)
        case <-ctx.Done():
            return // Proper exit on cancellation
        }
    }
}

// DO THIS - goroutine-safe counter
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

## Testing Standards

### COMPREHENSIVE TEST COVERAGE:
- Unit tests for all public functions
- Integration tests for complex interactions
- Table-driven tests for multiple scenarios
- Concurrent tests with race detector
- Edge case and boundary condition tests
- Test coverage ≥90% for critical code paths

### TEST ORGANIZATION:
```go
package mypackage_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "myproject/mypackage"
)

func TestNormalOperation(t *testing.T) {
    // Arrange
    input := setupTestData()

    // Act
    result, err := mypackage.Operation(input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}

func TestEdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Output
        wantErr  bool
    }{
        {
            name:     "empty input",
            input:    Input{},
            expected: Output{},
            wantErr:  true,
        },
        {
            name:     "maximum value",
            input:    Input{Value: math.MaxInt64},
            expected: Output{Result: math.MaxInt64},
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := mypackage.Operation(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestErrorConditions(t *testing.T) {
    // Test all error paths
    _, err := mypackage.Operation(invalidInput)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
}

func TestInvariantsPreserved(t *testing.T) {
    // Verify data structure invariants
    ds := mypackage.NewDataStructure()

    // Perform operations
    ds.Insert(1)
    ds.Insert(2)
    ds.Remove(1)

    // Verify invariants still hold
    assert.True(t, ds.IsValid())
}
```

### CONCURRENT TESTING:
```go
func TestConcurrentAccess(t *testing.T) {
    // Run with: go test -race
    counter := mypackage.NewSafeCounter()

    const goroutines = 100
    const increments = 1000

    var wg sync.WaitGroup
    wg.Add(goroutines)

    for i := 0; i < goroutines; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < increments; j++ {
                counter.Increment()
            }
        }()
    }

    wg.Wait()

    expected := goroutines * increments
    assert.Equal(t, expected, counter.Value())
}

func TestNoGoroutineLeaks(t *testing.T) {
    initial := runtime.NumGoroutine()

    // Create and cancel workers
    ctx, cancel := context.WithCancel(context.Background())
    for i := 0; i < 10; i++ {
        go mypackage.Worker(ctx, nil)
    }

    // Cancel and wait for cleanup
    cancel()
    time.Sleep(100 * time.Millisecond)

    final := runtime.NumGoroutine()

    // Allow for some variance but detect significant leaks
    assert.InDelta(t, initial, final, 2, "Goroutine leak detected")
}
```

### BENCHMARK TESTING:
```go
func BenchmarkOperation(b *testing.B) {
    input := setupBenchmarkData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mypackage.Operation(input)
    }
}

func BenchmarkOperationWithAllocs(b *testing.B) {
    input := setupBenchmarkData()

    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mypackage.Operation(input)
    }
}
```

## Documentation Standards

### GODOC COMMENTS:
```go
// Package mypackage provides functionality for managing GenTx operations.
//
// This package implements a B+ tree data structure optimized for
// on-disk storage with configurable order and caching.
package mypackage

// DataStructure represents a thread-safe collection that maintains
// the following invariants:
//   - All keys are unique
//   - Keys are maintained in sorted order
//   - Size is always accurate
//
// DataStructure is safe for concurrent use by multiple goroutines.
type DataStructure struct {
    mu    sync.RWMutex
    items map[string]Item
}

// Insert adds a key-value pair to the data structure.
//
// If the key already exists, the old value is returned and the new value
// replaces it. If the key is new, nil is returned.
//
// Insert is safe to call concurrently from multiple goroutines.
//
// Example:
//
//	ds := NewDataStructure()
//	oldValue, err := ds.Insert("key", "value")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if oldValue != nil {
//	    fmt.Printf("Replaced old value: %v\n", oldValue)
//	}
//
// Returns an error if:
//   - key is empty (ErrEmptyKey)
//   - key length exceeds maximum (ErrKeyTooLong)
//   - value is nil (ErrNilValue)
func (ds *DataStructure) Insert(key string, value interface{}) (interface{}, error) {
    if key == "" {
        return nil, ErrEmptyKey
    }

    ds.mu.Lock()
    defer ds.mu.Unlock()

    old := ds.items[key]
    ds.items[key] = value
    return old, nil
}
```

### ERROR DOCUMENTATION:
```go
// Common errors returned by this package.
var (
    // ErrNotFound is returned when a requested item does not exist.
    ErrNotFound = errors.New("item not found")

    // ErrEmptyKey is returned when an empty key is provided.
    ErrEmptyKey = errors.New("key cannot be empty")

    // ErrKeyTooLong is returned when a key exceeds the maximum length.
    ErrKeyTooLong = errors.New("key exceeds maximum length")

    // ErrDuplicateKey is returned when attempting to insert a duplicate key.
    ErrDuplicateKey = errors.New("duplicate key")
)
```

## Static Analysis and Tooling

### REQUIRED CHECKS:
```bash
# Before committing code, run:
go fmt ./...                              # Format code
go vet ./...                              # Detect suspicious constructs
staticcheck ./...                         # Advanced static analysis
golangci-lint run                         # Comprehensive linting
go test -race -cover ./...                # Tests with race detector
go test -coverprofile=coverage.out ./...  # Coverage analysis
go tool cover -func=coverage.out          # View coverage details
```

### GOLANGCI-LINT CONFIGURATION:
Key linters to enable:
- `errcheck` - Check for unchecked errors
- `gosimple` - Simplify code
- `govet` - Go vet compatibility
- `ineffassign` - Detect ineffectual assignments
- `staticcheck` - Advanced static analysis
- `unused` - Find unused code
- `gosec` - Security issues
- `goconst` - Find repeated strings that could be constants
- `gocyclo` - Cyclomatic complexity
- `dupl` - Code duplication

## Project-Specific Standards

### GENUTILS PROJECT:
- Follow TDD red-green-refactor cycle strictly
- Maintain ≥90% test coverage target for critical paths
- Document architectural decisions (defense-in-depth, etc.)
- Use domain-driven design patterns consistently
- Validate all domain object invariants at construction time
- Provide both positive and negative test cases
- Test integration points between layers

### NAMING CONVENTIONS:
- Use descriptive names: `validatorAddress` not `va`
- Exported identifiers: `PascalCase`
- Unexported identifiers: `camelCase`
- Constants: `PascalCase` or `SCREAMING_SNAKE_CASE` for errors
- Test functions: `TestFunctionName_Scenario`
- Benchmark functions: `BenchmarkFunctionName`

This system prompt addition establishes clear quality standards, testing requirements, and architectural principles for Go development that align with the current project's TDD approach and high-quality codebase goals.
