package diff

import (
	"testing"

	"github.com/n0madic/mysql-diff/pkg/parser"
)

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

func TestIdenticalTablesWithDifferentColumnOrder(t *testing.T) {
	// Same columns, different order
	oldTable := createTestTable("users", []parser.ColumnDefinition{
		createTestColumn("id", "INT"),
		createTestColumn("name", "VARCHAR"),
		createTestColumn("email", "VARCHAR"),
	})

	newTable := createTestTable("users", []parser.ColumnDefinition{
		createTestColumn("email", "VARCHAR"),
		createTestColumn("id", "INT"),
		createTestColumn("name", "VARCHAR"),
	})

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	// Order changes should not be considered changes for now
	// (This is a design decision - we could implement column position tracking)
	if diff.HasChanges() {
		t.Error("Column order changes should not be detected as table changes")
	}
}

func TestComplexDataTypeChanges(t *testing.T) {
	oldColumn := parser.ColumnDefinition{
		Name: "amount",
		DataType: parser.DataType{
			Name:       "DECIMAL",
			Parameters: []string{"10", "2"},
		},
	}

	newColumn := parser.ColumnDefinition{
		Name: "amount",
		DataType: parser.DataType{
			Name:       "DECIMAL",
			Parameters: []string{"12", "4"},
		},
	}

	oldTable := createTestTable("transactions", []parser.ColumnDefinition{oldColumn})
	newTable := createTestTable("transactions", []parser.ColumnDefinition{newColumn})

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for DECIMAL precision/scale change")
	}

	if len(diff.ColumnDiffs) != 1 {
		t.Fatalf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected MODIFIED change type, got %s", colDiff.ChangeType)
	}

	if colDiff.Changes.DataType == nil {
		t.Error("Expected data type change to be detected")
	}
}

func TestNullabilityChanges(t *testing.T) {
	falseVal := false
	trueVal := true

	oldColumn := parser.ColumnDefinition{
		Name:     "name",
		DataType: parser.DataType{Name: "VARCHAR"},
		Nullable: &trueVal, // NULL allowed
	}

	newColumn := parser.ColumnDefinition{
		Name:     "name",
		DataType: parser.DataType{Name: "VARCHAR"},
		Nullable: &falseVal, // NOT NULL
	}

	oldTable := createTestTable("users", []parser.ColumnDefinition{oldColumn})
	newTable := createTestTable("users", []parser.ColumnDefinition{newColumn})

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for nullability change")
	}

	if len(diff.ColumnDiffs) != 1 {
		t.Fatalf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.Nullable == nil {
		t.Error("Expected nullable change to be detected")
	}
}

func TestDefaultValueChanges(t *testing.T) {
	oldDefault := "active"
	newDefault := "pending"

	oldColumn := parser.ColumnDefinition{
		Name:         "status",
		DataType:     parser.DataType{Name: "VARCHAR"},
		DefaultValue: &oldDefault,
	}

	newColumn := parser.ColumnDefinition{
		Name:         "status",
		DataType:     parser.DataType{Name: "VARCHAR"},
		DefaultValue: &newDefault,
	}

	oldTable := createTestTable("users", []parser.ColumnDefinition{oldColumn})
	newTable := createTestTable("users", []parser.ColumnDefinition{newColumn})

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for default value change")
	}

	if len(diff.ColumnDiffs) != 1 {
		t.Fatalf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.DefaultValue == nil {
		t.Error("Expected default value change to be detected")
	}

	if colDiff.Changes.DefaultValue.Old != oldDefault {
		t.Errorf("Expected old default '%s', got %v", oldDefault, colDiff.Changes.DefaultValue.Old)
	}

	if colDiff.Changes.DefaultValue.New != newDefault {
		t.Errorf("Expected new default '%s', got %v", newDefault, colDiff.Changes.DefaultValue.New)
	}
}

