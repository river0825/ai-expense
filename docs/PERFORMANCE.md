# Performance Testing & Benchmarking Guide

## Overview

This document describes the performance benchmarking infrastructure for AIExpense, designed to monitor and optimize system performance across all critical paths.

## Benchmark Categories

### 1. Use Case Performance (`test/bench/usecase_bench_test.go`)

Benchmarks for core business logic operations:

#### User Operations
- **BenchmarkAutoSignup** - Measures user registration and category initialization
  - Operations: User creation + 5 category creations
  - Target: < 10ms per signup
  - Relevance: High - called on first user message

#### Expense Operations
- **BenchmarkCreateExpense** - Expense creation with AI categorization
  - Operations: Validate request + AI category suggestion + database insert
  - Target: < 50ms per expense
  - Relevance: High - called for each expense created

- **BenchmarkMultipleCreateExpenses** - Bulk expense creation pattern
  - Tests: Sequential creation of many expenses
  - Target: < 50ms per item
  - Relevance: Medium - batch imports or multi-message parsing

#### Parsing Operations
- **BenchmarkParseConversation** - Message parsing and extraction
  - Operations: AI service call + expense structure extraction
  - Target: < 100ms per message
  - Relevance: High - called for each user message

#### Retrieval Operations
- **BenchmarkGetExpenses** - Expense retrieval for user (100 items)
  - Operations: Filter expenses by user ID
  - Target: < 5ms for 100 items
  - Relevance: High - called when user views expenses

- **BenchmarkExpenseRetrieval** - Large dataset retrieval (1000 items)
  - Tests: Filtering 1000 expenses for specific user
  - Target: < 20ms for 1000 items
  - Relevance: Medium - users with 1+ year of data

#### Complex Operations
- **BenchmarkUserRegistration** - Complete user flow
  - Operations: Auto-signup + category initialization
  - Target: < 10ms
  - Relevance: High - critical user experience metric

- **BenchmarkExpenseCreationWithCategoryLookup** - Creation with category resolution
  - Operations: Create + lookup categories (10 available)
  - Target: < 60ms
  - Relevance: Medium - category suggestions

### 2. Repository Performance (`test/bench/repository_bench_test.go`)

Low-level data access benchmarks:

#### User Repository
- **BenchmarkUserRepositoryCreate** - Single user creation
  - Target: < 0.5ms
  - Note: Memory-based, actual SQLite will be slightly slower

- **BenchmarkUserRepositoryExists** - User lookup (100 users)
  - Target: < 0.1ms per lookup
  - Relevance: Called before each expense operation

#### Expense Repository
- **BenchmarkExpenseRepositoryCreate** - Single expense insert
  - Target: < 1ms
  - Actual SQLite: 2-5ms expected

- **BenchmarkExpenseRepositoryGetByUserID** - User filter (1000 items)
  - Target: < 5ms
  - Tests: Linear scan through 1000 expenses

- **BenchmarkExpenseRepositoryGetByDateRange** - Date range filter
  - Target: < 5ms for 100 items
  - Relevance: Common for report generation

- **BenchmarkExpenseRepositorySequential** - CRUD sequence
  - Tests: Create + Read + Update cycle
  - Target: < 3ms per cycle (memory)

#### Category Repository
- **BenchmarkCategoryRepositoryGetByUserID** - Category lookup (50 items)
  - Target: < 1ms
  - Relevance: Called for category suggestions

- **BenchmarkCategoryRepositoryGetByName** - Name-based lookup (20 items)
  - Target: < 1ms
  - Relevance: Auto-categorization

## Running Benchmarks

### Basic Execution

```bash
# Run all benchmarks
go test -bench=. -benchmem ./test/bench/...

# Run specific benchmark
go test -bench=BenchmarkCreateExpense -benchmem ./test/bench/...

# Run with high iteration count for better accuracy
go test -bench=. -benchmem -benchtime=10s ./test/bench/...
```

### Output Interpretation

```
BenchmarkCreateExpense-8     10000    115000 ns/op    4200 B/op    42 allocs/op
```

- `BenchmarkCreateExpense-8`: Test name with 8 CPUs
- `10000`: Number of iterations
- `115000 ns/op`: 115 microseconds per operation
- `4200 B/op`: 4.2 KB memory allocated per operation
- `42 allocs/op`: 42 memory allocations per operation

### Benchmark Comparisons

```bash
# Generate baseline
go test -bench=. -benchmem ./test/bench/... > baseline.txt

# After optimization
go test -bench=. -benchmem ./test/bench/... > optimized.txt

# Compare results
benchstat baseline.txt optimized.txt
```

## Performance Targets

### Critical Path Operations (Messenger Webhooks)

| Operation | Target | Priority |
|-----------|--------|----------|
| Auto-signup | < 10ms | Critical |
| Parse message | < 100ms | Critical |
| Create expense | < 50ms | Critical |
| Get expenses | < 20ms | High |
| Category lookup | < 5ms | High |

### Acceptable Performance Ranges

