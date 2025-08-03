package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/diff"
	"github.com/n0madic/mysql-diff/pkg/parser"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: sql-diff <old_sql_file> <new_sql_file>")
		fmt.Println("Examples:")
		fmt.Println("  sql-diff old_schema.sql new_schema.sql")
		fmt.Println("  sql-diff auth_v1.sql auth_v2.sql")
		os.Exit(1)
	}

	oldFile := os.Args[1]
	newFile := os.Args[2]

	// Read and parse old file
	fmt.Printf("Reading old schema from: %s\n", oldFile)
	oldContent, err := os.ReadFile(oldFile)
	if err != nil {
		fmt.Printf("Error reading old file: %v\n", err)
		os.Exit(1)
	}

	oldTables, err := parser.ParseSQLDump(string(oldContent))
	if err != nil {
		fmt.Printf("Error parsing old SQL: %v\n", err)
		os.Exit(1)
	}

	// Read and parse new file
	fmt.Printf("Reading new schema from: %s\n", newFile)
	newContent, err := os.ReadFile(newFile)
	if err != nil {
		fmt.Printf("Error reading new file: %v\n", err)
		os.Exit(1)
	}

	newTables, err := parser.ParseSQLDump(string(newContent))
	if err != nil {
		fmt.Printf("Error parsing new SQL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Old schema: %d tables\n", len(oldTables))
	fmt.Printf("New schema: %d tables\n", len(newTables))
	fmt.Println()

	// Create maps for easy lookup
	oldTablesMap := make(map[string]*parser.CreateTableStatement)
	newTablesMap := make(map[string]*parser.CreateTableStatement)

	for _, table := range oldTables {
		oldTablesMap[table.TableName] = table
	}
	for _, table := range newTables {
		newTablesMap[table.TableName] = table
	}

	// Find all table names
	allTableNames := make(map[string]bool)
	for name := range oldTablesMap {
		allTableNames[name] = true
	}
	for name := range newTablesMap {
		allTableNames[name] = true
	}

	// Analyze differences for each table
	var diffs []*diff.TableDiff
	tablesAdded := 0
	tablesRemoved := 0
	tablesModified := 0

	for tableName := range allTableNames {
		oldTable, hasOld := oldTablesMap[tableName]
		newTable, hasNew := newTablesMap[tableName]

		if !hasOld {
			// Table added
			tablesAdded++
			fmt.Printf("+ TABLE ADDED: %s\n", tableName)
			continue
		}

		if !hasNew {
			// Table removed
			tablesRemoved++
			fmt.Printf("- TABLE REMOVED: %s\n", tableName)
			continue
		}

		// Table exists in both, analyze differences
		tableDiff := diff.CompareTables(oldTable, newTable)
		if tableDiff.HasChanges() {
			tablesModified++
			diffs = append(diffs, tableDiff)
			diff.PrintDiffSummary(tableDiff)
		} else {
			fmt.Printf("  TABLE UNCHANGED: %s\n", tableName)
		}
	}

	// Print overall summary
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("OVERALL SUMMARY")
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Printf("Tables added: %d\n", tablesAdded)
	fmt.Printf("Tables removed: %d\n", tablesRemoved)
	fmt.Printf("Tables modified: %d\n", tablesModified)
	fmt.Printf("Tables unchanged: %d\n", len(allTableNames)-tablesAdded-tablesRemoved-tablesModified)

	// Print detailed diffs if requested
	if len(diffs) > 0 {
		fmt.Printf("\nDetailed differences for %d modified tables:\n", len(diffs))
		for _, tableDiff := range diffs {
			diff.PrintTableDiff(tableDiff, true)
		}
	}

	// Exit with non-zero if there are changes
	if tablesAdded > 0 || tablesRemoved > 0 || tablesModified > 0 {
		os.Exit(1)
	}
}
