package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/parser"
)

// ParseResult holds the results of parsing a single file
type ParseResult struct {
	File    string
	Size    int64
	Tables  []*parser.CreateTableStatement
	Success bool
	Error   string
}

func printTableInfo(table *parser.CreateTableStatement, index int) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Table #%d: %s\n", index, table.TableName)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	if table.Temporary {
		fmt.Println("Type: TEMPORARY TABLE")
	}
	if table.IfNotExists {
		fmt.Println("IF NOT EXISTS: Yes")
	}

	// Columns
	if len(table.Columns) > 0 {
		fmt.Printf("\nColumns (%d):\n", len(table.Columns))
		for _, col := range table.Columns {
			colInfo := fmt.Sprintf("  - %s: %s", col.Name, col.DataType.Name)

			if len(col.DataType.Parameters) > 0 {
				colInfo += fmt.Sprintf("(%s)", strings.Join(col.DataType.Parameters, ", "))
			}

			if col.DataType.Unsigned {
				colInfo += " UNSIGNED"
			}
			if col.DataType.Zerofill {
				colInfo += " ZEROFILL"
			}

			var attributes []string
			if col.Nullable != nil && !*col.Nullable {
				attributes = append(attributes, "NOT NULL")
			} else if col.Nullable != nil && *col.Nullable {
				attributes = append(attributes, "NULL")
			}
			if col.AutoIncrement {
				attributes = append(attributes, "AUTO_INCREMENT")
			}
			if col.PrimaryKey {
				attributes = append(attributes, "PRIMARY KEY")
			}
			if col.Unique {
				attributes = append(attributes, "UNIQUE")
			}
			if col.DefaultValue != nil {
				attributes = append(attributes, fmt.Sprintf("DEFAULT %s", *col.DefaultValue))
			}
			if col.Comment != nil {
				attributes = append(attributes, fmt.Sprintf("COMMENT '%s'", *col.Comment))
			}
			if col.Generated != nil {
				attributes = append(attributes, fmt.Sprintf("GENERATED %s AS (%s)", col.Generated.Type, col.Generated.Expression))
			}
			if col.CharacterSet != nil {
				attributes = append(attributes, fmt.Sprintf("CHARACTER SET %s", *col.CharacterSet))
			}
			if col.Collation != nil {
				attributes = append(attributes, fmt.Sprintf("COLLATE %s", *col.Collation))
			}
			if col.ColumnFormat != nil {
				attributes = append(attributes, fmt.Sprintf("COLUMN_FORMAT %s", *col.ColumnFormat))
			}
			if col.Storage != nil {
				attributes = append(attributes, fmt.Sprintf("STORAGE %s", *col.Storage))
			}

			if len(attributes) > 0 {
				colInfo += fmt.Sprintf(" [%s]", strings.Join(attributes, ", "))
			}

			fmt.Println(colInfo)
		}
	}

	// Primary key
	if table.PrimaryKey != nil {
		var pkColumnStrs []string
		for _, col := range table.PrimaryKey.Columns {
			colStr := col.Name
			if col.Length != nil {
				colStr = fmt.Sprintf("%s(%d)", col.Name, *col.Length)
			}
			if col.Direction != nil {
				colStr += fmt.Sprintf(" %s", *col.Direction)
			}
			pkColumnStrs = append(pkColumnStrs, colStr)
		}

		pkInfo := fmt.Sprintf("Primary Key: (%s)", strings.Join(pkColumnStrs, ", "))
		if table.PrimaryKey.Name != nil {
			pkInfo = fmt.Sprintf("Primary Key %s: (%s)", *table.PrimaryKey.Name, strings.Join(pkColumnStrs, ", "))
		}
		fmt.Printf("\n%s\n", pkInfo)
	}

	// Indexes
	if len(table.Indexes) > 0 {
		fmt.Printf("\nIndexes (%d):\n", len(table.Indexes))
		for _, idx := range table.Indexes {
			idxInfo := fmt.Sprintf("  - %s", idx.IndexType)
			if idx.Name != nil {
				idxInfo += fmt.Sprintf(" %s", *idx.Name)
			}

			// Format columns with their details
			var columnStrs []string
			for _, col := range idx.Columns {
				colStr := col.Name
				if col.Length != nil {
					colStr = fmt.Sprintf("%s(%d)", col.Name, *col.Length)
				}
				if col.Direction != nil {
					colStr += fmt.Sprintf(" %s", *col.Direction)
				}
				columnStrs = append(columnStrs, colStr)
			}

			idxInfo += fmt.Sprintf(": (%s)", strings.Join(columnStrs, ", "))

			// Add index options
			var options []string
			if idx.Using != nil {
				options = append(options, fmt.Sprintf("USING %s", *idx.Using))
			}
			if idx.KeyBlockSize != nil {
				options = append(options, fmt.Sprintf("KEY_BLOCK_SIZE=%d", *idx.KeyBlockSize))
			}
			if idx.Comment != nil {
				options = append(options, fmt.Sprintf("COMMENT '%s'", *idx.Comment))
			}
			if idx.Visible != nil {
				if *idx.Visible {
					options = append(options, "VISIBLE")
				} else {
					options = append(options, "INVISIBLE")
				}
			}
			if idx.Parser != nil {
				options = append(options, fmt.Sprintf("WITH PARSER %s", *idx.Parser))
			}
			if idx.Algorithm != nil {
				options = append(options, fmt.Sprintf("ALGORITHM=%s", *idx.Algorithm))
			}
			if idx.Lock != nil {
				options = append(options, fmt.Sprintf("LOCK=%s", *idx.Lock))
			}
			if idx.EngineAttribute != nil {
				options = append(options, fmt.Sprintf("ENGINE_ATTRIBUTE='%s'", *idx.EngineAttribute))
			}

			if len(options) > 0 {
				idxInfo += fmt.Sprintf(" [%s]", strings.Join(options, ", "))
			}

			fmt.Println(idxInfo)
		}
	}

	// Foreign keys
	if len(table.ForeignKeys) > 0 {
		fmt.Printf("\nForeign Keys (%d):\n", len(table.ForeignKeys))
		for _, fk := range table.ForeignKeys {
			fkInfo := "  - "
			if fk.Name != nil {
				fkInfo += fmt.Sprintf("%s: ", *fk.Name)
			}
			fkInfo += fmt.Sprintf("(%s) -> %s(%s)",
				strings.Join(fk.Columns, ", "),
				fk.Reference.TableName,
				strings.Join(fk.Reference.Columns, ", "))
			if fk.Reference.OnDelete != nil {
				fkInfo += fmt.Sprintf(" ON DELETE %s", *fk.Reference.OnDelete)
			}
			if fk.Reference.OnUpdate != nil {
				fkInfo += fmt.Sprintf(" ON UPDATE %s", *fk.Reference.OnUpdate)
			}
			fmt.Println(fkInfo)
		}
	}

	// Table options
	if table.TableOptions != nil {
		fmt.Println("\nTable Options:")
		if table.TableOptions.Engine != nil {
			fmt.Printf("  - ENGINE: %s\n", *table.TableOptions.Engine)
		}
		if table.TableOptions.CharacterSet != nil {
			fmt.Printf("  - CHARACTER SET: %s\n", *table.TableOptions.CharacterSet)
		}
		if table.TableOptions.Collate != nil {
			fmt.Printf("  - COLLATE: %s\n", *table.TableOptions.Collate)
		}
		if table.TableOptions.Comment != nil {
			fmt.Printf("  - COMMENT: %s\n", *table.TableOptions.Comment)
		}
		if table.TableOptions.AutoIncrement != nil {
			fmt.Printf("  - AUTO_INCREMENT: %d\n", *table.TableOptions.AutoIncrement)
		}
		if table.TableOptions.KeyBlockSize != nil {
			fmt.Printf("  - KEY_BLOCK_SIZE: %d\n", *table.TableOptions.KeyBlockSize)
		}
		if table.TableOptions.MaxRows != nil {
			fmt.Printf("  - MAX_ROWS: %d\n", *table.TableOptions.MaxRows)
		}
		if table.TableOptions.MinRows != nil {
			fmt.Printf("  - MIN_ROWS: %d\n", *table.TableOptions.MinRows)
		}
		if table.TableOptions.Compression != nil {
			fmt.Printf("  - COMPRESSION: %s\n", *table.TableOptions.Compression)
		}
		if table.TableOptions.Encryption != nil {
			fmt.Printf("  - ENCRYPTION: %s\n", *table.TableOptions.Encryption)
		}
		if table.TableOptions.DataDirectory != nil {
			fmt.Printf("  - DATA DIRECTORY: %s\n", *table.TableOptions.DataDirectory)
		}
		if table.TableOptions.IndexDirectory != nil {
			fmt.Printf("  - INDEX DIRECTORY: %s\n", *table.TableOptions.IndexDirectory)
		}
		if table.TableOptions.Tablespace != nil {
			fmt.Printf("  - TABLESPACE: %s\n", *table.TableOptions.Tablespace)
		}
		if table.TableOptions.RowFormat != nil {
			fmt.Printf("  - ROW_FORMAT: %s\n", *table.TableOptions.RowFormat)
		}
		if table.TableOptions.StatsPersistent != nil {
			fmt.Printf("  - STATS_PERSISTENT: %d\n", *table.TableOptions.StatsPersistent)
		}
		if table.TableOptions.StatsAutoRecalc != nil {
			fmt.Printf("  - STATS_AUTO_RECALC: %d\n", *table.TableOptions.StatsAutoRecalc)
		}
		if table.TableOptions.StatsSamplePages != nil {
			fmt.Printf("  - STATS_SAMPLE_PAGES: %d\n", *table.TableOptions.StatsSamplePages)
		}
		if table.TableOptions.PackKeys != nil {
			fmt.Printf("  - PACK_KEYS: %d\n", *table.TableOptions.PackKeys)
		}
		if table.TableOptions.Checksum != nil {
			fmt.Printf("  - CHECKSUM: %d\n", *table.TableOptions.Checksum)
		}
		if table.TableOptions.DelayKeyWrite != nil {
			fmt.Printf("  - DELAY_KEY_WRITE: %d\n", *table.TableOptions.DelayKeyWrite)
		}
		if len(table.TableOptions.Union) > 0 {
			fmt.Printf("  - UNION: (%s)\n", strings.Join(table.TableOptions.Union, ", "))
		}
		if table.TableOptions.InsertMethod != nil {
			fmt.Printf("  - INSERT_METHOD: %s\n", *table.TableOptions.InsertMethod)
		}
	}

	// Partitioning information
	if table.PartitionOptions != nil {
		fmt.Println("\nPartitioning:")
		partType := table.PartitionOptions.Type
		if table.PartitionOptions.Linear {
			partType = fmt.Sprintf("LINEAR %s", partType)
		}
		fmt.Printf("  - Type: %s\n", partType)

		if table.PartitionOptions.Expression != nil {
			fmt.Printf("  - Expression: %s\n", *table.PartitionOptions.Expression)
		}
		if len(table.PartitionOptions.Columns) > 0 {
			fmt.Printf("  - Columns: (%s)\n", strings.Join(table.PartitionOptions.Columns, ", "))
		}
		if table.PartitionOptions.PartitionCount != nil {
			fmt.Printf("  - Partitions: %d\n", *table.PartitionOptions.PartitionCount)
		}

		if len(table.PartitionOptions.Partitions) > 0 {
			fmt.Printf("  - Partition definitions: %d defined\n", len(table.PartitionOptions.Partitions))
		}
	}
}

