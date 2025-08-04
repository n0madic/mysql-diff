package alter

import (
	"strings"
	"testing"

	"github.com/n0madic/mysql-diff/pkg/diff"
	"github.com/n0madic/mysql-diff/pkg/parser"
)

// Helper function to create a test table diff
func createTestTableDiff(oldTable, newTable *parser.CreateTableStatement) *diff.TableDiff {
	analyzer := diff.NewTableDiffAnalyzer()
	return analyzer.CompareTables(oldTable, newTable)
}

// Helper function to create a test table
func createTestTable(name string, columns []parser.ColumnDefinition) *parser.CreateTableStatement {
	return &parser.CreateTableStatement{
		TableName: name,
		Columns:   columns,
	}
}

// Helper function to create a test column
func createTestColumn(name, dataType string) parser.ColumnDefinition {
	return parser.ColumnDefinition{
		Name: name,
		DataType: parser.DataType{
			Name: dataType,
		},
	}
}

func TestMultipleChangesInSingleTable(t *testing.T) {
	// Create a table with multiple types of changes
	oldTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
			{Name: "name", DataType: parser.DataType{Name: "VARCHAR", Parameters: []string{"100"}}},
			{Name: "old_field", DataType: parser.DataType{Name: "TEXT"}},
		},
	}

	trueVal := true
	newTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}, AutoIncrement: true},                                    // Modified
			{Name: "name", DataType: parser.DataType{Name: "VARCHAR", Parameters: []string{"255"}}},                      // Modified
			{Name: "email", DataType: parser.DataType{Name: "VARCHAR", Parameters: []string{"255"}}, Nullable: &trueVal}, // Added
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)
	if !tableDiff.HasChanges() {
		t.Fatal("Expected changes in table diff")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) == 0 {
		t.Fatal("Expected ALTER statements to be generated")
	}

	// Check that statements contain expected operations
	allStatements := strings.Join(statements, " ")

	// Should have ADD COLUMN
	if !strings.Contains(allStatements, "ADD COLUMN") {
		t.Error("Expected ADD COLUMN statement")
	}

	// Should have MODIFY COLUMN
	if !strings.Contains(allStatements, "MODIFY COLUMN") {
		t.Error("Expected MODIFY COLUMN statement")
	}

	// Should have DROP COLUMN
	if !strings.Contains(allStatements, "DROP COLUMN") {
		t.Error("Expected DROP COLUMN statement")
	}
}

func TestPrimaryKeyChangeWithExistingPK(t *testing.T) {
	// Test changing from single PK to composite PK
	oldTable := &parser.CreateTableStatement{
		TableName: "orders",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
			{Name: "customer_id", DataType: parser.DataType{Name: "INT"}},
		},
		PrimaryKey: &parser.PrimaryKeyDefinition{
			Columns: []parser.IndexColumn{{Name: "id"}},
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "orders",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
			{Name: "customer_id", DataType: parser.DataType{Name: "INT"}},
		},
		PrimaryKey: &parser.PrimaryKeyDefinition{
			Columns: []parser.IndexColumn{{Name: "id"}, {Name: "customer_id"}},
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)
	if !tableDiff.HasChanges() {
		t.Fatal("Expected changes for PK modification")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) == 0 {
		t.Fatal("Expected ALTER statements for PK change")
	}

	allStatements := strings.Join(statements, " ")

	// Should DROP the old PK and ADD new PK
	if !strings.Contains(allStatements, "DROP PRIMARY KEY") {
		t.Error("Expected DROP PRIMARY KEY statement")
	}

	if !strings.Contains(allStatements, "ADD PRIMARY KEY") {
		t.Error("Expected ADD PRIMARY KEY statement")
	}
}

func TestConflictingChanges(t *testing.T) {
	// Test scenario where column is both renamed and modified
	// This tests the generator's ability to handle complex scenarios

	oldTable := &parser.CreateTableStatement{
		TableName: "products",
		Columns: []parser.ColumnDefinition{
			{Name: "price", DataType: parser.DataType{Name: "DECIMAL", Parameters: []string{"10", "2"}}},
		},
	}

	falseVal := false
	newTable := &parser.CreateTableStatement{
		TableName: "products",
		Columns: []parser.ColumnDefinition{
			{Name: "cost", DataType: parser.DataType{Name: "DECIMAL", Parameters: []string{"12", "4"}}, Nullable: &falseVal},
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	// Should generate some statements (exact behavior depends on implementation)
	// This test mainly ensures no panic occurs
	if len(statements) == 0 {
		t.Log("No statements generated - this might be expected for column renames")
	}
}

func TestTableWithOnlyIndexChanges(t *testing.T) {
	// Test table where only indexes change, no column changes
	indexName1 := "idx_name"
	indexName2 := "idx_user_name"

	oldTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "name", DataType: parser.DataType{Name: "VARCHAR"}},
		},
		Indexes: []parser.IndexDefinition{
			{Name: &indexName1, IndexType: "INDEX", Columns: []parser.IndexColumn{{Name: "name"}}},
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "name", DataType: parser.DataType{Name: "VARCHAR"}},
		},
		Indexes: []parser.IndexDefinition{
			{Name: &indexName2, IndexType: "INDEX", Columns: []parser.IndexColumn{{Name: "name"}}},
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)
	if !tableDiff.HasChanges() {
		t.Fatal("Expected changes for index modification")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) == 0 {
		t.Fatal("Expected ALTER statements for index changes")
	}

	allStatements := strings.Join(statements, " ")

	// Should have index operations
	if !strings.Contains(allStatements, "INDEX") {
		t.Error("Expected INDEX operations in statements")
	}
}

