# Apache Spark Query Optimization

## Adaptive Query Execution (AQE)

Adaptive Query Execution is a feature in Apache Spark that makes query execution plans adaptive at runtime. It optimizes the execution plan dynamically based on runtime statistics instead of just using initial estimates.

### Key Features

1. **Dynamic Join Reordering**
   - Reorders joins based on actual data sizes at runtime
   - Prevents expensive broadcasts when data is larger than expected
   - Dynamically switches between broadcast and shuffle joins

2. **Coalescing Post-Shuffle Partitions**
   - Reduces number of partitions after shuffles to avoid small partitions
   - Improves performance by reducing overhead
   - Automatic optimization based on data statistics

3. **Skew Join Optimization**
   - Detects skewed data distribution in join keys
   - Splits large partitions into smaller ones
   - Replicates non-skewed side for efficient processing

## Dynamic Partition Pruning (DPP)

Dynamic Partition Pruning optimizes queries by pruning partitions at runtime based on the join condition output from the other side of the join.

Example:
```sql
SELECT * FROM events
WHERE event_date IN (
  SELECT date FROM campaigns WHERE status = 'active'
)
```

The optimizer can:
- Scan the campaigns table first
- Get the active dates
- Only scan event partitions that match these dates

## Column Pruning

Unnecessary columns are automatically removed from the query execution plan. Only columns that are actually needed are read from storage.

## Predicate Pushdown

Filter conditions are pushed down as early as possible in the execution plan to reduce the amount of data processed.

Example:
```sql
SELECT name FROM users WHERE age > 18
```

Instead of:
1. Read all users
2. Filter where age > 18

It executes as:
1. Read users WHERE age > 18 (pushed down)

## Broadcast Join Optimization

Small DataFrames are automatically broadcast to all worker nodes when:
- The small DataFrame fits in memory
- Cost of broadcasting < cost of shuffle

This avoids expensive network shuffles.

## Conclusion

Modern Spark versions include sophisticated optimizers that automatically apply these optimizations. Understanding them helps you write queries that leverage these optimizations effectively.