func parseSingleFile(dumpFile string) ParseResult {
	fmt.Printf("\nReading SQL dump from: %s\n", dumpFile)

	// Read file
	content, err := os.ReadFile(dumpFile)
	if err != nil {
		if os.IsNotExist(err) {
			errorMsg := fmt.Sprintf("File '%s' not found", dumpFile)
			fmt.Printf("Error: %s\n", errorMsg)
			return ParseResult{
				File:    dumpFile,
				Size:    0,
				Tables:  nil,
				Success: false,
				Error:   errorMsg,
			}
		}

		errorMsg := fmt.Sprintf("Error reading file: %v", err)
		fmt.Printf("Error: %s\n", errorMsg)
		return ParseResult{
			File:    dumpFile,
			Size:    0,
			Tables:  nil,
			Success: false,
			Error:   errorMsg,
		}
	}

	sqlContent := string(content)
	fileSize := int64(len(content))

	fmt.Printf("File size: %d bytes\n", fileSize)
	fmt.Println("Parsing SQL dump...")

	// Parse the dump
	tables, err := parser.ParseSQLDump(sqlContent)
	if err != nil {
		errorMsg := fmt.Sprintf("SQL parsing failed: %v", err)
		fmt.Printf("Error: %s\n", errorMsg)

		// Try to provide context for parsing errors
		errorStr := err.Error()

		// Simple line/column extraction (would need more sophisticated parsing for full error context)
		if strings.Contains(errorStr, "line") {
			fmt.Printf("\nError details: %s\n", errorStr)
		}

		fmt.Printf("\nFatal error: Stopping execution due to SQL parsing errors in '%s'\n", dumpFile)
		os.Exit(1)
	}

	fmt.Printf("Found %d CREATE TABLE statements\n", len(tables))

	return ParseResult{
		File:    dumpFile,
		Size:    fileSize,
		Tables:  tables,
		Success: true,
		Error:   "",
	}
}

