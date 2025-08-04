package diff

import (
	"fmt"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/output"
	"github.com/n0madic/mysql-diff/pkg/parser"
)

// PrintTableDiff prints a human-readable summary of table differences
func PrintTableDiff(diff *TableDiff, detailed bool) {
	fmt.Printf("\n%s\n", output.BoldText(strings.Repeat("=", 60)))
	fmt.Printf("TABLE DIFF: %s -> %s\n",
		output.ColorizeTableName(diff.OldTable.TableName),
		output.ColorizeTableName(diff.NewTable.TableName))
	fmt.Printf("%s\n", output.BoldText(strings.Repeat("=", 60)))

	if !diff.HasChanges() {
		fmt.Println("No changes detected.")
		return
	}

	// Table name change
	if diff.TableNameChanged {
		fmt.Printf("✏️  Table renamed: %s -> %s\n",
			output.ColorizeTableName(diff.OldTable.TableName),
			output.ColorizeTableName(diff.NewTable.TableName))
	}

	// Summary
	summary := diff.GetSummary()
	fmt.Printf("\n%s\n", output.BoldText("SUMMARY:"))
	fmt.Printf("  Columns: %s %s %s\n",
		output.GreenText(fmt.Sprintf("+%d", summary.Columns.Added)),
		output.RedText(fmt.Sprintf("-%d", summary.Columns.Removed)),
		output.YellowText(fmt.Sprintf("~%d", summary.Columns.Modified)))
	fmt.Printf("  Indexes: %s %s %s\n",
		output.GreenText(fmt.Sprintf("+%d", summary.Indexes.Added)),
		output.RedText(fmt.Sprintf("-%d", summary.Indexes.Removed)),
		output.YellowText(fmt.Sprintf("~%d", summary.Indexes.Modified)))
	fmt.Printf("  Foreign Keys: %s %s %s\n",
		output.GreenText(fmt.Sprintf("+%d", summary.ForeignKeys.Added)),
		output.RedText(fmt.Sprintf("-%d", summary.ForeignKeys.Removed)),
		output.YellowText(fmt.Sprintf("~%d", summary.ForeignKeys.Modified)))

	if summary.PrimaryKeyChanged {
		fmt.Printf("  Primary Key: %s\n", output.YellowText("CHANGED"))
	}
	if summary.TableOptionsChanged {
		fmt.Printf("  Table Options: %s\n", output.YellowText("CHANGED"))
	}
	if summary.PartitioningChanged {
		fmt.Printf("  Partitioning: %s\n", output.YellowText("CHANGED"))
	}

	if !detailed {
		return
	}

	// Detailed changes
	if len(diff.ColumnDiffs) > 0 {
		fmt.Printf("\n%s\n", output.BoldText("COLUMN CHANGES:"))
		for _, colDiff := range diff.ColumnDiffs {
			switch colDiff.ChangeType {
			case ChangeTypeAdded:
				fmt.Printf("  %s %s: %s\n",
					output.GreenText("+"),
					output.ColorizeColumnName(colDiff.Name),
					formatColumn(colDiff.NewColumn))
			case ChangeTypeRemoved:
				fmt.Printf("  %s %s: %s\n",
					output.RedText("-"),
					output.ColorizeColumnName(colDiff.Name),
					formatColumn(colDiff.OldColumn))
			case ChangeTypeModified:
				fmt.Printf("  %s %s:\n",
					output.YellowText("~"),
					output.ColorizeColumnName(colDiff.Name))
				printColumnChanges(colDiff.Changes)
			}
		}
	}

	if len(diff.IndexDiffs) > 0 {
		fmt.Println("\nINDEX CHANGES:")
		for _, idxDiff := range diff.IndexDiffs {
			switch idxDiff.ChangeType {
			case ChangeTypeAdded:
				fmt.Printf("  + %s\n", formatIndex(idxDiff.NewIndex))
			case ChangeTypeRemoved:
				fmt.Printf("  - %s\n", formatIndex(idxDiff.OldIndex))
			case ChangeTypeModified:
				fmt.Printf("  ~ %s:\n", formatIndex(idxDiff.OldIndex))
				printIndexChanges(idxDiff.Changes)
			}
		}
	}

	if len(diff.ForeignKeyDiffs) > 0 {
		fmt.Println("\nFOREIGN KEY CHANGES:")
		for _, fkDiff := range diff.ForeignKeyDiffs {
			switch fkDiff.ChangeType {
			case ChangeTypeAdded:
				fmt.Printf("  + %s\n", formatForeignKey(fkDiff.NewFK))
			case ChangeTypeRemoved:
				fmt.Printf("  - %s\n", formatForeignKey(fkDiff.OldFK))
			case ChangeTypeModified:
				fmt.Printf("  ~ %s:\n", formatForeignKey(fkDiff.OldFK))
				printForeignKeyChanges(fkDiff.Changes)
			}
		}
	}

	if diff.PrimaryKeyDiff != nil {
		fmt.Println("\nPRIMARY KEY CHANGES:")
		switch diff.PrimaryKeyDiff.ChangeType {
		case ChangeTypeAdded:
			fmt.Printf("  + %s\n", formatPrimaryKey(diff.PrimaryKeyDiff.NewPK))
		case ChangeTypeRemoved:
			fmt.Printf("  - %s\n", formatPrimaryKey(diff.PrimaryKeyDiff.OldPK))
		case ChangeTypeModified:
			fmt.Printf("  ~ %s:\n", formatPrimaryKey(diff.PrimaryKeyDiff.OldPK))
			printPrimaryKeyChanges(diff.PrimaryKeyDiff.Changes)
		}
	}

	if diff.TableOptionsDiff != nil {
		fmt.Println("\nTABLE OPTIONS CHANGES:")
		switch diff.TableOptionsDiff.ChangeType {
		case ChangeTypeAdded:
			fmt.Println("  + Table options added")
		case ChangeTypeRemoved:
			fmt.Println("  - Table options removed")
		case ChangeTypeModified:
			fmt.Println("  ~ Table options modified:")
			printTableOptionsChanges(diff.TableOptionsDiff.Changes)
		}
	}

	if diff.PartitionDiff != nil {
		fmt.Println("\nPARTITION CHANGES:")
		switch diff.PartitionDiff.ChangeType {
		case ChangeTypeAdded:
			fmt.Println("  + Partitioning added")
		case ChangeTypeRemoved:
			fmt.Println("  - Partitioning removed")
		case ChangeTypeModified:
			fmt.Println("  ~ Partitioning modified:")
			printPartitionChanges(diff.PartitionDiff.Changes)
		}
	}
}