func TestDefaultValueFromNullToValue(t *testing.T) {
	newDefault := "0"

	oldColumn := parser.ColumnDefinition{
		Name:         "count",
		DataType:     parser.DataType{Name: "INT"},
		DefaultValue: nil, // No default
	}

	newColumn := parser.ColumnDefinition{
		Name:         "count",
		DataType:     parser.DataType{Name: "INT"},
		DefaultValue: &newDefault,
	}

	oldTable := createTestTable("stats", []parser.ColumnDefinition{oldColumn})
	newTable := createTestTable("stats", []parser.ColumnDefinition{newColumn})

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for adding default value")
	}

	if len(diff.ColumnDiffs) != 1 {
		t.Fatalf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.DefaultValue == nil {
		t.Error("Expected default value change to be detected")
	}
}

func TestAutoIncrementChanges(t *testing.T) {
	oldColumn := parser.ColumnDefinition{
		Name:          "id",
		DataType:      parser.DataType{Name: "INT"},
		AutoIncrement: false,
	}

	newColumn := parser.ColumnDefinition{
		Name:          "id",
		DataType:      parser.DataType{Name: "INT"},
		AutoIncrement: true,
	}

	oldTable := createTestTable("users", []parser.ColumnDefinition{oldColumn})
	newTable := createTestTable("users", []parser.ColumnDefinition{newColumn})

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for AUTO_INCREMENT change")
	}

	if len(diff.ColumnDiffs) != 1 {
		t.Fatalf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.AutoIncrement == nil {
		t.Error("Expected auto_increment change to be detected")
	}
}

func TestComplexIndexChanges(t *testing.T) {
	indexName1 := "idx_name"
	indexName2 := "idx_user_name"

	oldIndex := &parser.IndexDefinition{
		Name:      &indexName1,
		IndexType: "INDEX",
		Columns: []parser.IndexColumn{
			{Name: "name"},
		},
	}

	newIndex := &parser.IndexDefinition{
		Name:      &indexName2,
		IndexType: "INDEX",
		Columns: []parser.IndexColumn{
			{Name: "name"},
		},
	}

	oldTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns:   []parser.ColumnDefinition{createTestColumn("name", "VARCHAR")},
		Indexes:   []parser.IndexDefinition{*oldIndex},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns:   []parser.ColumnDefinition{createTestColumn("name", "VARCHAR")},
		Indexes:   []parser.IndexDefinition{*newIndex},
	}

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for index name change")
	}

	if len(diff.IndexDiffs) != 1 {
		t.Fatalf("Expected 1 index diff, got %d", len(diff.IndexDiffs))
	}

	idxDiff := diff.IndexDiffs[0]
	if idxDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected MODIFIED change type, got %s", idxDiff.ChangeType)
	}
}

func TestPrimaryKeyChanges(t *testing.T) {
	oldPK := &parser.PrimaryKeyDefinition{
		Columns: []parser.IndexColumn{
			{Name: "id"},
		},
	}

	newPK := &parser.PrimaryKeyDefinition{
		Columns: []parser.IndexColumn{
			{Name: "id"},
			{Name: "tenant_id"},
		},
	}

	oldTable := &parser.CreateTableStatement{
		TableName:  "users",
		Columns:    []parser.ColumnDefinition{createTestColumn("id", "INT")},
		PrimaryKey: oldPK,
	}

	newTable := &parser.CreateTableStatement{
		TableName:  "users",
		Columns:    []parser.ColumnDefinition{createTestColumn("id", "INT")},
		PrimaryKey: newPK,
	}

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for primary key change")
	}

	if diff.PrimaryKeyDiff == nil {
		t.Error("Expected primary key diff to be detected")
	}

	if diff.PrimaryKeyDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected MODIFIED change type, got %s", diff.PrimaryKeyDiff.ChangeType)
	}
}