func printAggregatedSummary(results []ParseResult) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("AGGREGATED SUMMARY")
	fmt.Printf("%s\n", strings.Repeat("=", 80))

	totalFiles := len(results)
	fmt.Printf("Files processed: %d\n", totalFiles)
	fmt.Printf("Successfully parsed: %d\n", totalFiles)

	// Aggregate statistics from all files
	var allTables []*parser.CreateTableStatement
	var totalFileSize int64

	for _, result := range results {
		allTables = append(allTables, result.Tables...)
		totalFileSize += result.Size
	}

	fmt.Printf("\nTotal file size: %s bytes\n", addCommas(totalFileSize))
	fmt.Printf("Total tables found: %d\n", len(allTables))

	// Statistics
	totalColumns := 0
	totalIndexes := 0
	totalForeignKeys := 0

	for _, table := range allTables {
		totalColumns += len(table.Columns)
		totalIndexes += len(table.Indexes)
		totalForeignKeys += len(table.ForeignKeys)
	}

	fmt.Println("\nAggregated Statistics:")
	fmt.Printf("  - Total columns: %d\n", totalColumns)
	fmt.Printf("  - Total indexes: %d\n", totalIndexes)
	fmt.Printf("  - Total foreign keys: %d\n", totalForeignKeys)
	if len(allTables) > 0 {
		fmt.Printf("  - Average columns per table: %.1f\n", float64(totalColumns)/float64(len(allTables)))
	}

	// Engines used across all files
	engines := make(map[string]bool)
	for _, table := range allTables {
		if table.TableOptions != nil && table.TableOptions.Engine != nil {
			engines[*table.TableOptions.Engine] = true
		}
	}

	if len(engines) > 0 {
		var engineList []string
		for engine := range engines {
			engineList = append(engineList, engine)
		}
		sort.Strings(engineList)
		fmt.Printf("\nStorage engines used: %s\n", strings.Join(engineList, ", "))
	}
}