- **< 10ms**: Instant (user won't notice)
- **10-50ms**: Good (acceptable for most operations)
- **50-100ms**: Fair (user may notice delay)
- **> 100ms**: Slow (should optimize or async)

## Performance Optimization Strategies

### 1. Database Optimization
- Add indexes on frequently queried columns (user_id, created_at)
- Use prepared statements for repeated queries
- Connection pooling for concurrent operations
- Query result caching for frequently accessed data

### 2. Caching Strategies
- Cache user categories (per user) - invalidate on change
- Cache category-keyword mappings - refresh periodically
- Memo parsing results for identical messages
- LRU cache for recent lookups

### 3. Async Processing
- Process expensive AI calls asynchronously
- Batch notification sending
- Defer non-critical updates

### 4. Code Optimization
- Reduce allocations in hot paths
- Pre-allocate slices when size is known
- Avoid reflection in performance-critical code
- Use sync.Pool for temporary objects

### 5. Concurrency Improvements
- Connection pooling for database
- Read-write lock optimization
- Goroutine pooling for request handling

## Monitoring Performance in Production

### Key Metrics to Track

```go
// Example: Add timing to critical paths
start := time.Now()
expense, err := createExpenseUC.Execute(ctx, req)
duration := time.Since(start)

// Log slow operations
if duration > 50*time.Millisecond {
    log.Warnf("slow expense creation: %v", duration)
}
```

### Profiling Commands

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./test/bench/...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./test/bench/...
go tool pprof mem.prof

# Allocation profiling
go test -allocprofiler=alloc.prof -bench=. ./test/bench/...
go tool pprof alloc.prof
```

## Expected Performance Characteristics

### Current Architecture (Memory-Based Repositories)

- **Auto-signup**: 1-2ms
- **Create expense**: 0.5-1ms (in-memory)
- **Parse conversation**: 5-10ms (depends on AI service)
- **Get expenses**: 0.1-0.5ms (linear scan, 100 items)
- **Database operations**: None (mocked)

### With SQLite (Expected)

- **Auto-signup**: 5-10ms (+ 5 write operations)
- **Create expense**: 5-10ms (+ category lookup + write)
- **Parse conversation**: 5-10ms (unchanged)
- **Get expenses**: 5-10ms (indexed query on 1000 items)

### With Connection Pooling & Caching

- **Auto-signup**: 3-5ms
- **Create expense**: 3-5ms (category cached)
- **Parse conversation**: 2-5ms
- **Get expenses**: 1-2ms (cached)

## Benchmarking Best Practices

### For Accurate Results

1. **Run multiple times**: `-benchtime=10s` for longer runs
2. **On dedicated machine**: Minimize system load
3. **Warm up**: First few iterations are slower
4. **Use statistically significant samples**: Aim for > 1000 iterations

### Avoiding Common Pitfalls

- ❌ Benchmarking in debug mode (use Release builds)
- ❌ Benchmarking with -race flag (significant overhead)
- ❌ Benchmarking with other processes running
- ❌ Using `time.Sleep` instead of proper timing
- ✅ Reset timer after setup: `b.ResetTimer()`
- ✅ Run helper.StopTimer() during expensive setup
- ✅ Compare benchmarks consistently

## Performance Regression Detection

### Automated Checks

Add to CI/CD pipeline:

```bash
# Fail if performance regresses > 10%
go test -bench=. -benchmem ./test/bench/... | \
  grep -E "Benchmark" | awk '{
    if ($NF > baseline[$1] * 1.1) {
      print "Performance regression in " $1
      exit 1
    }
  }'
```

## Future Optimization Opportunities

### Phase 18+ Work

1. **Database Optimization**
   - [ ] Add SQLite indexes for common queries
   - [ ] Implement prepared statement caching
   - [ ] Connection pooling setup

2. **Caching Layer**
   - [ ] Redis integration for distributed caching
   - [ ] In-memory LRU cache for single-instance
   - [ ] Cache invalidation strategy

3. **Async Processing**
   - [ ] Background job queue for expensive operations
   - [ ] Batch processing for notifications
   - [ ] Async parsing for AI calls

4. **Profiling & Analysis**
   - [ ] Add pprof support to HTTP endpoints
   - [ ] Continuous profiling in staging
   - [ ] Performance dashboard

5. **Load Testing**
   - [ ] Create load test scenarios
   - [ ] Test concurrent webhook processing
   - [ ] Identify bottlenecks under load

## Benchmark Results Summary

### Current Status

- **Benchmark Files**: 2 files
- **Total Benchmarks**: 16+ benchmark functions
- **Coverage**: UseCase layer, Repository layer
- **Build Status**: ✅ All benchmarks compile
- **Execution Status**: Ready for local/CI execution

### How to Interpret Results

1. Look for `ns/op` (nanoseconds per operation)
2. Compare against target times (see Performance Targets)
3. Check `B/op` (bytes allocated) for memory efficiency
4. Monitor `allocs/op` (allocation count) for GC pressure

### Next Steps

1. Run benchmarks locally to establish baseline
2. Identify slow operations (> 100ms)
3. Profile hot spots with pprof
4. Implement optimizations
5. Re-benchmark to verify improvements
6. Integrate into CI/CD for regression detection

## Example Benchmark Analysis

```bash
$ go test -bench=BenchmarkCreateExpense -benchmem ./test/bench/...
BenchmarkCreateExpense-8     10000    115000 ns/op    4200 B/op    42 allocs/op

Analysis:
- 115 microseconds per operation ✓ (target: < 50ms)
- 4.2 KB allocated per operation (normal)
- 42 allocations per operation (consider reducing)

Conclusion: Acceptable performance, some optimization potential
```

## Related Documentation

- `TESTING.md` - Testing strategies and best practices
- `PERFORMANCE.md` - This file
- `DEPLOYMENT.md` - Production performance considerations

## Performance Checklist

- [ ] Run benchmarks locally before submitting PR
- [ ] Verify no performance regressions
- [ ] Check memory allocations are reasonable
- [ ] Monitor actual running system in staging
- [ ] Profile hot paths with pprof
- [ ] Add performance tests to CI/CD
- [ ] Document performance assumptions
- [ ] Plan for future scaling

---

**Last Updated**: Phase 18
**Benchmark Infrastructure**: Complete
**Optimization Work**: Phase 19+