// formatColumn formats column definition for display
func formatColumn(col *parser.ColumnDefinition) string {
	if col == nil {
		return ""
	}

	result := output.ColorizeDataType(col.DataType.Name)
	if len(col.DataType.Parameters) > 0 {
		result += fmt.Sprintf("(%s)", output.ColorizeNumber(strings.Join(col.DataType.Parameters, ",")))
	}
	if col.DataType.Unsigned {
		result += " " + output.BlueText("UNSIGNED")
	}
	if col.DataType.Zerofill {
		result += " " + output.BlueText("ZEROFILL")
	}
	if col.Nullable != nil && !*col.Nullable {
		result += " " + output.BlueText("NOT NULL")
	}
	if col.AutoIncrement {
		result += " " + output.BlueText("AUTO_INCREMENT")
	}
	if col.Unique {
		result += " " + output.BlueText("UNIQUE")
	}
	if col.PrimaryKey {
		result += " " + output.BlueText("PRIMARY KEY")
	}
	if col.DefaultValue != nil {
		result += fmt.Sprintf(" %s %s", output.BlueText("DEFAULT"), output.ColorizeString(*col.DefaultValue))
	}
	if col.Comment != nil {
		result += fmt.Sprintf(" %s %s", output.BlueText("COMMENT"), output.ColorizeString("'"+*col.Comment+"'"))
	}
	return result
}

// formatIndex formats index definition for display
func formatIndex(idx *parser.IndexDefinition) string {
	if idx == nil {
		return ""
	}

	cols := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		cols[i] = col.Name
	}

	name := "UNNAMED"
	if idx.Name != nil {
		name = *idx.Name
	}

	return fmt.Sprintf("%s %s (%s)", idx.IndexType, name, strings.Join(cols, ", "))
}

// formatForeignKey formats foreign key definition for display
func formatForeignKey(fk *parser.ForeignKeyDefinition) string {
	if fk == nil {
		return ""
	}

	name := "UNNAMED"
	if fk.Name != nil {
		name = *fk.Name
	}

	result := fmt.Sprintf("FK %s: (%s) -> %s(%s)",
		name,
		strings.Join(fk.Columns, ", "),
		fk.Reference.TableName,
		strings.Join(fk.Reference.Columns, ", "))

	if fk.Reference.OnDelete != nil {
		result += fmt.Sprintf(" ON DELETE %s", *fk.Reference.OnDelete)
	}
	if fk.Reference.OnUpdate != nil {
		result += fmt.Sprintf(" ON UPDATE %s", *fk.Reference.OnUpdate)
	}

	return result
}

// formatPrimaryKey formats primary key definition for display
func formatPrimaryKey(pk *parser.PrimaryKeyDefinition) string {
	if pk == nil {
		return ""
	}

	cols := make([]string, len(pk.Columns))
	for i, col := range pk.Columns {
		cols[i] = col.Name
	}

	name := ""
	if pk.Name != nil {
		name = fmt.Sprintf(" %s", *pk.Name)
	}

	return fmt.Sprintf("PRIMARY KEY%s (%s)", name, strings.Join(cols, ", "))
}

