# Goje - Go SQL Query Builder

A lightweight, fluent SQL query builder for Go that generates safe SQL queries with proper parameter binding. Goje provides an intuitive interface for building complex SQL queries while preventing SQL injection attacks.

## Features

- 🔒 **SQL Injection Safe** - Uses parameter binding for all values
- 🔧 **Fluent Interface** - Chainable methods for building queries
- 📊 **Complex Query Support** - Joins, subqueries, aggregations, and more
- 🚀 **Transaction Support** - Built-in transaction handling
- 🎯 **MySQL Optimized** - Currently supports MySQL with room for expansion
- 🧪 **Well Tested** - Comprehensive test coverage
- ⚡ **Performance Focused** - Connection pooling and efficient query building

## Installation

```bash
go get github.com/genigo/goje
```

## Quick Start

### Database Configuration

```go
package main

import (
    "log"
    "github.com/genigo/goje"
)

func main() {
    config := &goje.DBConfig{
        Driver:   "mysql",
        Host:     "127.0.0.1",
        Port:     3306,
        User:     "root",
        Password: "password",
        Schema:   "mydb",
        MaxOpenConns:    25,
        MaxIdleConns:    10,
        MaxIdleTime:     5 * time.Minute,
        ConnMaxLifetime: 30 * time.Minute,
    }

    err := goje.InitDB(config)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Basic SELECT Query

```go
// Simple select
query, args, err := goje.SelectQueryBuilder(
    "users", 
    []string{"id", "name", "email"}, 
    []goje.QueryInterface{
        goje.Where("active = ?", true),
        goje.Order("created_at DESC"),
        goje.Limit(10),
    },
)
// Result: SELECT `id`,`name`,`email` FROM users WHERE (active = ?) ORDER BY created_at DESC LIMIT ?
```

### Using the Context Handler

```go
// Get a database handler
handler := goje.H()

// Execute query
rows, err := handler.DB.QueryContext(handler.Ctx, query, args...)
```

## Query Building

### WHERE Conditions

```go
// Basic WHERE
goje.Where("age > ?", 18)
goje.Where("name = ?", "John")

// Contains (LIKE with wildcards)
goje.Contains("name", "john") // name LIKE '%john%'
goje.EndsWith("name", "john") // name LIKE '%john'
goje.StartsWith("name", "john") // name LIKE 'john%'
goje.Eq("name", "john") // name = 'john'
goje.Lt("age",11) // age < 1
goje.Lte("age",11) // age <= 1
goje.Gt("age",11) // age > 1
goje.Gte("age",11) // age >= 1

// IN conditions  
goje.WhereIn("status", "active", "pending", "approved")
goje.WhereNotIn("role", "admin", "moderator")

// OR conditions
goje.OR(
    goje.Where("status = ?", "active"),
    goje.Where("priority = ?", "high"),
)
```

### JOINs

```go
// Different types of joins
goje.InnerJoin("orders", "orders.user_id = users.id")
goje.LeftJoin("profiles", "profiles.user_id = users.id")
goje.RightJoin("addresses", "addresses.user_id = users.id")
goje.OuterJoin("permissions", "permissions.user_id = users.id")
goje.NaturalJoin("roles", "")
```

### GROUP BY and HAVING

```go
goje.GroupBy("department")
goje.GroupBy("status", "priority") // Multiple columns
goje.Having("COUNT(*) > ?", 5)
goje.Having("AVG(salary) > ?", 50000)
```

### ORDER BY and LIMITS

```go
goje.Order("created_at DESC")
goje.Order("priority ASC, created_at DESC")
goje.Limit(20)
goje.Offset(40)
```

## Complex Query Examples

### Multi-table Query with Aggregation

```go
query, args, err := goje.SelectQueryBuilder(
    "users",
    []string{
        "users.id",
        "users.name", 
        "COUNT(orders.id) as order_count",
        "SUM(orders.total) as total_spent",
    },
    []goje.QueryInterface{
        goje.LeftJoin("orders", "orders.user_id = users.id"),
        goje.Where("users.active = ?", true),
        goje.Where("users.created_at > ?", "2023-01-01"),
        goje.GroupBy("users.id", "users.name"),
        goje.Having("COUNT(orders.id) > ?", 0),
        goje.Order("total_spent DESC"),
        goje.Limit(50),
    },
)
```

### Complex OR Conditions

```go
query, args, err := goje.SelectQueryBuilder(
    "products",
    []string{"*"},
    []goje.QueryInterface{
        goje.Where("category = ?", "electronics"),
        goje.OR(
            goje.Where("price < ?", 100),
            goje.Where("on_sale = ?", true),
            goje.WhereIn("brand", "apple", "samsung", "sony"),
        ),
        goje.Order("price ASC"),
    },
)
```

## Raw Operations

### Raw Delete

```go
handler := goje.H()
affected, err := handler.RawDelete("users", []goje.QueryInterface{
    goje.Where("active = ?", false),
    goje.Where("last_login < ?", "2022-01-01"),
})
```

### Raw Update

```go
updates := map[string]any{
    "status": "inactive",
    "updated_at": time.Now(),
}

