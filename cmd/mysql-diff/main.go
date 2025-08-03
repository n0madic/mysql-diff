package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/alter"
	"github.com/n0madic/mysql-diff/pkg/diff"
	"github.com/n0madic/mysql-diff/pkg/parser"
)

func main() {
	// Define command line flags
	verbose := flag.Bool("v", false, "Show verbose output with analysis details")
	verboseLong := flag.Bool("verbose", false, "Show verbose output with analysis details")
	includeDrops := flag.Bool("include-drops", false, "Include DROP TABLE statements for removed tables")
	includeCreates := flag.Bool("include-creates", false, "Include CREATE TABLE statements for new tables (as comments)")

	// New flags for enhanced functionality
	tableName := flag.String("table", "", "Compare only specific table")
	detailedMode := flag.Bool("detailed", false, "Output detailed diff report")
	jsonMode := flag.Bool("json", false, "Output results in JSON format")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MySQL Schema Diff Tool - Compare MySQL schemas and generate migration statements\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [OPTIONS] old_schema.sql new_schema.sql\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s old_schema.sql new_schema.sql                    # Generate ALTER statements\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --table users old_schema.sql new_schema.sql      # Compare only 'users' table\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --detailed old_schema.sql new_schema.sql         # Show detailed diff report\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json old_schema.sql new_schema.sql             # Output JSON format\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Output modes:\n")
		fmt.Fprintf(os.Stderr, "  default:           Generate ALTER statements for migration\n")
		fmt.Fprintf(os.Stderr, "  --detailed:        Human-readable diff report\n")
		fmt.Fprintf(os.Stderr, "  --json:            Structured JSON output for programmatic use\n")
	}

	flag.Parse()

	// Combine verbose flags
	isVerbose := *verbose || *verboseLong

	// Validate output mode flags
	modeCount := 0
	if *detailedMode {
		modeCount++
	}
	if *jsonMode {
		modeCount++
	}

	if modeCount > 1 {
		fmt.Fprintf(os.Stderr, "Error: Only one output mode can be specified (--detailed, or --json)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check arguments
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Error: Expected 2 arguments, got %d\n\n", flag.NArg())
		flag.Usage()
		os.Exit(1)
	}

	oldSchemaPath := flag.Arg(0)
	newSchemaPath := flag.Arg(1)

	// Read and parse old schema
	oldSQL, err := os.ReadFile(oldSchemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Old schema file '%s' not found\n", oldSchemaPath)
		os.Exit(1)
	}

	oldTables, err := parser.ParseSQLDump(string(oldSQL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing old schema: %v\n", err)
		os.Exit(1)
	}

	// Read and parse new schema
	newSQL, err := os.ReadFile(newSchemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: New schema file '%s' not found\n", newSchemaPath)
		os.Exit(1)
	}

	newTables, err := parser.ParseSQLDump(string(newSQL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing new schema: %v\n", err)
		os.Exit(1)
	}

	if isVerbose {
		fmt.Fprintf(os.Stderr, "-- Parsed %d tables from old schema\n", len(oldTables))
		fmt.Fprintf(os.Stderr, "-- Parsed %d tables from new schema\n", len(newTables))
	}

	// Filter tables by name if specified
	if *tableName != "" {
		oldTables = filterTablesByName(oldTables, *tableName)
		newTables = filterTablesByName(newTables, *tableName)

		if len(oldTables) == 0 && len(newTables) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Table '%s' not found in either schema\n", *tableName)
			os.Exit(1)
		}
	}

	// Match tables by name
	tableMatches := alter.MatchTablesByName(oldTables, newTables)

	// Process based on output mode
	if *jsonMode {
		handleJSONOutput(tableMatches, isVerbose)
		return
	}

	if *detailedMode {
		handleDetailedOutput(tableMatches, isVerbose)
		return
	}

	// Default: Generate ALTER statements
	generator := alter.NewStatementGenerator()
	allStatements := []string{}

	// Process table drops first (if requested)
	if *includeDrops {
		oldNames := make(map[string]bool)
		for _, table := range oldTables {
			oldNames[table.TableName] = true
		}
		newNames := make(map[string]bool)
		for _, table := range newTables {
			newNames[table.TableName] = true
		}
		dropStatements := alter.GenerateDropTableStatements(oldTables, newNames)
		allStatements = append(allStatements, dropStatements...)
	}

	// Process existing tables with changes
	analyzer := diff.NewTableDiffAnalyzer()
	for tableName, match := range tableMatches {
		if match.Old != nil && match.New != nil {
			// Table exists in both schemas, check for differences
			tableDiff := analyzer.CompareTables(match.Old, match.New)
			if tableDiff.HasChanges() {
				if isVerbose {
					fmt.Fprintf(os.Stderr, "-- Processing changes for table: %s\n", tableName)
				}
				statements := generator.GenerateAlterStatements(tableDiff)
				allStatements = append(allStatements, statements...)
			}
		}
	}

	// Process new tables (if requested)
	if *includeCreates {
		oldNames := make(map[string]bool)
		for _, table := range oldTables {
			oldNames[table.TableName] = true
		}
		createStatements := alter.GenerateCreateTableStatements(newTables, oldNames)
		allStatements = append(allStatements, createStatements...)
	}

	// Output results
	if len(allStatements) == 0 {
		if isVerbose {
			fmt.Fprintf(os.Stderr, "-- No differences found between schemas\n")
		}
		os.Exit(0)
	}

	// Print all ALTER statements
	for _, statement := range allStatements {
		fmt.Println(statement)
	}

	if isVerbose {
		fmt.Fprintf(os.Stderr, "-- Generated %d statements\n", len(allStatements))
	}
}

// filterTablesByName filters tables by name, returning only matching tables
func filterTablesByName(tables []*parser.CreateTableStatement, name string) []*parser.CreateTableStatement {
	var filtered []*parser.CreateTableStatement
	for _, table := range tables {
		if table.TableName == name {
			filtered = append(filtered, table)
		}
	}
	return filtered
}

// handleJSONOutput outputs results in JSON format
func handleJSONOutput(tableMatches map[string]struct {
	Old *parser.CreateTableStatement
	New *parser.CreateTableStatement
}, isVerbose bool) {
	analyzer := diff.NewTableDiffAnalyzer()
	results := make(map[string]*diff.TableDiff)

	for tableName, match := range tableMatches {
		if match.Old != nil && match.New != nil {
			tableDiff := analyzer.CompareTables(match.Old, match.New)
			if tableDiff.HasChanges() {
				results[tableName] = tableDiff
			}
		} else if match.Old != nil {
			// Table was removed
			tableDiff := &diff.TableDiff{
				OldTable: match.Old,
				NewTable: nil,
			}
			results[tableName] = tableDiff
		} else if match.New != nil {
			// Table was added
			tableDiff := &diff.TableDiff{
				OldTable: nil,
				NewTable: match.New,
			}
			results[tableName] = tableDiff
		}
	}

	jsonOutput, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))

	if isVerbose {
		fmt.Fprintf(os.Stderr, "-- Generated JSON output for %d tables\n", len(results))
	}
}