// PrintDiffSummary prints a concise summary of changes
func PrintDiffSummary(diff *TableDiff) {
	if !diff.HasChanges() {
		fmt.Printf("Table %s: No changes\n", diff.OldTable.TableName)
		return
	}

	var changes []string

	if diff.ColumnsAdded > 0 {
		changes = append(changes, fmt.Sprintf("+%d cols", diff.ColumnsAdded))
	}
	if diff.ColumnsRemoved > 0 {
		changes = append(changes, fmt.Sprintf("-%d cols", diff.ColumnsRemoved))
	}
	if diff.ColumnsModified > 0 {
		changes = append(changes, fmt.Sprintf("~%d cols", diff.ColumnsModified))
	}

	if diff.IndexesAdded > 0 {
		changes = append(changes, fmt.Sprintf("+%d idx", diff.IndexesAdded))
	}
	if diff.IndexesRemoved > 0 {
		changes = append(changes, fmt.Sprintf("-%d idx", diff.IndexesRemoved))
	}
	if diff.IndexesModified > 0 {
		changes = append(changes, fmt.Sprintf("~%d idx", diff.IndexesModified))
	}

	if diff.ForeignKeysAdded > 0 {
		changes = append(changes, fmt.Sprintf("+%d fk", diff.ForeignKeysAdded))
	}
	if diff.ForeignKeysRemoved > 0 {
		changes = append(changes, fmt.Sprintf("-%d fk", diff.ForeignKeysRemoved))
	}
	if diff.ForeignKeysModified > 0 {
		changes = append(changes, fmt.Sprintf("~%d fk", diff.ForeignKeysModified))
	}

	if diff.PrimaryKeyDiff != nil {
		changes = append(changes, "pk changed")
	}
	if diff.TableOptionsDiff != nil {
		changes = append(changes, "options changed")
	}
	if diff.PartitionDiff != nil {
		changes = append(changes, "partitions changed")
	}

	fmt.Printf("Table %s: %s\n", diff.OldTable.TableName, strings.Join(changes, ", "))
}

// Helper functions for printing typed changes

func printColumnChanges(changes *ColumnChanges) {
	if changes.DataType != nil {
		fmt.Printf("      data_type: %v -> %v\n", changes.DataType.Old, changes.DataType.New)
	}
	if changes.Nullable != nil {
		fmt.Printf("      nullable: %v -> %v\n", changes.Nullable.Old, changes.Nullable.New)
	}
	if changes.DefaultValue != nil {
		fmt.Printf("      default_value: %v -> %v\n", changes.DefaultValue.Old, changes.DefaultValue.New)
	}
	if changes.AutoIncrement != nil {
		fmt.Printf("      auto_increment: %v -> %v\n", changes.AutoIncrement.Old, changes.AutoIncrement.New)
	}
	if changes.Unique != nil {
		fmt.Printf("      unique: %v -> %v\n", changes.Unique.Old, changes.Unique.New)
	}
	if changes.PrimaryKey != nil {
		fmt.Printf("      primary_key: %v -> %v\n", changes.PrimaryKey.Old, changes.PrimaryKey.New)
	}
	if changes.Comment != nil {
		fmt.Printf("      comment: %v -> %v\n", changes.Comment.Old, changes.Comment.New)
	}
	if changes.Collation != nil {
		fmt.Printf("      collation: %v -> %v\n", changes.Collation.Old, changes.Collation.New)
	}
	if changes.CharacterSet != nil {
		fmt.Printf("      character_set: %v -> %v\n", changes.CharacterSet.Old, changes.CharacterSet.New)
	}
	if changes.Visible != nil {
		fmt.Printf("      visible: %v -> %v\n", changes.Visible.Old, changes.Visible.New)
	}
	if changes.ColumnFormat != nil {
		fmt.Printf("      column_format: %v -> %v\n", changes.ColumnFormat.Old, changes.ColumnFormat.New)
	}
	if changes.Storage != nil {
		fmt.Printf("      storage: %v -> %v\n", changes.Storage.Old, changes.Storage.New)
	}
	if changes.Generated != nil {
		fmt.Printf("      generated: %v -> %v\n", changes.Generated.Old, changes.Generated.New)
	}
}

