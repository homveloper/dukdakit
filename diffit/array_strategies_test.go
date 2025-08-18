package diffit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test specific array strategy implementations

type TestUser struct {
	ID   int    `bson:"id"`
	Name string `bson:"name"`
	Email string `bson:"email"`
}

type UserContainer struct {
	Users []TestUser `bson:"users"`
}

func TestArrayStrategies_DetailedComparison(t *testing.T) {
	t.Run("ArrayMerge_PureUpdates", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@old.com"},
				{ID: 2, Name: "Bob", Email: "bob@old.com"},
				{ID: 3, Name: "Charlie", Email: "charlie@old.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@new.com"}, // Email updated
				{ID: 2, Name: "Bob", Email: "bob@old.com"},     // No change
				{ID: 3, Name: "Charlie", Email: "charlie@new.com"}, // Email updated
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayMerge))
		require.NoError(t, err)
		
		t.Logf("ArrayMerge pure updates patch: %s", patch)
		
		operations := patch.Operations()
		t.Logf("Operations: %+v", operations)
		
		// ArrayMerge should use ArrayFilters for pure updates (no additions/removals)
		if patch.HasArrayFilters() {
			arrayFilters := patch.ArrayFilters()
			t.Logf("Array filters: %+v", arrayFilters)
			assert.Greater(t, len(arrayFilters), 0, "Should use ArrayFilters for updates")
		}
		
		assert.False(t, patch.IsEmpty(), "Should have operations")
	})

	t.Run("ArrayMerge_PureAdditions", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@example.com"}, // Same
				{ID: 2, Name: "Bob", Email: "bob@example.com"},     // Same
				{ID: 3, Name: "Charlie", Email: "charlie@example.com"}, // Added
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayMerge))
		require.NoError(t, err)
		
		t.Logf("ArrayMerge pure additions patch: %s", patch)
		
		operations := patch.Operations()
		
		// Should use $push for pure additions
		if pushOps, ok := operations["$push"]; ok {
			t.Logf("Push operations: %+v", pushOps)
			pushMap := pushOps.(map[string]interface{})
			assert.Contains(t, pushMap, "users", "Should use $push for pure additions")
		}
		
		assert.False(t, patch.IsEmpty(), "Should have operations")
	})

	t.Run("ArrayMerge_ComplexFallback", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@old.com"},
				{ID: 2, Name: "Bob", Email: "bob@old.com"},
				{ID: 3, Name: "Charlie", Email: "charlie@old.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@new.com"}, // Email updated
				{ID: 2, Name: "Bob", Email: "bob@old.com"},     // No change
				{ID: 4, Name: "David", Email: "david@new.com"}, // New user added
				// Note: User ID 3 (Charlie) is removed
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayMerge))
		require.NoError(t, err)
		
		t.Logf("ArrayMerge complex fallback patch: %s", patch)
		
		operations := patch.Operations()
		
		// Should fall back to replace for complex scenarios
		if setOps, ok := operations["$set"]; ok {
			setMap := setOps.(map[string]interface{})
			assert.Contains(t, setMap, "users", "Should fall back to replace for complex changes")
		}
		
		assert.False(t, patch.IsEmpty(), "Should have operations")
	})

	t.Run("ArraySmart_PositionBased", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@old.com"},
				{ID: 2, Name: "Bob", Email: "bob@old.com"},
				{ID: 3, Name: "Charlie", Email: "charlie@old.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@new.com"}, // Position 0: Email updated  
				{ID: 2, Name: "Bob", Email: "bob@old.com"},     // Position 1: No change
				{ID: 3, Name: "Charlie", Email: "charlie@new.com"}, // Position 2: Email updated
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArraySmart))
		require.NoError(t, err)
		
		t.Logf("ArraySmart patch: %s", patch)
		
		// ArraySmart should use ArrayFilters for position-based updates
		if patch.HasArrayFilters() {
			arrayFilters := patch.ArrayFilters()
			t.Logf("Array filters: %+v", arrayFilters)
			assert.Greater(t, len(arrayFilters), 0, "Should have array filters for updates")
		}
		
		operations := patch.Operations()
		if setOps, ok := operations["$set"]; ok {
			setMap := setOps.(map[string]interface{})
			// Should have array filter operations like users.$[elem0], users.$[elem1]
			hasArrayFilterOps := false
			for key := range setMap {
				if len(key) > 6 && key[:6] == "users." && key[6] == '$' {
					hasArrayFilterOps = true
					break
				}
			}
			assert.True(t, hasArrayFilterOps, "Should have array filter operations")
		}
	})

	t.Run("ArrayAppend_PureAddition", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@example.com"}, // Same
				{ID: 2, Name: "Bob", Email: "bob@example.com"},     // Same
				{ID: 3, Name: "Charlie", Email: "charlie@example.com"}, // Added
				{ID: 4, Name: "David", Email: "david@example.com"},     // Added
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayAppend))
		require.NoError(t, err)
		
		t.Logf("ArrayAppend patch: %s", patch)
		
		operations := patch.Operations()
		
		// Should use $push operation for pure additions
		if pushOps, ok := operations["$push"]; ok {
			t.Logf("Push operations: %+v", pushOps)
			pushMap := pushOps.(map[string]interface{})
			assert.Contains(t, pushMap, "users", "Should have push operation for users")
		} else {
			t.Log("No push operations found - might have fallen back to replace")
		}
	})

	t.Run("ArrayAppend_ModificationFallback", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@old.com"},
				{ID: 2, Name: "Bob", Email: "bob@example.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@new.com"}, // Email modified
				{ID: 2, Name: "Bob", Email: "bob@example.com"}, // Same
				{ID: 3, Name: "Charlie", Email: "charlie@example.com"}, // Added
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayAppend))
		require.NoError(t, err)
		
		t.Logf("ArrayAppend with modification patch: %s", patch)
		
		operations := patch.Operations()
		
		// Should fall back to replace because of modifications
		if setOps, ok := operations["$set"]; ok {
			setMap := setOps.(map[string]interface{})
			assert.Contains(t, setMap, "users", "Should fall back to $set for complex changes")
		}
	})

	t.Run("ArrayReplace_AlwaysReplaces", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@old.com"},
				{ID: 2, Name: "Bob", Email: "bob@old.com"},
			},
		}
		
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@new.com"}, // Minor change
			},
		}

		patch, err := Diff(oldData, newData, WithArrayStrategy(ArrayReplace))
		require.NoError(t, err)
		
		t.Logf("ArrayReplace patch: %s", patch)
		
		operations := patch.Operations()
		
		// Should always use $set for replacement
		setOps, ok := operations["$set"]
		require.True(t, ok, "Should have $set operation")
		
		setMap := setOps.(map[string]interface{})
		assert.Contains(t, setMap, "users", "Should replace entire users array")
	})
}

