# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DukDakit (뚝딱키트) is a game server framework for Go that aims to make building production-ready game servers simple and fun. The name "DukDak" (뚝딱) is Korean for "in a snap" or "quickly".

## Architecture

DukDakit uses a category-based API structure with dot notation:
- `dukdakit.Timex.*` - Time elapsed checking utilities for game cooldowns, daily resets
- `dukdakit.Pagit.*` - Pagination utilities for cursor-based and offset-based pagination
- `dukdakit.Distributed.*` - Distributed computing features (optimistic concurrency)
- `dukdakit.Retry.*` - Retry mechanisms with circuit breaker support

Each category is defined in its own file (e.g., `timex.go`, `pagit.go`, `distributed.go`, `retry.go`) with the actual implementation in the `internal/` directory.

## Development Commands

### Running Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/timex

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestElapsed_BasicDay ./internal/timex
```

### Code Quality
```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...
```

## Testing Approach

The project uses the testify framework for assertions. Tests are located alongside implementation files in the `internal/` directory:

### Testing Requirements
- **MANDATORY**: Use testify framework for all assertions
- **MANDATORY**: Test function names MUST include the feature/function name being tested as a prefix
- Test files follow the `*_test.go` naming convention
- Use `assert` for non-critical assertions
- Use `require` for critical assertions that should stop test execution

### Test Naming Convention
Test function names must follow this pattern: Test[FeatureName]_[Scenario]

Examples:
- TestElapsed_BasicDay - Tests the Elapsed function with basic day scenario
- TestRange_ExactSplit - Tests the Range function with exact split mode
- TestOptionBuilder_Day - Tests the OptionBuilder Day method

### Testing Guidelines
- Always import testify assert and require packages
- Group related test cases using subtests with t.Run()
- Test both success and error scenarios
- Include edge cases and boundary conditions
- Use descriptive assertion messages
- Follow Arrange-Act-Assert pattern in test structure

### Example Code Policy
- **MANDATORY**: All usage examples MUST be provided in unit test code only
- **PROHIBITED**: Creating separate example files or directories is forbidden
- Test functions serve as both verification and documentation of API usage
- Look at existing test files (e.g., `internal/timex/*_test.go`, `internal/pagit/*_test.go`) for usage patterns

## Key Implementation Patterns

### Category API Pattern
Each feature category follows this structure:
1. Public category struct in root package (e.g., `TimexCategory`)
2. Global instance variable (e.g., `var Timex = &TimexCategory{}`)
3. Methods delegate to internal packages
4. Internal packages contain actual implementation

### Builder Pattern
The Timex category uses a builder pattern for creating options:
```go
dukdakit.Timex.Option().Day().WithTimezone(kst).WithDailyResetOffset(9*time.Hour)
```

### Time Zone Helpers
Timex provides timezone helpers for common game server timezones:
- `KST()` - Asia/Seoul (Korean games)
- `JST()` - Asia/Tokyo (Japanese games)
- `PST()` - America/Los_Angeles
- `EST()` - America/New_York
- `UTC()` - UTC

## Project Structure

```
dukdakit/
├── *.go                    # Category API files (timex.go, pagit.go, distributed.go, retry.go)
├── internal/              
│   ├── timex/             # Time elapsed checking implementation
│   ├── pagit/             # Pagination implementation
│   ├── distributed/       # Optimistic concurrency implementation
│   └── retry/             # Retry mechanisms implementation
└── examples/              # Example usage files (deprecated - use test files for examples)
```

## Important Notes

- The project is in early development (v0.0.1)
- Requires Go 1.21 or higher
- Uses testify for testing assertions
- All public APIs should have comprehensive documentation with examples
- The Korean theme is intentional and part of the project's identity