func addCommas(n int64) string {
	str := strconv.FormatInt(n, 10)
	if len(str) < 4 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-dump-parser <sql_dump_file1> [sql_dump_file2] [...]")
		fmt.Println("Examples:")
		fmt.Println("  test-dump-parser auth.db.sql")
		fmt.Println("  test-dump-parser auth.db.sql users.sql products.sql")
		fmt.Println("  test-dump-parser *.sql")
		os.Exit(1)
	}

	dumpFiles := os.Args[1:]

	// Expand wildcards
	var expandedFiles []string
	for _, pattern := range dumpFiles {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Error expanding pattern '%s': %v\n", pattern, err)
			continue
		}
		if len(matches) == 0 {
			// If no matches found, treat as literal filename
			expandedFiles = append(expandedFiles, pattern)
		} else {
			expandedFiles = append(expandedFiles, matches...)
		}
	}

	// Remove duplicates and sort
	fileMap := make(map[string]bool)
	var uniqueFiles []string
	for _, file := range expandedFiles {
		if !fileMap[file] {
			fileMap[file] = true
			uniqueFiles = append(uniqueFiles, file)
		}
	}
	sort.Strings(uniqueFiles)
	dumpFiles = uniqueFiles

	fmt.Printf("Processing %d file(s)...\n", len(dumpFiles))

	var results []ParseResult

	// Process each file
	for _, dumpFile := range dumpFiles {
		result := parseSingleFile(dumpFile)
		results = append(results, result)

		// Print detailed information for each table in this file
		tables := result.Tables

		if len(dumpFiles) > 1 {
			fmt.Printf("\n%s\n", strings.Repeat("=", 80))
			fmt.Printf("DETAILED TABLES FROM: %s\n", dumpFile)
			fmt.Printf("%s\n", strings.Repeat("=", 80))
		}

		for i, table := range tables {
			printTableInfo(table, i+1)
		}

		// Print per-file summary
		if len(dumpFiles) > 1 {
			fmt.Printf("\n%s\n", strings.Repeat("=", 60))
			fmt.Printf("SUMMARY FOR: %s\n", dumpFile)
			fmt.Printf("%s\n", strings.Repeat("=", 60))
			fmt.Printf("Total tables parsed: %d\n", len(tables))

			if len(tables) > 0 {
				fmt.Println("\nTable names:")
				for _, table := range tables {
					fmt.Printf("  - %s\n", table.TableName)
				}

				// Statistics
				totalColumns := 0
				totalIndexes := 0
				totalForeignKeys := 0

				for _, table := range tables {
					totalColumns += len(table.Columns)
					totalIndexes += len(table.Indexes)
					totalForeignKeys += len(table.ForeignKeys)
				}

				fmt.Println("\nStatistics:")
				fmt.Printf("  - Total columns: %d\n", totalColumns)
				fmt.Printf("  - Total indexes: %d\n", totalIndexes)
				fmt.Printf("  - Total foreign keys: %d\n", totalForeignKeys)

				// Engines used
				engines := make(map[string]bool)
				for _, table := range tables {
					if table.TableOptions != nil && table.TableOptions.Engine != nil {
						engines[*table.TableOptions.Engine] = true
					}
				}

				if len(engines) > 0 {
					var engineList []string
					for engine := range engines {
						engineList = append(engineList, engine)
					}
					sort.Strings(engineList)
					fmt.Printf("  - Storage engines used: %s\n", strings.Join(engineList, ", "))
				}
			}
		}
	}

	// Print aggregated summary if multiple files
	if len(dumpFiles) > 1 {
		printAggregatedSummary(results)
	} else if len(dumpFiles) == 1 {
		// Single file summary
		result := results[0]
		tables := result.Tables

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Println("SUMMARY")
		fmt.Printf("%s\n", strings.Repeat("=", 60))
		fmt.Printf("Total tables parsed: %d\n", len(tables))

		if len(tables) > 0 {
			fmt.Println("\nTable names:")
			for _, table := range tables {
				fmt.Printf("  - %s\n", table.TableName)
			}

			// Statistics
			totalColumns := 0
			totalIndexes := 0
			totalForeignKeys := 0

			for _, table := range tables {
				totalColumns += len(table.Columns)
				totalIndexes += len(table.Indexes)
				totalForeignKeys += len(table.ForeignKeys)
			}

			fmt.Println("\nStatistics:")
			fmt.Printf("  - Total columns: %d\n", totalColumns)
			fmt.Printf("  - Total indexes: %d\n", totalIndexes)
			fmt.Printf("  - Total foreign keys: %d\n", totalForeignKeys)

			// Engines used
			engines := make(map[string]bool)
			for _, table := range tables {
				if table.TableOptions != nil && table.TableOptions.Engine != nil {
					engines[*table.TableOptions.Engine] = true
				}
			}

			if len(engines) > 0 {
				var engineList []string
				for engine := range engines {
					engineList = append(engineList, engine)
				}
				sort.Strings(engineList)
				fmt.Printf("\nStorage engines used: %s\n", strings.Join(engineList, ", "))
			}
		}
	}
}