func TestForeignKeyChanges(t *testing.T) {
	fkName := "fk_user_role"

	oldFK := parser.ForeignKeyDefinition{
		Name:    &fkName,
		Columns: []string{"role_id"},
		Reference: parser.ForeignKeyReference{
			TableName: "roles",
			Columns:   []string{"id"},
		},
	}

	onDeleteCascade := "CASCADE"
	newFK := parser.ForeignKeyDefinition{
		Name:    &fkName,
		Columns: []string{"role_id"},
		Reference: parser.ForeignKeyReference{
			TableName: "roles",
			Columns:   []string{"id"},
			OnDelete:  &onDeleteCascade,
		},
	}

	oldTable := &parser.CreateTableStatement{
		TableName:   "users",
		Columns:     []parser.ColumnDefinition{createTestColumn("role_id", "INT")},
		ForeignKeys: []parser.ForeignKeyDefinition{oldFK},
	}

	newTable := &parser.CreateTableStatement{
		TableName:   "users",
		Columns:     []parser.ColumnDefinition{createTestColumn("role_id", "INT")},
		ForeignKeys: []parser.ForeignKeyDefinition{newFK},
	}

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for foreign key ON DELETE change")
	}

	if len(diff.ForeignKeyDiffs) != 1 {
		t.Fatalf("Expected 1 foreign key diff, got %d", len(diff.ForeignKeyDiffs))
	}

	fkDiff := diff.ForeignKeyDiffs[0]
	if fkDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected MODIFIED change type, got %s", fkDiff.ChangeType)
	}
}

func TestTableOptionsChanges(t *testing.T) {
	oldEngine := "MyISAM"
	newEngine := "InnoDB"
	oldCharset := "latin1"
	newCharset := "utf8mb4"

	oldOptions := &parser.TableOptions{
		Engine:       &oldEngine,
		CharacterSet: &oldCharset,
	}

	newOptions := &parser.TableOptions{
		Engine:       &newEngine,
		CharacterSet: &newCharset,
	}

	oldTable := &parser.CreateTableStatement{
		TableName:    "users",
		Columns:      []parser.ColumnDefinition{createTestColumn("id", "INT")},
		TableOptions: oldOptions,
	}

	newTable := &parser.CreateTableStatement{
		TableName:    "users",
		Columns:      []parser.ColumnDefinition{createTestColumn("id", "INT")},
		TableOptions: newOptions,
	}

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for table options change")
	}

	if diff.TableOptionsDiff == nil {
		t.Error("Expected table options diff to be detected")
	}

	if diff.TableOptionsDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected MODIFIED change type, got %s", diff.TableOptionsDiff.ChangeType)
	}
}

func TestEmptyTablesComparisonEdgeCase(t *testing.T) {
	// Tables with no columns
	oldTable := &parser.CreateTableStatement{
		TableName: "empty_table",
		Columns:   []parser.ColumnDefinition{},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "empty_table",
		Columns:   []parser.ColumnDefinition{},
	}

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if diff.HasChanges() {
		t.Error("Empty tables should not have changes")
	}
}

func TestNilTablesHandling(t *testing.T) {
	analyzer := NewTableDiffAnalyzer()

	// Test with nil old table
	newTable := createTestTable("test", []parser.ColumnDefinition{createTestColumn("id", "INT")})
	diff := analyzer.CompareTables(nil, newTable)

	if diff.OldTable != nil {
		t.Error("Expected OldTable to be nil")
	}
	if diff.NewTable != newTable {
		t.Error("Expected NewTable to match input")
	}

	// Test with nil new table
	oldTable := createTestTable("test", []parser.ColumnDefinition{createTestColumn("id", "INT")})
	diff = analyzer.CompareTables(oldTable, nil)

	if diff.OldTable != oldTable {
		t.Error("Expected OldTable to match input")
	}
	if diff.NewTable != nil {
		t.Error("Expected NewTable to be nil")
	}

	// Test with both nil
	diff = analyzer.CompareTables(nil, nil)
	if diff.OldTable != nil || diff.NewTable != nil {
		t.Error("Both tables should be nil")
	}
}
