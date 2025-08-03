package diff

import (
	"fmt"
	"slices"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/parser"
)

// TableDiffAnalyzer analyzes differences between two table structures
type TableDiffAnalyzer struct{}

// NewTableDiffAnalyzer creates a new analyzer instance
func NewTableDiffAnalyzer() *TableDiffAnalyzer {
	return &TableDiffAnalyzer{}
}

// CompareTables compares two table structures and returns a complete diff analysis
func (a *TableDiffAnalyzer) CompareTables(oldTable, newTable *parser.CreateTableStatement) *TableDiff {
	diff := &TableDiff{
		OldTable:        oldTable,
		NewTable:        newTable,
		ColumnDiffs:     []ColumnDiff{},
		IndexDiffs:      []IndexDiff{},
		ForeignKeyDiffs: []ForeignKeyDiff{},
	}

	// Check table name change
	diff.TableNameChanged = oldTable.TableName != newTable.TableName

	// Compare each component
	diff.ColumnDiffs = a.compareColumns(oldTable.Columns, newTable.Columns)
	diff.PrimaryKeyDiff = a.comparePrimaryKeys(oldTable.PrimaryKey, newTable.PrimaryKey)
	diff.IndexDiffs = a.compareIndexes(oldTable.Indexes, newTable.Indexes)
	diff.ForeignKeyDiffs = a.compareForeignKeys(oldTable.ForeignKeys, newTable.ForeignKeys)
	diff.TableOptionsDiff = a.compareTableOptions(oldTable.TableOptions, newTable.TableOptions)
	diff.PartitionDiff = a.comparePartitions(oldTable.PartitionOptions, newTable.PartitionOptions)

	// Update counters
	a.updateCounters(diff)

	return diff
}

// compareColumns compares column definitions between old and new tables
func (a *TableDiffAnalyzer) compareColumns(oldColumns, newColumns []parser.ColumnDefinition) []ColumnDiff {
	var diffs []ColumnDiff

	// Create maps for easy lookup
	oldColsMap := make(map[string]parser.ColumnDefinition)
	newColsMap := make(map[string]parser.ColumnDefinition)

	for _, col := range oldColumns {
		oldColsMap[col.Name] = col
	}
	for _, col := range newColumns {
		newColsMap[col.Name] = col
	}

	// Find all column names
	allColumnNames := make(map[string]bool)
	for name := range oldColsMap {
		allColumnNames[name] = true
	}
	for name := range newColsMap {
		allColumnNames[name] = true
	}

	for colName := range allColumnNames {
		oldCol, hasOld := oldColsMap[colName]
		newCol, hasNew := newColsMap[colName]

		if !hasOld {
			// Column added
			diffs = append(diffs, ColumnDiff{
				Name:       colName,
				ChangeType: ChangeTypeAdded,
				NewColumn:  &newCol,
				Changes:    &ColumnChanges{},
			})
		} else if !hasNew {
			// Column removed
			diffs = append(diffs, ColumnDiff{
				Name:       colName,
				ChangeType: ChangeTypeRemoved,
				OldColumn:  &oldCol,
				Changes:    &ColumnChanges{},
			})
		} else {
			// Column exists in both, check for changes
			changes := a.compareColumnDefinitions(oldCol, newCol)
			if changes.HasChanges() {
				diffs = append(diffs, ColumnDiff{
					Name:       colName,
					ChangeType: ChangeTypeModified,
					OldColumn:  &oldCol,
					NewColumn:  &newCol,
					Changes:    changes,
				})
			}
		}
	}

	return diffs
}

