package alter

import (
	"fmt"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/diff"
	"github.com/n0madic/mysql-diff/pkg/parser"
)

// StatementGenerator generates ALTER TABLE statements from table differences
type StatementGenerator struct{}

// NewStatementGenerator creates a new ALTER statement generator
func NewStatementGenerator() *StatementGenerator {
	return &StatementGenerator{}
}

// GenerateAlterStatements generates all ALTER statements needed to transform old table to new table
func (g *StatementGenerator) GenerateAlterStatements(tableDiff *diff.TableDiff) []string {
	statements := []string{}
	tableName := tableDiff.OldTable.TableName

	// Handle table rename first if needed
	if tableDiff.TableNameChanged {
		statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`;", tableName, tableDiff.NewTable.TableName))
		tableName = tableDiff.NewTable.TableName // Use new name for subsequent operations
	}

	// Collect all column, index, and constraint changes
	alterClauses := []string{}

	// Process column changes
	alterClauses = append(alterClauses, g.generateColumnChanges(tableDiff)...)

	// Process primary key changes
	if tableDiff.PrimaryKeyDiff != nil {
		alterClauses = append(alterClauses, g.generatePrimaryKeyChanges(tableDiff.PrimaryKeyDiff)...)
	}

	// Process index changes
	alterClauses = append(alterClauses, g.generateIndexChanges(tableDiff)...)

	// Process foreign key changes
	alterClauses = append(alterClauses, g.generateForeignKeyChanges(tableDiff)...)

	// Generate main ALTER TABLE statement if there are changes
	if len(alterClauses) > 0 {
		alterStmt := fmt.Sprintf("ALTER TABLE `%s`\n  %s;", tableName, strings.Join(alterClauses, ",\n  "))
		statements = append(statements, alterStmt)
	}

	// Process table options changes (separate ALTER statement)
	if tableDiff.TableOptionsDiff != nil {
		tableOptionsStmt := g.generateTableOptionsChanges(tableName, tableDiff.TableOptionsDiff)
		if tableOptionsStmt != "" {
			statements = append(statements, tableOptionsStmt)
		}
	}

	// Process partitioning changes (separate ALTER statement)
	if tableDiff.PartitionDiff != nil {
		partitionStmt := g.generatePartitionChanges(tableName, tableDiff.PartitionDiff)
		if partitionStmt != "" {
			statements = append(statements, partitionStmt)
		}
	}

	return statements
}

func (g *StatementGenerator) generateColumnChanges(tableDiff *diff.TableDiff) []string {
	clauses := []string{}

	for _, colDiff := range tableDiff.ColumnDiffs {
		switch colDiff.ChangeType {
		case diff.ChangeTypeAdded:
			clauses = append(clauses, g.generateAddColumn(colDiff.NewColumn))
		case diff.ChangeTypeRemoved:
			clauses = append(clauses, fmt.Sprintf("DROP COLUMN `%s`", colDiff.Name))
		case diff.ChangeTypeModified:
			clauses = append(clauses, g.generateModifyColumn(colDiff.NewColumn))
		}
	}

	return clauses
}

func (g *StatementGenerator) generateAddColumn(column *parser.ColumnDefinition) string {
	colDef := g.formatColumnDefinition(column)
	return fmt.Sprintf("ADD COLUMN %s", colDef)
}

func (g *StatementGenerator) generateModifyColumn(column *parser.ColumnDefinition) string {
	colDef := g.formatColumnDefinition(column)
	return fmt.Sprintf("MODIFY COLUMN %s", colDef)
}

func (g *StatementGenerator) formatColumnDefinition(column *parser.ColumnDefinition) string {
	parts := []string{fmt.Sprintf("`%s`", column.Name)}

	// Data type
	dataType := column.DataType.Name
	if len(column.DataType.Parameters) > 0 {
		dataType += fmt.Sprintf("(%s)", strings.Join(column.DataType.Parameters, ","))
	}
	if column.DataType.Unsigned {
		dataType += " UNSIGNED"
	}
	if column.DataType.Zerofill {
		dataType += " ZEROFILL"
	}
	parts = append(parts, dataType)

	// Character set and collation
	if column.CharacterSet != nil && *column.CharacterSet != "" {
		parts = append(parts, fmt.Sprintf("CHARACTER SET %s", *column.CharacterSet))
	}
	if column.Collation != nil && *column.Collation != "" {
		parts = append(parts, fmt.Sprintf("COLLATE %s", *column.Collation))
	}

	// NULL/NOT NULL
	if column.Nullable != nil {
		if *column.Nullable {
			parts = append(parts, "NULL")
		} else {
			parts = append(parts, "NOT NULL")
		}
	}

	// AUTO_INCREMENT
	if column.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// UNIQUE
	if column.Unique {
		parts = append(parts, "UNIQUE")
	}

	// PRIMARY KEY (column level)
	if column.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	// DEFAULT
	if column.DefaultValue != nil && *column.DefaultValue != "" {
		upperDefault := strings.ToUpper(*column.DefaultValue)
		if upperDefault == "CURRENT_TIMESTAMP" || upperDefault == "NULL" {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", *column.DefaultValue))
		} else {
			parts = append(parts, fmt.Sprintf("DEFAULT '%s'", *column.DefaultValue))
		}
	}

	// GENERATED column
	if column.Generated != nil {
		expr := column.Generated.Expression
		genType := column.Generated.Type
		if genType == "" {
			genType = "VIRTUAL"
		}
		parts = append(parts, fmt.Sprintf("GENERATED ALWAYS AS (%s) %s", expr, genType))
	}

	// VISIBLE/INVISIBLE
	if column.Visible != nil {
		if *column.Visible {
			parts = append(parts, "VISIBLE")
		} else {
			parts = append(parts, "INVISIBLE")
		}
	}

	// COMMENT
	if column.Comment != nil && *column.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", *column.Comment))
	}

	// COLUMN_FORMAT
	if column.ColumnFormat != nil && *column.ColumnFormat != "" {
		parts = append(parts, fmt.Sprintf("COLUMN_FORMAT %s", *column.ColumnFormat))
	}

	// STORAGE
	if column.Storage != nil && *column.Storage != "" {
		parts = append(parts, fmt.Sprintf("STORAGE %s", *column.Storage))
	}

	return strings.Join(parts, " ")
}

func (g *StatementGenerator) generatePrimaryKeyChanges(pkDiff *diff.PrimaryKeyDiff) []string {
	clauses := []string{}

	switch pkDiff.ChangeType {
	case diff.ChangeTypeRemoved:
		clauses = append(clauses, "DROP PRIMARY KEY")
	case diff.ChangeTypeAdded:
		pkDef := g.formatPrimaryKeyDefinition(pkDiff.NewPK)
		clauses = append(clauses, fmt.Sprintf("ADD %s", pkDef))
	case diff.ChangeTypeModified:
		// Drop and recreate
		clauses = append(clauses, "DROP PRIMARY KEY")
		pkDef := g.formatPrimaryKeyDefinition(pkDiff.NewPK)
		clauses = append(clauses, fmt.Sprintf("ADD %s", pkDef))
	}

	return clauses
}

func (g *StatementGenerator) formatPrimaryKeyDefinition(pk *parser.PrimaryKeyDefinition) string {
	columns := []string{}
	for _, col := range pk.Columns {
		columns = append(columns, fmt.Sprintf("`%s`", col.Name))
	}
	colList := strings.Join(columns, ", ")

	if pk.Name != nil && *pk.Name != "" {
		return fmt.Sprintf("CONSTRAINT `%s` PRIMARY KEY (%s)", *pk.Name, colList)
	}
	return fmt.Sprintf("PRIMARY KEY (%s)", colList)
}

func (g *StatementGenerator) generateIndexChanges(tableDiff *diff.TableDiff) []string {
	clauses := []string{}

	for _, idxDiff := range tableDiff.IndexDiffs {
		switch idxDiff.ChangeType {
		case diff.ChangeTypeRemoved:
			if idxDiff.OldIndex.Name != nil && *idxDiff.OldIndex.Name != "" {
				clauses = append(clauses, fmt.Sprintf("DROP INDEX `%s`", *idxDiff.OldIndex.Name))
			} else {
				// For unnamed indexes, we need to identify by columns
				cols := []string{}
				for _, col := range idxDiff.OldIndex.Columns {
					cols = append(cols, fmt.Sprintf("`%s`", col.Name))
				}
				colList := strings.Join(cols, ", ")
				clauses = append(clauses, fmt.Sprintf("DROP INDEX (%s)", colList))
			}

		case diff.ChangeTypeAdded:
			idxDef := g.formatIndexDefinition(idxDiff.NewIndex)
			clauses = append(clauses, fmt.Sprintf("ADD %s", idxDef))

		case diff.ChangeTypeModified:
			// Drop old and add new
			if idxDiff.OldIndex.Name != nil && *idxDiff.OldIndex.Name != "" {
				clauses = append(clauses, fmt.Sprintf("DROP INDEX `%s`", *idxDiff.OldIndex.Name))
			}
			idxDef := g.formatIndexDefinition(idxDiff.NewIndex)
			clauses = append(clauses, fmt.Sprintf("ADD %s", idxDef))
		}
	}

	return clauses
}

func (g *StatementGenerator) formatIndexDefinition(idx *parser.IndexDefinition) string {
	parts := []string{}

	// Index type
	switch idx.IndexType {
	case "UNIQUE":
		parts = append(parts, "UNIQUE INDEX")
	case "FULLTEXT":
		parts = append(parts, "FULLTEXT INDEX")
	case "SPATIAL":
		parts = append(parts, "SPATIAL INDEX")
	default:
		parts = append(parts, "INDEX")
	}

	// Index name
	if idx.Name != nil && *idx.Name != "" {
		parts = append(parts, fmt.Sprintf("`%s`", *idx.Name))
	}

	// Columns
	colParts := []string{}
	for _, col := range idx.Columns {
		colPart := fmt.Sprintf("`%s`", col.Name)
		if col.Length != nil && *col.Length > 0 {
			colPart += fmt.Sprintf("(%d)", *col.Length)
		}
		if col.Direction != nil && *col.Direction != "" {
			colPart += fmt.Sprintf(" %s", *col.Direction)
		}
		colParts = append(colParts, colPart)
	}

	parts = append(parts, fmt.Sprintf("(%s)", strings.Join(colParts, ", ")))

	// Index options
	options := []string{}
	if idx.Using != nil && *idx.Using != "" {
		options = append(options, fmt.Sprintf("USING %s", *idx.Using))
	}
	if idx.KeyBlockSize != nil && *idx.KeyBlockSize > 0 {
		options = append(options, fmt.Sprintf("KEY_BLOCK_SIZE=%d", *idx.KeyBlockSize))
	}
	if idx.Parser != nil && *idx.Parser != "" {
		options = append(options, fmt.Sprintf("WITH PARSER %s", *idx.Parser))
	}
	if idx.Comment != nil && *idx.Comment != "" {
		options = append(options, fmt.Sprintf("COMMENT '%s'", *idx.Comment))
	}
	if idx.Visible != nil && !*idx.Visible {
		options = append(options, "INVISIBLE")
	}
	if idx.Algorithm != nil && *idx.Algorithm != "" {
		options = append(options, fmt.Sprintf("ALGORITHM=%s", *idx.Algorithm))
	}
	if idx.Lock != nil && *idx.Lock != "" {
		options = append(options, fmt.Sprintf("LOCK=%s", *idx.Lock))
	}
	if idx.EngineAttribute != nil && *idx.EngineAttribute != "" {
		options = append(options, fmt.Sprintf("ENGINE_ATTRIBUTE='%s'", *idx.EngineAttribute))
	}

	if len(options) > 0 {
		parts = append(parts, strings.Join(options, " "))
	}

	return strings.Join(parts, " ")
}

func (g *StatementGenerator) generateForeignKeyChanges(tableDiff *diff.TableDiff) []string {
	clauses := []string{}

	for _, fkDiff := range tableDiff.ForeignKeyDiffs {
		switch fkDiff.ChangeType {
		case diff.ChangeTypeRemoved:
			if fkDiff.OldFK.Name != nil && *fkDiff.OldFK.Name != "" {
				clauses = append(clauses, fmt.Sprintf("DROP FOREIGN KEY `%s`", *fkDiff.OldFK.Name))
			}
			// For unnamed FKs, MySQL requires a name, so we can't handle this case easily

		case diff.ChangeTypeAdded:
			fkDef := g.formatForeignKeyDefinition(fkDiff.NewFK)
			clauses = append(clauses, fmt.Sprintf("ADD %s", fkDef))

		case diff.ChangeTypeModified:
			// Drop old and add new
			if fkDiff.OldFK.Name != nil && *fkDiff.OldFK.Name != "" {
				clauses = append(clauses, fmt.Sprintf("DROP FOREIGN KEY `%s`", *fkDiff.OldFK.Name))
			}
			fkDef := g.formatForeignKeyDefinition(fkDiff.NewFK)
			clauses = append(clauses, fmt.Sprintf("ADD %s", fkDef))
		}
	}

	return clauses
}

func (g *StatementGenerator) formatForeignKeyDefinition(fk *parser.ForeignKeyDefinition) string {
	parts := []string{}

	if fk.Name != nil && *fk.Name != "" {
		parts = append(parts, fmt.Sprintf("CONSTRAINT `%s`", *fk.Name))
	}

	// Columns
	cols := []string{}
	for _, col := range fk.Columns {
		cols = append(cols, fmt.Sprintf("`%s`", col))
	}
	colList := strings.Join(cols, ", ")
	parts = append(parts, fmt.Sprintf("FOREIGN KEY (%s)", colList))

	// Reference
	refCols := []string{}
	for _, col := range fk.Reference.Columns {
		refCols = append(refCols, fmt.Sprintf("`%s`", col))
	}
	refColList := strings.Join(refCols, ", ")
	parts = append(parts, fmt.Sprintf("REFERENCES `%s` (%s)", fk.Reference.TableName, refColList))

	// Referential actions
	if fk.Reference.OnDelete != nil && *fk.Reference.OnDelete != "" {
		parts = append(parts, fmt.Sprintf("ON DELETE %s", *fk.Reference.OnDelete))
	}
	if fk.Reference.OnUpdate != nil && *fk.Reference.OnUpdate != "" {
		parts = append(parts, fmt.Sprintf("ON UPDATE %s", *fk.Reference.OnUpdate))
	}

	return strings.Join(parts, " ")
}

func (g *StatementGenerator) generateTableOptionsChanges(tableName string, optionsDiff *diff.TableOptionsDiff) string {
	if optionsDiff.ChangeType == diff.ChangeTypeRemoved {
		// Can't remove all table options, skip
		return ""
	}

	options := []string{}
	var opts *parser.TableOptions

	if optionsDiff.ChangeType == diff.ChangeTypeAdded {
		opts = optionsDiff.NewOptions
	} else { // MODIFIED
		opts = optionsDiff.NewOptions
	}

	// Build options list
	if opts.Engine != nil && *opts.Engine != "" {
		options = append(options, fmt.Sprintf("ENGINE=%s", *opts.Engine))
	}
	if opts.AutoIncrement != nil && *opts.AutoIncrement > 0 {
		options = append(options, fmt.Sprintf("AUTO_INCREMENT=%d", *opts.AutoIncrement))
	}
	if opts.CharacterSet != nil && *opts.CharacterSet != "" {
		options = append(options, fmt.Sprintf("DEFAULT CHARSET=%s", *opts.CharacterSet))
	}
	if opts.Collate != nil && *opts.Collate != "" {
		options = append(options, fmt.Sprintf("COLLATE=%s", *opts.Collate))
	}
	if opts.Comment != nil && *opts.Comment != "" {
		options = append(options, fmt.Sprintf("COMMENT='%s'", *opts.Comment))
	}
	if opts.RowFormat != nil && *opts.RowFormat != "" {
		options = append(options, fmt.Sprintf("ROW_FORMAT=%s", *opts.RowFormat))
	}
	if opts.KeyBlockSize != nil && *opts.KeyBlockSize > 0 {
		options = append(options, fmt.Sprintf("KEY_BLOCK_SIZE=%d", *opts.KeyBlockSize))
	}
	if opts.MaxRows != nil && *opts.MaxRows > 0 {
		options = append(options, fmt.Sprintf("MAX_ROWS=%d", *opts.MaxRows))
	}
	if opts.MinRows != nil && *opts.MinRows > 0 {
		options = append(options, fmt.Sprintf("MIN_ROWS=%d", *opts.MinRows))
	}
	if opts.Compression != nil && *opts.Compression != "" {
		options = append(options, fmt.Sprintf("COMPRESSION='%s'", *opts.Compression))
	}
	if opts.Encryption != nil && *opts.Encryption != "" {
		options = append(options, fmt.Sprintf("ENCRYPTION='%s'", *opts.Encryption))
	}
	if opts.StatsPersistent != nil && *opts.StatsPersistent != 0 {
		options = append(options, fmt.Sprintf("STATS_PERSISTENT=%d", *opts.StatsPersistent))
	}
	if opts.StatsAutoRecalc != nil && *opts.StatsAutoRecalc != 0 {
		options = append(options, fmt.Sprintf("STATS_AUTO_RECALC=%d", *opts.StatsAutoRecalc))
	}
	if opts.StatsSamplePages != nil && *opts.StatsSamplePages > 0 {
		options = append(options, fmt.Sprintf("STATS_SAMPLE_PAGES=%d", *opts.StatsSamplePages))
	}
	if opts.PackKeys != nil && *opts.PackKeys != 0 {
		options = append(options, fmt.Sprintf("PACK_KEYS=%d", *opts.PackKeys))
	}
	if opts.Checksum != nil && *opts.Checksum != 0 {
		options = append(options, fmt.Sprintf("CHECKSUM=%d", *opts.Checksum))
	}
	if opts.DelayKeyWrite != nil && *opts.DelayKeyWrite != 0 {
		options = append(options, fmt.Sprintf("DELAY_KEY_WRITE=%d", *opts.DelayKeyWrite))
	}

	if len(options) > 0 {
		return fmt.Sprintf("ALTER TABLE `%s` %s;", tableName, strings.Join(options, " "))
	}

	return ""
}

func (g *StatementGenerator) generatePartitionChanges(tableName string, partitionDiff *diff.PartitionDiff) string {
	switch partitionDiff.ChangeType {
	case diff.ChangeTypeRemoved:
		return fmt.Sprintf("ALTER TABLE `%s` REMOVE PARTITIONING;", tableName)

	case diff.ChangeTypeAdded:
		partitionDef := g.formatPartitionDefinition(partitionDiff.NewPartition)
		return fmt.Sprintf("ALTER TABLE `%s` %s;", tableName, partitionDef)

	case diff.ChangeTypeModified:
		// For simplicity, we'll remove and re-add partitioning
		partitionDef := g.formatPartitionDefinition(partitionDiff.NewPartition)
		return fmt.Sprintf("ALTER TABLE `%s` REMOVE PARTITIONING;\nALTER TABLE `%s` %s;", tableName, tableName, partitionDef)
	}

	return ""
}

func (g *StatementGenerator) formatPartitionDefinition(partitionOpts *parser.PartitionOptions) string {
	parts := []string{"PARTITION BY"}

	if partitionOpts.Linear {
		parts = append(parts, "LINEAR")
	}

	parts = append(parts, partitionOpts.Type)

	if partitionOpts.Expression != nil && *partitionOpts.Expression != "" {
		parts = append(parts, fmt.Sprintf("(%s)", *partitionOpts.Expression))
	} else if len(partitionOpts.Columns) > 0 {
		cols := []string{}
		for _, col := range partitionOpts.Columns {
			cols = append(cols, fmt.Sprintf("`%s`", col))
		}
		colList := strings.Join(cols, ", ")
		parts = append(parts, fmt.Sprintf("COLUMNS(%s)", colList))
	} else {
		parts = append(parts, "()")
	}

	if partitionOpts.PartitionCount != nil && *partitionOpts.PartitionCount > 0 {
		parts = append(parts, fmt.Sprintf("PARTITIONS %d", *partitionOpts.PartitionCount))
	}

	// Add partition definitions if present
	if len(partitionOpts.Partitions) > 0 {
		partDefs := []string{}
		for _, partDef := range partitionOpts.Partitions {
			partStr := fmt.Sprintf("PARTITION `%s`", partDef.Name)
			if len(partDef.Values) > 0 {
				switch partDef.Type {
				case "RANGE":
					partStr += fmt.Sprintf(" VALUES LESS THAN (%s)", strings.Join(partDef.Values, ", "))
				case "LIST":
					partStr += fmt.Sprintf(" VALUES IN (%s)", strings.Join(partDef.Values, ", "))
				}
			}
			partDefs = append(partDefs, partStr)
		}

		if len(partDefs) > 0 {
			parts = append(parts, "(")
			parts = append(parts, strings.Join(partDefs, ", "))
			parts = append(parts, ")")
		}
	}

	return strings.Join(parts, " ")
}

// MatchTablesByName matches tables from old and new schemas by name
func MatchTablesByName(oldTables, newTables []*parser.CreateTableStatement) map[string]struct {
	Old *parser.CreateTableStatement
	New *parser.CreateTableStatement
} {
	oldMap := make(map[string]*parser.CreateTableStatement)
	for _, table := range oldTables {
		oldMap[table.TableName] = table
	}

	newMap := make(map[string]*parser.CreateTableStatement)
	for _, table := range newTables {
		newMap[table.TableName] = table
	}

	allTableNames := make(map[string]bool)
	for name := range oldMap {
		allTableNames[name] = true
	}
	for name := range newMap {
		allTableNames[name] = true
	}

	matches := make(map[string]struct {
		Old *parser.CreateTableStatement
		New *parser.CreateTableStatement
	})

	for tableName := range allTableNames {
		matches[tableName] = struct {
			Old *parser.CreateTableStatement
			New *parser.CreateTableStatement
		}{
			Old: oldMap[tableName],
			New: newMap[tableName],
		}
	}

	return matches
}

// GenerateCreateTableStatements generates CREATE TABLE comments for completely new tables
func GenerateCreateTableStatements(newTables []*parser.CreateTableStatement, existingNames map[string]bool) []string {
	statements := []string{}

	for _, table := range newTables {
		if !existingNames[table.TableName] {
			statements = append(statements, fmt.Sprintf("-- CREATE TABLE `%s` (...); -- New table, full definition needed", table.TableName))
		}
	}

	return statements
}

// GenerateDropTableStatements generates DROP TABLE statements for removed tables
func GenerateDropTableStatements(oldTables []*parser.CreateTableStatement, existingNames map[string]bool) []string {
	statements := []string{}

	for _, table := range oldTables {
		if !existingNames[table.TableName] {
			statements = append(statements, fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", table.TableName))
		}
	}

	return statements
}