// handleDetailedOutput outputs human-readable detailed diff reports
func handleDetailedOutput(tableMatches map[string]struct {
	Old *parser.CreateTableStatement
	New *parser.CreateTableStatement
}, isVerbose bool) {
	analyzer := diff.NewTableDiffAnalyzer()
	hasAnyChanges := false

	for tableName, match := range tableMatches {
		if match.Old != nil && match.New != nil {
			tableDiff := analyzer.CompareTables(match.Old, match.New)
			if tableDiff.HasChanges() {
				hasAnyChanges = true
				diff.PrintTableDiff(tableDiff, true) // detailed=true
			}
		} else if match.Old != nil {
			// Table was removed
			hasAnyChanges = true
			fmt.Printf("\n%s\n", strings.Repeat("=", 60))
			fmt.Printf("TABLE REMOVED: %s\n", tableName)
			fmt.Printf("%s\n", strings.Repeat("=", 60))
			fmt.Printf("❌ Table '%s' was removed from the schema\n", tableName)
		} else if match.New != nil {
			// Table was added
			hasAnyChanges = true
			fmt.Printf("\n%s\n", strings.Repeat("=", 60))
			fmt.Printf("TABLE ADDED: %s\n", tableName)
			fmt.Printf("%s\n", strings.Repeat("=", 60))
			fmt.Printf("✅ Table '%s' was added to the schema\n", tableName)
		}
	}

	if !hasAnyChanges {
		fmt.Println("No differences found between schemas.")
		if isVerbose {
			fmt.Fprintf(os.Stderr, "-- Compared %d tables, no changes detected\n", len(tableMatches))
		}
	} else {
		// Print overall summary
		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("SUMMARY\n")
		fmt.Printf("%s\n", strings.Repeat("=", 60))

		totalTables := 0
		tablesWithChanges := 0
		for _, match := range tableMatches {
			if match.Old != nil && match.New != nil {
				totalTables++
				tableDiff := analyzer.CompareTables(match.Old, match.New)
				if tableDiff.HasChanges() {
					tablesWithChanges++
				}
			}
		}

		fmt.Printf("Tables analyzed: %d\n", totalTables)
		fmt.Printf("Tables with changes: %d\n", tablesWithChanges)

		if isVerbose {
			fmt.Fprintf(os.Stderr, "-- Detailed analysis complete\n")
		}
	}
}