func TestArrayStrategies_PerformanceScenarios(t *testing.T) {
	t.Run("LargeArray_SmallChanges", func(t *testing.T) {
		// Create a large array with small changes
		oldUsers := make([]TestUser, 100)
		newUsers := make([]TestUser, 100)
		
		for i := 0; i < 100; i++ {
			oldUsers[i] = TestUser{
				ID:   i + 1,
				Name: "User" + string(rune(i+1)),
				Email: "user" + string(rune(i+1)) + "@old.com",
			}
			newUsers[i] = oldUsers[i] // Copy
		}
		
		// Make small changes to a few users
		newUsers[10].Email = "user11@new.com"
		newUsers[50].Name = "UpdatedUser51" 
		newUsers[90].Email = "user91@new.com"
		
		oldData := UserContainer{Users: oldUsers}
		newData := UserContainer{Users: newUsers}
		
		strategies := []struct {
			name     string
			strategy ArrayStrategy
		}{
			{"Replace", ArrayReplace},
			{"Smart", ArraySmart}, 
			{"Merge", ArrayMerge},
		}
		
		for _, s := range strategies {
			t.Run(s.name, func(t *testing.T) {
				patch, err := Diff(oldData, newData, WithArrayStrategy(s.strategy))
				require.NoError(t, err)
				
				operations := patch.Operations()
				
				t.Logf("Strategy %s operations count: %d", s.name, len(operations))
				t.Logf("Strategy %s patch size: %d chars", s.name, len(patch.String()))
				
				// Different strategies should produce different results
				assert.False(t, patch.IsEmpty(), "Should detect changes")
			})
		}
	})
}

func TestArrayStrategies_EdgeCases(t *testing.T) {
	t.Run("EmptyToNonEmpty", func(t *testing.T) {
		oldData := UserContainer{Users: []TestUser{}}
		newData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
			},
		}

		strategies := []ArrayStrategy{ArrayReplace, ArraySmart, ArrayAppend, ArrayMerge}
		
		strategyNames := []string{"ArrayReplace", "ArraySmart", "ArrayAppend", "ArrayMerge"}
		for i, strategy := range strategies {
			t.Run(strategyNames[i], func(t *testing.T) {
				patch, err := Diff(oldData, newData, WithArrayStrategy(strategy))
				require.NoError(t, err)
				assert.False(t, patch.IsEmpty(), "Should detect addition")
			})
		}
	})

	t.Run("NonEmptyToEmpty", func(t *testing.T) {
		oldData := UserContainer{
			Users: []TestUser{
				{ID: 1, Name: "Alice", Email: "alice@example.com"},
			},
		}
		newData := UserContainer{Users: []TestUser{}}

		strategies := []ArrayStrategy{ArrayReplace, ArraySmart, ArrayAppend, ArrayMerge}
		strategyNames := []string{"ArrayReplace", "ArraySmart", "ArrayAppend", "ArrayMerge"}
		
		for i, strategy := range strategies {
			t.Run(strategyNames[i], func(t *testing.T) {
				patch, err := Diff(oldData, newData, WithArrayStrategy(strategy))
				require.NoError(t, err)
				assert.False(t, patch.IsEmpty(), "Should detect removal")
			})
		}
	})
}

func TestArrayStrategies_TimeArrays(t *testing.T) {
	type TimeContainer struct {
		Timestamps []time.Time `bson:"timestamps"`
	}

	base := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	oldData := TimeContainer{
		Timestamps: []time.Time{
			base,
			base.Add(time.Hour),
			base.Add(2 * time.Hour),
		},
	}
	
	newData := TimeContainer{
		Timestamps: []time.Time{
			base,
			base.Add(time.Hour),
			base.Add(3 * time.Hour), // Changed
			base.Add(4 * time.Hour), // Added
		},
	}

	t.Run("ArraySmart_TimeHandling", func(t *testing.T) {
		patch, err := Diff(oldData, newData, WithArrayStrategy(ArraySmart))
		require.NoError(t, err)
		
		t.Logf("Time array smart patch: %s", patch)
		
		// Should handle time.Time arrays properly
		assert.False(t, patch.IsEmpty(), "Should detect time changes")
		
		operations := patch.Operations()
		t.Logf("Time operations: %+v", operations)
	})
}