// compareColumnDefinitions compares two column definitions and returns changes
func (a *TableDiffAnalyzer) compareColumnDefinitions(oldCol, newCol parser.ColumnDefinition) *ColumnChanges {
	changes := &ColumnChanges{}

	// Compare data type
	if !a.dataTypesEqual(oldCol.DataType, newCol.DataType) {
		changes.DataType = &FieldChange[string]{
			Old: a.dataTypeToString(oldCol.DataType),
			New: a.dataTypeToString(newCol.DataType),
		}
	}

	// Compare nullable
	if !ptrEqual(oldCol.Nullable, newCol.Nullable) {
		changes.Nullable = &FieldChange[any]{
			Old: ptrToValue(oldCol.Nullable),
			New: ptrToValue(newCol.Nullable),
		}
	}

	// Compare default value
	if !ptrEqual(oldCol.DefaultValue, newCol.DefaultValue) {
		changes.DefaultValue = &FieldChange[any]{
			Old: ptrToValue(oldCol.DefaultValue),
			New: ptrToValue(newCol.DefaultValue),
		}
	}

	// Compare boolean attributes
	if oldCol.AutoIncrement != newCol.AutoIncrement {
		changes.AutoIncrement = &FieldChange[bool]{
			Old: oldCol.AutoIncrement,
			New: newCol.AutoIncrement,
		}
	}

	if oldCol.Unique != newCol.Unique {
		changes.Unique = &FieldChange[bool]{
			Old: oldCol.Unique,
			New: newCol.Unique,
		}
	}

	if oldCol.PrimaryKey != newCol.PrimaryKey {
		changes.PrimaryKey = &FieldChange[bool]{
			Old: oldCol.PrimaryKey,
			New: newCol.PrimaryKey,
		}
	}

	// Compare string pointer attributes
	if !ptrEqual(oldCol.Comment, newCol.Comment) {
		changes.Comment = &FieldChange[any]{
			Old: ptrToValue(oldCol.Comment),
			New: ptrToValue(newCol.Comment),
		}
	}

	if !ptrEqual(oldCol.Collation, newCol.Collation) {
		changes.Collation = &FieldChange[any]{
			Old: ptrToValue(oldCol.Collation),
			New: ptrToValue(newCol.Collation),
		}
	}

	if !ptrEqual(oldCol.CharacterSet, newCol.CharacterSet) {
		changes.CharacterSet = &FieldChange[any]{
			Old: ptrToValue(oldCol.CharacterSet),
			New: ptrToValue(newCol.CharacterSet),
		}
	}

	if !ptrEqual(oldCol.Visible, newCol.Visible) {
		changes.Visible = &FieldChange[any]{
			Old: ptrToValue(oldCol.Visible),
			New: ptrToValue(newCol.Visible),
		}
	}

	if !ptrEqual(oldCol.ColumnFormat, newCol.ColumnFormat) {
		changes.ColumnFormat = &FieldChange[any]{
			Old: ptrToValue(oldCol.ColumnFormat),
			New: ptrToValue(newCol.ColumnFormat),
		}
	}

	if !ptrEqual(oldCol.Storage, newCol.Storage) {
		changes.Storage = &FieldChange[any]{
			Old: ptrToValue(oldCol.Storage),
			New: ptrToValue(newCol.Storage),
		}
	}

	// Compare generated columns
	if !generatedColumnEqual(oldCol.Generated, newCol.Generated) {
		changes.Generated = &FieldChange[*parser.GeneratedColumn]{
			Old: oldCol.Generated,
			New: newCol.Generated,
		}
	}

	return changes
}

// dataTypesEqual checks if two data types are equal
func (a *TableDiffAnalyzer) dataTypesEqual(oldDT, newDT parser.DataType) bool {
	return oldDT.Name == newDT.Name &&
		slices.Equal(oldDT.Parameters, newDT.Parameters) &&
		oldDT.Unsigned == newDT.Unsigned &&
		oldDT.Zerofill == newDT.Zerofill
}

// dataTypeToString converts DataType to string representation
func (a *TableDiffAnalyzer) dataTypeToString(dt parser.DataType) string {
	result := dt.Name
	if len(dt.Parameters) > 0 {
		result += fmt.Sprintf("(%s)", strings.Join(dt.Parameters, ","))
	}
	if dt.Unsigned {
		result += " UNSIGNED"
	}
	if dt.Zerofill {
		result += " ZEROFILL"
	}
	return result
}

// updateCounters updates summary counters in the diff object
func (a *TableDiffAnalyzer) updateCounters(diff *TableDiff) {
	// Count column changes
	for _, colDiff := range diff.ColumnDiffs {
		switch colDiff.ChangeType {
		case ChangeTypeAdded:
			diff.ColumnsAdded++
		case ChangeTypeRemoved:
			diff.ColumnsRemoved++
		case ChangeTypeModified:
			diff.ColumnsModified++
		}
	}

	// Count index changes
	for _, idxDiff := range diff.IndexDiffs {
		switch idxDiff.ChangeType {
		case ChangeTypeAdded:
			diff.IndexesAdded++
		case ChangeTypeRemoved:
			diff.IndexesRemoved++
		case ChangeTypeModified:
			diff.IndexesModified++
		}
	}

	// Count foreign key changes
	for _, fkDiff := range diff.ForeignKeyDiffs {
		switch fkDiff.ChangeType {
		case ChangeTypeAdded:
			diff.ForeignKeysAdded++
		case ChangeTypeRemoved:
			diff.ForeignKeysRemoved++
		case ChangeTypeModified:
			diff.ForeignKeysModified++
		}
	}

	// Update table-level flags
	diff.TableOptionsChanged = diff.TableOptionsDiff != nil
}

// CompareTables is a convenience function to compare two tables
func CompareTables(oldTable, newTable *parser.CreateTableStatement) *TableDiff {
	analyzer := NewTableDiffAnalyzer()
	return analyzer.CompareTables(oldTable, newTable)
}