func TestForeignKeyWithCascadeOptions(t *testing.T) {
	// Test foreign key with ON DELETE/UPDATE options
	fkName := "fk_user_role"
	onDeleteCascade := "CASCADE"
	onUpdateRestrict := "RESTRICT"

	oldTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "role_id", DataType: parser.DataType{Name: "INT"}},
		},
		ForeignKeys: []parser.ForeignKeyDefinition{
			{
				Name:    &fkName,
				Columns: []string{"role_id"},
				Reference: parser.ForeignKeyReference{
					TableName: "roles",
					Columns:   []string{"id"},
				},
			},
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "role_id", DataType: parser.DataType{Name: "INT"}},
		},
		ForeignKeys: []parser.ForeignKeyDefinition{
			{
				Name:    &fkName,
				Columns: []string{"role_id"},
				Reference: parser.ForeignKeyReference{
					TableName: "roles",
					Columns:   []string{"id"},
					OnDelete:  &onDeleteCascade,
					OnUpdate:  &onUpdateRestrict,
				},
			},
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)
	if !tableDiff.HasChanges() {
		t.Fatal("Expected changes for FK options")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) == 0 {
		t.Fatal("Expected ALTER statements for FK changes")
	}

	allStatements := strings.Join(statements, " ")

	// Should handle foreign key changes
	if !strings.Contains(allStatements, "FOREIGN KEY") {
		t.Error("Expected FOREIGN KEY operations")
	}
}

func TestTableOptionsChanges(t *testing.T) {
	// Test changes in table options
	oldEngine := "MyISAM"
	newEngine := "InnoDB"
	oldCharset := "latin1"
	newCharset := "utf8mb4"

	oldTable := &parser.CreateTableStatement{
		TableName: "legacy_table",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
		},
		TableOptions: &parser.TableOptions{
			Engine:       &oldEngine,
			CharacterSet: &oldCharset,
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "legacy_table",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
		},
		TableOptions: &parser.TableOptions{
			Engine:       &newEngine,
			CharacterSet: &newCharset,
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)
	if !tableDiff.HasChanges() {
		t.Fatal("Expected changes for table options")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) == 0 {
		t.Fatal("Expected ALTER statements for table options")
	}

	allStatements := strings.Join(statements, " ")

	// Should contain engine and charset changes
	if !strings.Contains(allStatements, "ENGINE") {
		t.Error("Expected ENGINE change in statements")
	}

	if !strings.Contains(allStatements, "CHARSET") {
		t.Error("Expected CHARSET change in statements")
	}
}

func TestEmptyTableDiff(t *testing.T) {
	// Test with no changes
	table := createTestTable("test", []parser.ColumnDefinition{
		createTestColumn("id", "INT"),
	})

	tableDiff := createTestTableDiff(table, table)
	if tableDiff.HasChanges() {
		t.Fatal("Expected no changes for identical tables")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) != 0 {
		t.Errorf("Expected 0 statements for identical tables, got %d", len(statements))
	}
}

func TestNilTableDiffHandling(t *testing.T) {
	// Test generator with nil input
	generator := NewStatementGenerator()

	// This should not panic
	statements := generator.GenerateAlterStatements(nil)

	if len(statements) != 0 {
		t.Errorf("Expected 0 statements for nil diff, got %d", len(statements))
	}
}

func TestComplexDataTypeChanges(t *testing.T) {
	// Test complex data type changes that might require special handling
	oldTable := &parser.CreateTableStatement{
		TableName: "measurements",
		Columns: []parser.ColumnDefinition{
			{
				Name: "value",
				DataType: parser.DataType{
					Name:       "DECIMAL",
					Parameters: []string{"10", "2"},
					Unsigned:   false,
				},
			},
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "measurements",
		Columns: []parser.ColumnDefinition{
			{
				Name: "value",
				DataType: parser.DataType{
					Name:       "DECIMAL",
					Parameters: []string{"15", "4"},
					Unsigned:   true,
				},
			},
		},
	}

	tableDiff := createTestTableDiff(oldTable, newTable)
	if !tableDiff.HasChanges() {
		t.Fatal("Expected changes for complex data type modification")
	}

	generator := NewStatementGenerator()
	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) == 0 {
		t.Fatal("Expected ALTER statements for data type changes")
	}

	allStatements := strings.Join(statements, " ")

	// Should contain MODIFY COLUMN
	if !strings.Contains(allStatements, "MODIFY COLUMN") {
		t.Error("Expected MODIFY COLUMN for data type change")
	}

	// Should reflect the new precision and UNSIGNED
	if !strings.Contains(allStatements, "DECIMAL(15,4)") {
		t.Error("Expected new DECIMAL precision in statement")
	}

	if !strings.Contains(allStatements, "UNSIGNED") {
		t.Error("Expected UNSIGNED in statement")
	}
}
