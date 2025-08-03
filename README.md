# MySQL Diff

A powerful Go tool for analyzing differences between MySQL database schemas and generating ALTER statements for schema migrations.

## Features

- **Schema Comparison**: Compare two MySQL table definitions and identify all differences
- **ALTER Statement Generation**: Automatically generate SQL ALTER statements to migrate from one schema to another
- **Comprehensive Analysis**: Detects changes in:
  - Column definitions (data types, constraints, defaults, etc.)
  - Primary keys
  - Indexes
  - Foreign keys
  - Table options (engine, charset, collation, etc.)
  - Partitioning
- **Type-Safe API**: Built with Go generics for robust, compile-time type safety
- **Detailed Reporting**: Human-readable diff summaries with change counts
- **JSON Output**: Structured output for programmatic integration

## Installation

### From Source

```bash
git clone https://github.com/n0madic/mysql-diff.git
cd mysql-diff
go build ./cmd/mysql-diff
```

### Using Go Install

```bash
go install github.com/n0madic/mysql-diff/cmd/mysql-diff@latest
```

## Usage

### Command Line Interface

```bash
# Compare two SQL files
mysql-diff old_schema.sql new_schema.sql

# Compare specific tables
mysql-diff --table users old_schema.sql new_schema.sql

# Output detailed diff report
mysql-diff --detailed old_schema.sql new_schema.sql

# JSON output for programmatic use
mysql-diff --json old_schema.sql new_schema.sql
```

### Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/n0madic/mysql-diff/pkg/parser"
    "github.com/n0madic/mysql-diff/pkg/diff"
)

func main() {
    // Parse SQL schemas
    oldTables, err := parser.ParseSQLDump(oldSQL)
    if err != nil {
        panic(err)
    }

    newTables, err := parser.ParseSQLDump(newSQL)
    if err != nil {
        panic(err)
    }

    // Compare tables
    analyzer := diff.NewTableDiffAnalyzer()
    tableDiff := analyzer.CompareTables(oldTables[0], newTables[0])

    // Check for changes
    if tableDiff.HasChanges() {
        fmt.Printf("Found %d column changes\n", len(tableDiff.ColumnDiffs))

        // Access typed changes
        for _, colDiff := range tableDiff.ColumnDiffs {
            if colDiff.Changes.DataType != nil {
                fmt.Printf("Column %s: %s -> %s\n",
                    colDiff.Name,
                    colDiff.Changes.DataType.Old,
                    colDiff.Changes.DataType.New)
            }
        }
    }
}
```

## API Reference

### Core Types

#### TableDiff
Represents the complete difference analysis between two tables:

```go
type TableDiff struct {
    OldTable *parser.CreateTableStatement
    NewTable *parser.CreateTableStatement

    // Change flags
    TableNameChanged    bool
    TableOptionsChanged bool

    // Component differences
    ColumnDiffs      []ColumnDiff
    PrimaryKeyDiff   *PrimaryKeyDiff
    IndexDiffs       []IndexDiff
    ForeignKeyDiffs  []ForeignKeyDiff
    TableOptionsDiff *TableOptionsDiff
    PartitionDiff    *PartitionDiff

    // Summary counters
    ColumnsAdded        int
    ColumnsRemoved      int
    ColumnsModified     int
    // ... more counters
}
```

#### FieldChange[T]
Generic type representing a change in a specific field:

```go
type FieldChange[T any] struct {
    Old T `json:"old"`
    New T `json:"new"`
}
```

#### Typed Change Structures

All changes are strongly typed for better IDE support and compile-time safety:

- **ColumnChanges**: DataType, Nullable, DefaultValue, AutoIncrement, etc.
- **IndexChanges**: Name, IndexType, Columns, KeyBlockSize, etc.
- **ForeignKeyChanges**: Name, Columns, ReferenceTable, OnDelete, etc.
- **PrimaryKeyChanges**: Columns, Name, Using, Comment
- **TableOptionsChanges**: Engine, AutoIncrement, CharacterSet, etc.
- **PartitionChanges**: Type, Linear, Expression, Columns, etc.

### Key Methods

#### HasChanges()
All change structures implement `HasChanges()` to check if any modifications exist:

```go
if tableDiff.HasChanges() {
    // Process changes
}

if colDiff.Changes.HasChanges() {
    // Process column changes
}
```

#### GetSummary()
Get a typed summary of all changes:

```go
summary := tableDiff.GetSummary()
fmt.Printf("Columns: +%d -%d ~%d\n",
    summary.Columns.Added,
    summary.Columns.Removed,
    summary.Columns.Modified)
```

## Examples

### Detecting Column Changes

```go
// Column data type change
if colDiff.Changes.DataType != nil {
    fmt.Printf("Data type: %s -> %s\n",
        colDiff.Changes.DataType.Old,
        colDiff.Changes.DataType.New)
}

// Column nullability change
if colDiff.Changes.Nullable != nil {
    fmt.Printf("Nullable: %v -> %v\n",
        colDiff.Changes.Nullable.Old,
        colDiff.Changes.Nullable.New)
}
```

### Detecting Index Changes

```go
for _, idxDiff := range tableDiff.IndexDiffs {
    switch idxDiff.ChangeType {
    case diff.ChangeTypeAdded:
        fmt.Printf("Index added: %s\n", *idxDiff.NewIndex.Name)
    case diff.ChangeTypeRemoved:
        fmt.Printf("Index removed: %s\n", *idxDiff.OldIndex.Name)
    case diff.ChangeTypeModified:
        if idxDiff.Changes.IndexType != nil {
            fmt.Printf("Index type changed: %s -> %s\n",
                idxDiff.Changes.IndexType.Old,
                idxDiff.Changes.IndexType.New)
        }
    }
}
```

### Working with Foreign Keys

```go
for _, fkDiff := range tableDiff.ForeignKeyDiffs {
    if fkDiff.Changes.OnDelete != nil {
        fmt.Printf("Foreign key ON DELETE changed: %v -> %v\n",
            fkDiff.Changes.OnDelete.Old,
            fkDiff.Changes.OnDelete.New)
    }
}
```

## Supported MySQL Features

- **Column Types**: All MySQL data types including DECIMAL, VARCHAR, TEXT, JSON, etc.
- **Column Attributes**: NULL/NOT NULL, DEFAULT values, AUTO_INCREMENT, UNIQUE, etc.
- **Indexes**: PRIMARY, UNIQUE, INDEX, FULLTEXT, SPATIAL
- **Foreign Keys**: Including ON DELETE/UPDATE actions
- **Table Options**: ENGINE, CHARACTER SET, COLLATION, COMMENT, AUTO_INCREMENT
- **Partitioning**: RANGE, LIST, HASH, KEY partitioning
- **Generated Columns**: VIRTUAL and STORED generated columns

## Requirements

- Go 1.23 or higher (uses generics)
- MySQL 5.7+ compatible SQL syntax

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
