# CASCADE Delete Testing Limitations

## Overview

This document explains why CASCADE delete operations cannot be properly tested with our current sqlmock-based testing infrastructure.

## The Problem

sqlmock is a mock library that intercepts SQL queries and returns predefined results. However, it does **NOT** simulate:
- Foreign key constraints
- CASCADE delete operations
- Database triggers
- Check constraints
- Any other database-level integrity rules

## Why This Matters

The bug with `DeleteEncounter` failing due to foreign key constraints was not caught by our tests because:

1. The test only verified that a DELETE query was executed
2. sqlmock returned a success result without checking constraints
3. In the real database, the DELETE would fail due to foreign key violations in `combatant_conditions`

## Tables with CASCADE Deletes

The following tables have ON DELETE CASCADE relationships:

### When deleting from `encounters`:
- `encounter_monsters` - CASCADE deletes all monster associations
- `encounter_players` - CASCADE deletes all player associations
- `combatant_conditions` - CASCADE deletes all conditions (after migration 000013)

### When deleting from `parties`:
- `players` - CASCADE deletes all players in the party

### When deleting from `users`:
- `encounters` - CASCADE deletes all user's encounters
- `parties` - CASCADE deletes all user's parties

### When deleting from `monsters` or `players`:
- Related records in `encounter_monsters` or `encounter_players` CASCADE delete

## Testing Recommendations

### 1. Current Unit Tests (with sqlmock)
- Continue using for testing business logic
- Add comments documenting CASCADE behavior that cannot be tested
- Focus on testing error handling, data transformation, and query construction

### 2. Integration Tests (recommended addition)
- Use a real PostgreSQL database (possibly with testcontainers)
- Test actual CASCADE delete behavior
- Verify foreign key constraints work as expected
- Can catch database-level issues that mocks miss

### 3. Manual Testing
- Always test DELETE operations in a development environment
- Verify CASCADE deletes work as expected
- Check for orphaned records

## Example Integration Test Structure

```go
// tests/integration/cascade_test.go
// +build integration

func TestDeleteEncounter_WithCascade(t *testing.T) {
    // Use real database connection
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    // Create test data
    encounterID := createTestEncounter(t, db)
    addTestMonsters(t, db, encounterID)
    addTestPlayers(t, db, encounterID)
    addTestConditions(t, db, encounterID)
    
    // Delete encounter
    err := models.DeleteEncounter(db, encounterID)
    require.NoError(t, err)
    
    // Verify CASCADE deletes
    assertNoOrphanedMonsters(t, db, encounterID)
    assertNoOrphanedPlayers(t, db, encounterID)
    assertNoOrphanedConditions(t, db, encounterID)
}
```

## Conclusion

While sqlmock is excellent for unit testing business logic, it cannot catch database constraint violations. Critical operations like DELETE that rely on CASCADE behavior should be additionally tested with integration tests against a real database.