affected, err := handler.RawUpdate("users", updates, 
    goje.Where("last_login < ?", "2022-01-01"),
)
```

### Bulk Insert

```go
users := []map[string]any{
    {"name": "John Doe", "email": "john@example.com", "age": 30},
    {"name": "Jane Smith", "email": "jane@example.com", "age": 25},
    {"name": "Bob Johnson", "email": "bob@example.com", "age": 35},
}

affected, err := handler.RawBulkInsert("users", users)
```

## Transactions

### Basic Transaction

```go
ctx := context.Background()
txHandler, err := goje.MakeTxHandler(ctx, nil)
if err != nil {
    log.Fatal(err)
}

// Perform operations
_, err = txHandler.RawUpdate("users", map[string]any{
    "credits": 100,
}, goje.Where("id = ?", userID))

if err != nil {
    txHandler.Rollback()
    return err
}

// Commit transaction
err = txHandler.Commit()
```

### Transaction with Custom Options

```go
opts := &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
    ReadOnly:  false,
}

txHandler, err := goje.MakeTxHandler(ctx, opts)
```

## Configuration Options

### Database Configuration

```yaml
# config.yaml example
driver: mysql
host: 127.0.0.1
port: 3306
user: myuser
password: mypassword
schema: mydatabase
flags:
  charset: utf8mb4
  parseTime: "True"
  loc: Local
MaxIdleTime: 300s
MaxOpenConns: 25
MaxIdleConns: 10
ConnMaxLifetime: 1800s
```

### Connection Pool Settings

- **MaxOpenConns**: Maximum number of open connections to the database
- **MaxIdleConns**: Maximum number of connections in the idle connection pool
- **MaxIdleTime**: Maximum time a connection may be idle before being closed
- **ConnMaxLifetime**: Maximum time a connection may be reused

## Error Handling

```go
// Common errors
var (
    ErrHandlerIsNil        = errors.New("context handler doesn't set properly")
    ErrRecursiveLoad       = errors.New("recursive load is forbidden")
    ErrNoColsSetForUpdate  = errors.New("cols should have at least one property for update")
    ErrNoRowsForInsert     = errors.New("there isn't any row for insert into database")
    ErrUnknownDBDriver     = errors.New("goje doesn't support this driver")
    ErrIsntATx             = errors.New("it isn't a transactional context")
    ErrTxIsntSet           = errors.New("there is not any transaction context")
)
```

## Best Practices

### 1. Use Context Handlers

```go
// Preferred: Use context for timeouts and cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

handler := goje.MakeHandler(ctx)
```

### 2. Connection Management

```go
// Use connection pools effectively
config.MaxOpenConns = 25    // Adjust based on your database capacity
config.MaxIdleConns = 10    // Keep some connections ready
config.ConnMaxLifetime = 30 * time.Minute // Prevent stale connections
```

### 3. Query Building

```go
// Build reusable query components
baseQuery := []goje.QueryInterface{
    goje.Where("active = ?", true),
    goje.Order("created_at DESC"),
}

// Extend for specific use cases
adminQuery := append(baseQuery, goje.Where("role = ?", "admin"))
```

### 4. Error Handling

```go
query, args, err := goje.SelectQueryBuilder(table, columns, conditions)
if err != nil {
    log.Printf("Query building failed: %v", err)
    return err
}

rows, err := handler.DB.QueryContext(handler.Ctx, query, args...)
if err != nil {
    log.Printf("Query execution failed: %v", err)
    return err
}
defer rows.Close()
```

## Supported SQL Features

- ✅ SELECT queries with complex WHERE conditions
- ✅ INSERT (single and bulk)
- ✅ UPDATE with conditions  
- ✅ DELETE with conditions
- ✅ JOINs (INNER, LEFT, RIGHT, OUTER, NATURAL)
- ✅ GROUP BY and HAVING clauses
- ✅ ORDER BY with multiple columns
- ✅ LIMIT and OFFSET
- ✅ IN and NOT IN conditions
- ✅ OR conditions with nesting
- ✅ Transaction support
- ✅ Connection pooling
- ✅ Parameter binding (SQL injection safe)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## Testing

```bash
go test ./...
go test -v ./... # verbose output
go test -cover ./... # with coverage
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Roadmap

- [ ] PostgreSQL support
- [ ] SQLite support  
- [ ] Subquery support
- [ ] CTE (Common Table Expressions)
- [ ] Window functions
- [ ] Schema migrations
- [ ] Query caching
- [ ] Performance benchmarks

## Support

For questions, issues, or contributions, please visit our [GitHub repository](https://github.com/genigo/goje).