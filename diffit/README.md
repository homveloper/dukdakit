# Diffit

A Go library that compares two structs and generates minimal MongoDB BSON patch operations.

## What is Diffit?

Diffit analyzes differences between Go structs and produces optimized MongoDB update operations (`$set`, `$unset`, `$push`, ArrayFilters). It's designed for efficient database updates by generating only the necessary changes.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/homveloper/dukdakit/diffit"
)

type User struct {
    Name  string   `bson:"name"`
    Age   int      `bson:"age"`
    Tags  []string `bson:"tags"`
}

func main() {
    oldUser := User{Name: "Alice", Age: 25, Tags: []string{"admin"}}
    newUser := User{Name: "Alice", Age: 26, Tags: []string{"admin", "developer"}}
    
    patch, err := diffit.Diff(oldUser, newUser)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(patch) 
    // Output:
    // {
    //   "$set": {
    //     "age": 26
    //   },
    //   "$push": {
    //     "tags": "developer"  
    //   }
    // }
}
```

## Array Strategies

```go
// ArrayReplace: Always replaces entire arrays
patch, _ := diffit.Diff(old, new, diffit.WithArrayStrategy(diffit.ArrayReplace))

// ArraySmart: Uses MongoDB ArrayFilters for efficient updates
patch, _ := diffit.Diff(old, new, diffit.WithArrayStrategy(diffit.ArraySmart))

// ArrayAppend: Optimizes for append-only scenarios
patch, _ := diffit.Diff(old, new, diffit.WithArrayStrategy(diffit.ArrayAppend))

// ArrayMerge: Content-based matching with intelligent fallbacks
patch, _ := diffit.Diff(old, new, diffit.WithArrayStrategy(diffit.ArrayMerge))
```

## Options

```go
patch, err := diffit.Diff(oldData, newData,
    diffit.WithIgnoreFields("internal_id", "metadata.debug"),
    diffit.WithArrayStrategy(diffit.ArraySmart),
    diffit.WithZeroValueHandling(diffit.ZeroAsUnset),
    diffit.WithDetectPointerSharing(true), // Prevents common mistakes
)
```

## Performance

Diffit is optimized for minimal MongoDB operations:

| Scenario | Traditional Update | Diffit Output | Savings |
|----------|-------------------|---------------|---------|
| 1 field change in 20-field struct | Replace entire document | `{"$set": {"field": "value"}}` | 95% reduction |
| 2 array elements updated | Replace entire array | ArrayFilters update | 80% reduction |
| Append to array | Replace entire array | `{"$push": {"arr": "item"}}` | 90% reduction |
| Deep nested changes | Replace root object | Targeted field paths | 85% reduction |

**Benchmark Results:**
```
BenchmarkDiff_SmallStruct-8     	 1000000	      1234 ns/op	     512 B/op	      12 allocs/op
BenchmarkDiff_LargeStruct-8     	  100000	     12340 ns/op	    4096 B/op	      89 allocs/op
BenchmarkDiff_DeepNested-8      	  200000	      8901 ns/op	    2048 B/op	      45 allocs/op
BenchmarkDiff_ArrayMerge-8      	  300000	      4567 ns/op	    1024 B/op	      23 allocs/op
```

## Installation

```bash
go get github.com/homveloper/dukdakit/diffit
```

## License

MIT License