func printIndexChanges(changes *IndexChanges) {
	if changes.Name != nil {
		fmt.Printf("      name: %v -> %v\n", changes.Name.Old, changes.Name.New)
	}
	if changes.IndexType != nil {
		fmt.Printf("      index_type: %v -> %v\n", changes.IndexType.Old, changes.IndexType.New)
	}
	if changes.Columns != nil {
		fmt.Printf("      columns: %v -> %v\n", changes.Columns.Old, changes.Columns.New)
	}
	if changes.KeyBlockSize != nil {
		fmt.Printf("      key_block_size: %v -> %v\n", changes.KeyBlockSize.Old, changes.KeyBlockSize.New)
	}
	if changes.Using != nil {
		fmt.Printf("      using: %v -> %v\n", changes.Using.Old, changes.Using.New)
	}
	if changes.Comment != nil {
		fmt.Printf("      comment: %v -> %v\n", changes.Comment.Old, changes.Comment.New)
	}
	if changes.Visible != nil {
		fmt.Printf("      visible: %v -> %v\n", changes.Visible.Old, changes.Visible.New)
	}
	if changes.Parser != nil {
		fmt.Printf("      parser: %v -> %v\n", changes.Parser.Old, changes.Parser.New)
	}
	if changes.Algorithm != nil {
		fmt.Printf("      algorithm: %v -> %v\n", changes.Algorithm.Old, changes.Algorithm.New)
	}
	if changes.Lock != nil {
		fmt.Printf("      lock: %v -> %v\n", changes.Lock.Old, changes.Lock.New)
	}
	if changes.EngineAttribute != nil {
		fmt.Printf("      engine_attribute: %v -> %v\n", changes.EngineAttribute.Old, changes.EngineAttribute.New)
	}
}

func printForeignKeyChanges(changes *ForeignKeyChanges) {
	if changes.Name != nil {
		fmt.Printf("      name: %v -> %v\n", changes.Name.Old, changes.Name.New)
	}
	if changes.Columns != nil {
		fmt.Printf("      columns: %v -> %v\n", changes.Columns.Old, changes.Columns.New)
	}
	if changes.ReferenceTable != nil {
		fmt.Printf("      reference_table: %v -> %v\n", changes.ReferenceTable.Old, changes.ReferenceTable.New)
	}
	if changes.ReferenceColumns != nil {
		fmt.Printf("      reference_columns: %v -> %v\n", changes.ReferenceColumns.Old, changes.ReferenceColumns.New)
	}
	if changes.OnDelete != nil {
		fmt.Printf("      on_delete: %v -> %v\n", changes.OnDelete.Old, changes.OnDelete.New)
	}
	if changes.OnUpdate != nil {
		fmt.Printf("      on_update: %v -> %v\n", changes.OnUpdate.Old, changes.OnUpdate.New)
	}
}

func printPrimaryKeyChanges(changes *PrimaryKeyChanges) {
	if changes.Columns != nil {
		fmt.Printf("      columns: %v -> %v\n", changes.Columns.Old, changes.Columns.New)
	}
	if changes.Name != nil {
		fmt.Printf("      name: %v -> %v\n", changes.Name.Old, changes.Name.New)
	}
	if changes.Using != nil {
		fmt.Printf("      using: %v -> %v\n", changes.Using.Old, changes.Using.New)
	}
	if changes.Comment != nil {
		fmt.Printf("      comment: %v -> %v\n", changes.Comment.Old, changes.Comment.New)
	}
}

func printTableOptionsChanges(changes *TableOptionsChanges) {
	if changes.Engine != nil {
		fmt.Printf("      engine: %v -> %v\n", changes.Engine.Old, changes.Engine.New)
	}
	if changes.AutoIncrement != nil {
		fmt.Printf("      auto_increment: %v -> %v\n", changes.AutoIncrement.Old, changes.AutoIncrement.New)
	}
	if changes.CharacterSet != nil {
		fmt.Printf("      character_set: %v -> %v\n", changes.CharacterSet.Old, changes.CharacterSet.New)
	}
	if changes.Collate != nil {
		fmt.Printf("      collate: %v -> %v\n", changes.Collate.Old, changes.Collate.New)
	}
	if changes.Comment != nil {
		fmt.Printf("      comment: %v -> %v\n", changes.Comment.Old, changes.Comment.New)
	}
}

func printPartitionChanges(changes *PartitionChanges) {
	if changes.Type != nil {
		fmt.Printf("      type: %v -> %v\n", changes.Type.Old, changes.Type.New)
	}
	if changes.Linear != nil {
		fmt.Printf("      linear: %v -> %v\n", changes.Linear.Old, changes.Linear.New)
	}
	if changes.Expression != nil {
		fmt.Printf("      expression: %v -> %v\n", changes.Expression.Old, changes.Expression.New)
	}
	if changes.Columns != nil {
		fmt.Printf("      columns: %v -> %v\n", changes.Columns.Old, changes.Columns.New)
	}
	if changes.PartitionsCount != nil {
		fmt.Printf("      partitions_count: %v -> %v\n", changes.PartitionsCount.Old, changes.PartitionsCount.New)
	}
	if changes.PartitionDefinitions != nil {
		fmt.Printf("      partition_definitions: %v -> %v\n", changes.PartitionDefinitions.Old, changes.PartitionDefinitions.New)
	}
}
