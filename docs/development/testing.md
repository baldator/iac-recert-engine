# 🧪 Testing

This guide covers the testing practices and procedures for IaC Recertification Engine.

## Testing Overview

ICE uses a comprehensive testing strategy that includes:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete workflows
- **Test Coverage**: Ensure adequate code coverage

## Testing Framework

### Dependencies
- **testify**: Assertion and mocking library
- **Go testing**: Standard Go testing framework
- **httptest**: HTTP testing utilities
- **tempfile**: Temporary file creation for tests

### Test Structure
Tests are organized alongside source code with `_test.go` suffix:

```
internal/scan/
├── file_scanner.go
├── file_scanner_test.go
├── recert_checker.go
└── recert_checker_test.go
```

## Running Tests

### Basic Test Execution
```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test ./internal/scan/...

# Run a specific test
go test -run TestScanner_Scan ./internal/scan/
```

### Test Options

#### Coverage Analysis
```bash
# Generate coverage report
go test -cover ./...

# Generate detailed coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage for specific package
go test -cover ./internal/scan/
```

#### Race Detection
```bash
# Run tests with race detector
go test -race ./...

# Race detection with coverage
go test -race -cover ./...
```

#### Benchmarking
```bash
# Run benchmarks
go test -bench=. ./...

# Benchmark with memory allocation stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkScanner ./internal/scan/
```

#### Other Options
```bash
# Run tests in parallel (default)
go test -parallel=4 ./...

# Run tests with timeout
go test -timeout=30s ./...

# Run tests with CPU profiling
go test -cpuprofile=cpu.out ./...

# Run tests with memory profiling
go test -memprofile=mem.out ./...
```

## Writing Tests

### Unit Test Structure
```go
package scan

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionName(t *testing.T) {
	// Arrange
	input := "test input"
	expected := "expected output"

	// Act
	result := FunctionName(input)

	// Assert
	assert.Equal(t, expected, result)
}
```

### Table-Driven Tests
```go
func TestScanner_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid input",
			input:    "valid",
			expected: "result",
			wantErr:  false,
		},
		{
			name:     "invalid input",
			input:    "invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Scanner_Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
```

### Test Helpers

#### Temporary Files and Directories
```go
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-*")
	require.NoError(t, err)
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	return tmpFile.Name()
}

func createTempDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	return tmpDir
}
```

#### Mock Setup
```go
type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) GetPullRequests(owner, repo string) ([]PullRequest, error) {
	args := m.Called(owner, repo)
	return args.Get(0).([]PullRequest), args.Error(1)
}

func TestEngine_Run(t *testing.T) {
	mockProv := &mockProvider{}
	mockProv.On("GetPullRequests", "owner", "repo").Return([]PullRequest{}, nil)

	engine := &Engine{provider: mockProv}
	err := engine.Run()

	mockProv.AssertExpectations(t)
	assert.NoError(t, err)
}
```

### Integration Tests

#### Repository Setup
```go
func setupTestRepo(t *testing.T) string {
	t.Helper()

	// Create temporary directory
	tmpDir := createTempDir(t)

	// Initialize git repo
	require.NoError(t, exec.Command("git", "init", tmpDir).Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run())

	// Create and commit test files
	testFile := filepath.Join(tmpDir, "main.tf")
	err := os.WriteFile(testFile, []byte("resource \"aws_instance\" \"test\" {}"), 0644)
	require.NoError(t, err)

	require.NoError(t, exec.Command("git", "-C", tmpDir, "add", ".").Run())
	require.NoError(t, exec.Command("git", "-C", tmpDir, "commit", "-m", "Initial commit").Run())

	return tmpDir
}
```

#### HTTP Testing
```go
func TestProvider_GetPullRequests(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]PullRequest{{Number: 1, Title: "Test PR"}})
	}))
	defer server.Close()

	// Create provider with test server URL
	provider := &GitHubProvider{
		baseURL: server.URL,
		client:  server.Client(),
	}

	prs, err := provider.GetPullRequests("owner", "repo")
	require.NoError(t, err)
	assert.Len(t, prs, 1)
	assert.Equal(t, 1, prs[0].Number)
}
```

## Test Categories

### Unit Tests
- Test individual functions and methods
- Mock external dependencies
- Fast execution
- No external resources required

### Integration Tests
- Test component interactions
- May use real dependencies (database, API calls)
- Slower execution
- May require test environment setup

### End-to-End Tests
- Test complete user workflows
- Use real configurations and repositories
- Slowest execution
- Most comprehensive validation

## Test Coverage Goals

### Coverage Targets
- **Overall Coverage**: >80%
- **Critical Paths**: >90%
- **New Code**: >85%

### Coverage Analysis
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by function
go tool cover -func=coverage.out

# View coverage by package
go tool cover -func=coverage.out | grep -E "^[^/]+/"
```

### Coverage Badges
Coverage reports can be integrated with:
- GitHub Actions
- Codecov
- Coveralls
- Local CI/CD pipelines

## Continuous Integration

### GitHub Actions Setup
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - name: Run tests
        run: go test -v -cover ./...
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Quality Gates
- All tests must pass
- Coverage must meet minimum thresholds
- No linting errors
- Build must succeed

## Best Practices

### Test Naming
- Use descriptive names: `TestFunctionName_Scenario_Result`
- Follow Go naming conventions
- Use underscores for readability

### Test Organization
- Group related tests in subtests
- Use table-driven tests for similar scenarios
- Keep tests focused and independent

### Assertions
- Use `require` for setup and critical checks
- Use `assert` for result validation
- Provide clear error messages

### Mocking
- Mock external dependencies
- Use interfaces for testable design
- Verify mock expectations

### Cleanup
- Use `t.Cleanup()` for resource cleanup
- Remove temporary files and directories
- Reset global state between tests

### Performance
- Keep unit tests fast (<100ms)
- Use build tags to skip slow tests
- Parallelize independent tests

## Debugging Tests

### Verbose Output
```bash
# Run with detailed logging
go test -v -args -debug

# Run single failing test
go test -run TestName -v
```

### Debugging Tools
```bash
# Use delve for debugging
dlv test ./internal/scan/

# Profile test performance
go test -cpuprofile=cpu.out -memprofile=mem.out ./...

# Analyze profiles
go tool pprof cpu.out
```

### Common Issues
- **Race conditions**: Use `-race` flag
- **Flaky tests**: Check for external dependencies
- **Timeout issues**: Increase timeout or optimize test
- **Resource leaks**: Check for proper cleanup

## Contributing Tests

When contributing code:

1. **Write tests first** (TDD approach)
2. **Test edge cases** and error conditions
3. **Update existing tests** when changing functionality
4. **Ensure coverage** meets requirements
5. **Run full test suite** before submitting PR

### Test Checklist
- [ ] Unit tests for new functions
- [ ] Integration tests for new features
- [ ] Error handling tests
- [ ] Edge case coverage
- [ ] Documentation updated
- [ ] CI passes all tests

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [Go Testing Best Practices](https://github.com/golang/go/wiki/TestComments)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
