package diff

import (
	"strings"
	"testing"

	"github.com/n0madic/mysql-diff/pkg/parser"
)

// TestIdenticalTablesNoDiff tests that identical tables show no differences
func TestIdenticalTablesNoDiff(t *testing.T) {
	sql := `
		CREATE TABLE test (
			id INT NOT NULL AUTO_INCREMENT,
			name VARCHAR(255),
			PRIMARY KEY (id)
		)
	`

	tables, err := parser.ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("Failed to parse SQL: %v", err)
	}

	table1 := tables[0]
	table2 := tables[0] // Same table

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(table1, table2)

	if diff.HasChanges() {
		t.Error("Expected no changes for identical tables")
	}
	if diff.TableNameChanged {
		t.Error("Expected table name not to be changed")
	}
	if len(diff.ColumnDiffs) != 0 {
		t.Errorf("Expected 0 column diffs, got %d", len(diff.ColumnDiffs))
	}
	if diff.PrimaryKeyDiff != nil {
		t.Error("Expected no primary key diff")
	}
	if len(diff.IndexDiffs) != 0 {
		t.Errorf("Expected 0 index diffs, got %d", len(diff.IndexDiffs))
	}
	if len(diff.ForeignKeyDiffs) != 0 {
		t.Errorf("Expected 0 foreign key diffs, got %d", len(diff.ForeignKeyDiffs))
	}
	if diff.TableOptionsDiff != nil {
		t.Error("Expected no table options diff")
	}
	if diff.PartitionDiff != nil {
		t.Error("Expected no partition diff")
	}
}

// TestTableNameChange tests detection of table name change
func TestTableNameChange(t *testing.T) {
	sql1 := "CREATE TABLE old_name (id INT)"
	sql2 := "CREATE TABLE new_name (id INT)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for table name change")
	}
	if !diff.TableNameChanged {
		t.Error("Expected table name to be changed")
	}
}

// TestColumnAdded tests detection of added columns
func TestColumnAdded(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT)"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for added column")
	}
	if len(diff.ColumnDiffs) != 1 {
		t.Errorf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}
	if diff.ColumnsAdded != 1 {
		t.Errorf("Expected 1 column added, got %d", diff.ColumnsAdded)
	}
	if diff.ColumnsRemoved != 0 {
		t.Errorf("Expected 0 columns removed, got %d", diff.ColumnsRemoved)
	}
	if diff.ColumnsModified != 0 {
		t.Errorf("Expected 0 columns modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Name != "name" {
		t.Errorf("Expected column name 'name', got '%s'", colDiff.Name)
	}
	if colDiff.ChangeType != ChangeTypeAdded {
		t.Errorf("Expected change type ADDED, got %s", colDiff.ChangeType)
	}
	if colDiff.OldColumn != nil {
		t.Error("Expected old column to be nil for added column")
	}
	if colDiff.NewColumn == nil {
		t.Error("Expected new column to be not nil for added column")
	}
}

// TestColumnRemoved tests detection of removed columns
func TestColumnRemoved(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, name VARCHAR(255))"
	sql2 := "CREATE TABLE test (id INT)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for removed column")
	}
	if len(diff.ColumnDiffs) != 1 {
		t.Errorf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}
	if diff.ColumnsAdded != 0 {
		t.Errorf("Expected 0 columns added, got %d", diff.ColumnsAdded)
	}
	if diff.ColumnsRemoved != 1 {
		t.Errorf("Expected 1 column removed, got %d", diff.ColumnsRemoved)
	}
	if diff.ColumnsModified != 0 {
		t.Errorf("Expected 0 columns modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Name != "name" {
		t.Errorf("Expected column name 'name', got '%s'", colDiff.Name)
	}
	if colDiff.ChangeType != ChangeTypeRemoved {
		t.Errorf("Expected change type REMOVED, got %s", colDiff.ChangeType)
	}
	if colDiff.OldColumn == nil {
		t.Error("Expected old column to be not nil for removed column")
	}
	if colDiff.NewColumn != nil {
		t.Error("Expected new column to be nil for removed column")
	}
}

// TestColumnDataTypeModified tests detection of column data type changes
func TestColumnDataTypeModified(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, amount DECIMAL(10,2))"
	sql2 := "CREATE TABLE test (id INT, amount DECIMAL(12,4))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for column data type change")
	}
	if len(diff.ColumnDiffs) != 1 {
		t.Errorf("Expected 1 column diff, got %d", len(diff.ColumnDiffs))
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Name != "amount" {
		t.Errorf("Expected column name 'amount', got '%s'", colDiff.Name)
	}
	if colDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected change type MODIFIED, got %s", colDiff.ChangeType)
	}

	if colDiff.Changes.DataType == nil {
		t.Error("Expected data_type change in column diff")
	}

	dataTypeChange := colDiff.Changes.DataType
	if dataTypeChange.Old != "DECIMAL(10,2)" {
		t.Errorf("Expected old data type 'DECIMAL(10,2)', got '%v'", dataTypeChange.Old)
	}
	if dataTypeChange.New != "DECIMAL(12,4)" {
		t.Errorf("Expected new data type 'DECIMAL(12,4)', got '%v'", dataTypeChange.New)
	}
}

// TestColumnNullableChange tests detection of nullable changes
func TestColumnNullableChange(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT NULL)"
	sql2 := "CREATE TABLE test (id INT NOT NULL)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for nullable change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.Nullable == nil {
		t.Error("Expected nullable change in column diff")
	}

	nullableChange := colDiff.Changes.Nullable
	if nullableChange.Old != true { // NULL
		t.Errorf("Expected old nullable to be true, got %v", nullableChange.Old)
	}
	if nullableChange.New != false { // NOT NULL
		t.Errorf("Expected new nullable to be false, got %v", nullableChange.New)
	}
}

// TestColumnDefaultValueChange tests detection of default value changes
func TestColumnDefaultValueChange(t *testing.T) {
	sql1 := "CREATE TABLE test (status VARCHAR(10) DEFAULT 'active')"
	sql2 := "CREATE TABLE test (status VARCHAR(10) DEFAULT 'pending')"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for default value change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.DefaultValue == nil {
		t.Error("Expected default_value change in column diff")
	}

	defaultChange := colDiff.Changes.DefaultValue
	if defaultChange.Old != "active" {
		t.Errorf("Expected old default value 'active', got '%v'", defaultChange.Old)
	}
	if defaultChange.New != "pending" {
		t.Errorf("Expected new default value 'pending', got '%v'", defaultChange.New)
	}
}

// TestColumnAutoIncrementChange tests detection of auto_increment changes
func TestColumnAutoIncrementChange(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT)"
	sql2 := "CREATE TABLE test (id INT AUTO_INCREMENT)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for auto_increment change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.AutoIncrement == nil {
		t.Error("Expected auto_increment change in column diff")
	}

	autoIncChange := colDiff.Changes.AutoIncrement
	if autoIncChange.Old != false {
		t.Errorf("Expected old auto_increment to be false, got %v", autoIncChange.Old)
	}
	if autoIncChange.New != true {
		t.Errorf("Expected new auto_increment to be true, got %v", autoIncChange.New)
	}
}

// TestPrimaryKeyAdded tests detection of added primary key
func TestPrimaryKeyAdded(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, name VARCHAR(255))"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255), PRIMARY KEY (id))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for added primary key")
	}
	if diff.PrimaryKeyDiff == nil {
		t.Fatal("Expected primary key diff to be not nil")
	}
	if diff.PrimaryKeyDiff.ChangeType != ChangeTypeAdded {
		t.Errorf("Expected primary key change type ADDED, got %s", diff.PrimaryKeyDiff.ChangeType)
	}
	if diff.PrimaryKeyDiff.OldPK != nil {
		t.Error("Expected old primary key to be nil for added primary key")
	}
	if diff.PrimaryKeyDiff.NewPK == nil {
		t.Error("Expected new primary key to be not nil for added primary key")
	}
}

// TestPrimaryKeyRemoved tests detection of removed primary key
func TestPrimaryKeyRemoved(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, name VARCHAR(255), PRIMARY KEY (id))"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for removed primary key")
	}
	if diff.PrimaryKeyDiff == nil {
		t.Fatal("Expected primary key diff to be not nil")
	}
	if diff.PrimaryKeyDiff.ChangeType != ChangeTypeRemoved {
		t.Errorf("Expected primary key change type REMOVED, got %s", diff.PrimaryKeyDiff.ChangeType)
	}
	if diff.PrimaryKeyDiff.OldPK == nil {
		t.Error("Expected old primary key to be not nil for removed primary key")
	}
	if diff.PrimaryKeyDiff.NewPK != nil {
		t.Error("Expected new primary key to be nil for removed primary key")
	}
}

// TestPrimaryKeyModified tests detection of modified primary key
func TestPrimaryKeyModified(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, name VARCHAR(255), PRIMARY KEY (id))"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255), PRIMARY KEY (id, name))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for modified primary key")
	}
	if diff.PrimaryKeyDiff == nil {
		t.Fatal("Expected primary key diff to be not nil")
	}
	if diff.PrimaryKeyDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected primary key change type MODIFIED, got %s", diff.PrimaryKeyDiff.ChangeType)
	}

	if diff.PrimaryKeyDiff.Changes.Columns == nil {
		t.Error("Expected columns change in primary key diff")
	}

	columnsChange := diff.PrimaryKeyDiff.Changes.Columns
	oldColumns := columnsChange.Old
	if len(oldColumns) != 1 || oldColumns[0] != "id" {
		t.Errorf("Expected old columns to be ['id'], got %v", oldColumns)
	}

	newColumns := columnsChange.New
	if len(newColumns) != 2 || newColumns[0] != "id" || newColumns[1] != "name" {
		t.Errorf("Expected new columns to be ['id', 'name'], got %v", newColumns)
	}
}

// TestIndexAdded tests detection of added indexes
func TestIndexAdded(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, name VARCHAR(255))"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255), INDEX idx_name (name))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for added index")
	}
	if len(diff.IndexDiffs) != 1 {
		t.Errorf("Expected 1 index diff, got %d", len(diff.IndexDiffs))
	}
	if diff.IndexesAdded != 1 {
		t.Errorf("Expected 1 index added, got %d", diff.IndexesAdded)
	}

	idxDiff := diff.IndexDiffs[0]
	if idxDiff.ChangeType != ChangeTypeAdded {
		t.Errorf("Expected index change type ADDED, got %s", idxDiff.ChangeType)
	}
	if idxDiff.Name == nil || *idxDiff.Name != "idx_name" {
		t.Errorf("Expected index name 'idx_name', got %v", idxDiff.Name)
	}
	if idxDiff.OldIndex != nil {
		t.Error("Expected old index to be nil for added index")
	}
	if idxDiff.NewIndex == nil {
		t.Error("Expected new index to be not nil for added index")
	}
}

// TestIndexRemoved tests detection of removed indexes
func TestIndexRemoved(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, name VARCHAR(255), INDEX idx_name (name))"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for removed index")
	}
	if len(diff.IndexDiffs) != 1 {
		t.Errorf("Expected 1 index diff, got %d", len(diff.IndexDiffs))
	}
	if diff.IndexesRemoved != 1 {
		t.Errorf("Expected 1 index removed, got %d", diff.IndexesRemoved)
	}

	idxDiff := diff.IndexDiffs[0]
	if idxDiff.ChangeType != ChangeTypeRemoved {
		t.Errorf("Expected index change type REMOVED, got %s", idxDiff.ChangeType)
	}
}

// TestForeignKeyAdded tests detection of added foreign keys
func TestForeignKeyAdded(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, user_id INT)"
	sql2 := `
		CREATE TABLE test (
			id INT,
			user_id INT,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for added foreign key")
	}
	if len(diff.ForeignKeyDiffs) != 1 {
		t.Errorf("Expected 1 foreign key diff, got %d", len(diff.ForeignKeyDiffs))
	}
	if diff.ForeignKeysAdded != 1 {
		t.Errorf("Expected 1 foreign key added, got %d", diff.ForeignKeysAdded)
	}

	fkDiff := diff.ForeignKeyDiffs[0]
	if fkDiff.ChangeType != ChangeTypeAdded {
		t.Errorf("Expected foreign key change type ADDED, got %s", fkDiff.ChangeType)
	}
	if fkDiff.OldFK != nil {
		t.Error("Expected old foreign key to be nil for added foreign key")
	}
	if fkDiff.NewFK == nil {
		t.Error("Expected new foreign key to be not nil for added foreign key")
	}
}

// TestForeignKeyModified tests detection of modified foreign keys
func TestForeignKeyModified(t *testing.T) {
	sql1 := `
		CREATE TABLE test (
			id INT,
			user_id INT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`
	sql2 := `
		CREATE TABLE test (
			id INT,
			user_id INT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
		)
	`

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for modified foreign key")
	}
	if len(diff.ForeignKeyDiffs) != 1 {
		t.Errorf("Expected 1 foreign key diff, got %d", len(diff.ForeignKeyDiffs))
	}
	if diff.ForeignKeysModified != 1 {
		t.Errorf("Expected 1 foreign key modified, got %d", diff.ForeignKeysModified)
	}

	fkDiff := diff.ForeignKeyDiffs[0]
	if fkDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected foreign key change type MODIFIED, got %s", fkDiff.ChangeType)
	}

	if fkDiff.Changes.OnDelete == nil {
		t.Error("Expected on_delete change in foreign key diff")
	}

	onDeleteChange := fkDiff.Changes.OnDelete
	if onDeleteChange.Old != "CASCADE" {
		t.Errorf("Expected old on_delete 'CASCADE', got '%v'", onDeleteChange.Old)
	}
	if onDeleteChange.New != "RESTRICT" {
		t.Errorf("Expected new on_delete 'RESTRICT', got '%v'", onDeleteChange.New)
	}
}

// TestTableOptionsModified tests detection of table options changes
func TestTableOptionsModified(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT) ENGINE=MyISAM"
	sql2 := "CREATE TABLE test (id INT) ENGINE=InnoDB"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for table options change")
	}
	if diff.TableOptionsDiff == nil {
		t.Fatal("Expected table options diff to be not nil")
	}
	if diff.TableOptionsDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected table options change type MODIFIED, got %s", diff.TableOptionsDiff.ChangeType)
	}

	if diff.TableOptionsDiff.Changes.Engine == nil {
		t.Error("Expected engine change in table options diff")
	}

	engineChange := diff.TableOptionsDiff.Changes.Engine
	if engineChange.Old != "MyISAM" {
		t.Errorf("Expected old engine 'MyISAM', got '%v'", engineChange.Old)
	}
	if engineChange.New != "InnoDB" {
		t.Errorf("Expected new engine 'InnoDB', got '%v'", engineChange.New)
	}
}

// TestMultipleTableOptionsChanges tests detection of multiple table options changes
func TestMultipleTableOptionsChanges(t *testing.T) {
	sql1 := `
		CREATE TABLE test (id INT)
		ENGINE=MyISAM
		DEFAULT CHARSET=latin1
		COMMENT='Old table'
	`
	sql2 := `
		CREATE TABLE test (id INT)
		ENGINE=InnoDB
		DEFAULT CHARSET=utf8mb4
		COMMENT='New table'
		AUTO_INCREMENT=1000
	`

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for multiple table options changes")
	}
	if diff.TableOptionsDiff == nil {
		t.Fatal("Expected table options diff to be not nil")
	}
	if diff.TableOptionsDiff.ChangeType != ChangeTypeModified {
		t.Errorf("Expected table options change type MODIFIED, got %s", diff.TableOptionsDiff.ChangeType)
	}

	changes := diff.TableOptionsDiff.Changes

	// Check engine change
	if changes.Engine == nil {
		t.Error("Expected engine change in table options diff")
	}
	if changes.Engine.Old != "MyISAM" || changes.Engine.New != "InnoDB" {
		t.Errorf("Expected engine change MyISAM->InnoDB, got %v->%v", changes.Engine.Old, changes.Engine.New)
	}

	// Check character_set change
	if changes.CharacterSet == nil {
		t.Error("Expected character_set change in table options diff")
	}
	if changes.CharacterSet.Old != "latin1" || changes.CharacterSet.New != "utf8mb4" {
		t.Errorf("Expected character_set change latin1->utf8mb4, got %v->%v", changes.CharacterSet.Old, changes.CharacterSet.New)
	}

	// Check comment change
	if changes.Comment == nil {
		t.Error("Expected comment change in table options diff")
	}
	if changes.Comment.Old != "Old table" || changes.Comment.New != "New table" {
		t.Errorf("Expected comment change 'Old table'->'New table', got %v->%v", changes.Comment.Old, changes.Comment.New)
	}

	// Check auto_increment change (added)
	if changes.AutoIncrement == nil {
		t.Error("Expected auto_increment change in table options diff")
	}
	if changes.AutoIncrement.Old != nil || changes.AutoIncrement.New != 1000 {
		t.Errorf("Expected auto_increment change nil->1000, got %v->%v", changes.AutoIncrement.Old, changes.AutoIncrement.New)
	}
}

// TestPartitionOptionsAdded tests detection of added partitioning
func TestPartitionOptionsAdded(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT, data VARCHAR(100))"
	sql2 := `
		CREATE TABLE test (
			id INT,
			data VARCHAR(100)
		) PARTITION BY HASH(id) PARTITIONS 4
	`

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for added partitioning")
	}
	if diff.PartitionDiff == nil {
		t.Fatal("Expected partition diff to be not nil")
	}
	if diff.PartitionDiff.ChangeType != ChangeTypeAdded {
		t.Errorf("Expected partition change type ADDED, got %s", diff.PartitionDiff.ChangeType)
	}
	if diff.PartitionDiff.OldPartition != nil {
		t.Error("Expected old partition to be nil for added partitioning")
	}
	if diff.PartitionDiff.NewPartition == nil {
		t.Error("Expected new partition to be not nil for added partitioning")
	}
}

// TestDataTypeUnsignedZerofillChanges tests detection of UNSIGNED and ZEROFILL changes
func TestDataTypeUnsignedZerofillChanges(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT)"
	sql2 := "CREATE TABLE test (id INT UNSIGNED ZEROFILL)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for UNSIGNED ZEROFILL change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.DataType == nil {
		t.Error("Expected data_type change in column diff")
	}

	dataTypeChange := colDiff.Changes.DataType
	if dataTypeChange.Old != "INT" {
		t.Errorf("Expected old data type 'INT', got '%v'", dataTypeChange.Old)
	}
	if dataTypeChange.New != "INT UNSIGNED ZEROFILL" {
		t.Errorf("Expected new data type 'INT UNSIGNED ZEROFILL', got '%v'", dataTypeChange.New)
	}
}

// TestCharacterSetCollationChanges tests detection of character set and collation changes
func TestCharacterSetCollationChanges(t *testing.T) {
	sql1 := "CREATE TABLE test (name VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci)"
	sql2 := "CREATE TABLE test (name VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for character set and collation change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]

	if colDiff.Changes.CharacterSet == nil {
		t.Error("Expected character_set change in column diff")
	}
	if colDiff.Changes.Collation == nil {
		t.Error("Expected collation change in column diff")
	}

	charsetChange := colDiff.Changes.CharacterSet
	if charsetChange.Old != "utf8" || charsetChange.New != "utf8mb4" {
		t.Errorf("Expected character_set change utf8->utf8mb4, got %v->%v", charsetChange.Old, charsetChange.New)
	}

	collationChange := colDiff.Changes.Collation
	if collationChange.Old != "utf8_general_ci" || collationChange.New != "utf8mb4_unicode_ci" {
		t.Errorf("Expected collation change utf8_general_ci->utf8mb4_unicode_ci, got %v->%v", collationChange.Old, collationChange.New)
	}
}

// TestComplexDiffScenario tests a complex scenario with multiple types of changes
func TestComplexDiffScenario(t *testing.T) {
	sql1 := `
		CREATE TABLE users (
			id INT NOT NULL AUTO_INCREMENT,
			name VARCHAR(100),
			email VARCHAR(255) UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			INDEX idx_name (name)
		) ENGINE=MyISAM DEFAULT CHARSET=latin1
	`

	sql2 := `
		CREATE TABLE users (
			id BIGINT NOT NULL AUTO_INCREMENT,
			name VARCHAR(150) NOT NULL,
			email VARCHAR(255) UNIQUE,
			phone VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			INDEX idx_name_phone (name, phone),
			FOREIGN KEY (id) REFERENCES profiles(user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User accounts'
	`

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for complex diff scenario")
	}

	// Check summary
	summary := diff.GetSummary()

	if summary.Columns.Added == 0 {
		t.Error("Expected some columns to be added (phone, updated_at)")
	}
	if summary.Columns.Modified == 0 {
		t.Error("Expected some columns to be modified (id type, name nullable)")
	}

	if summary.Indexes.Added == 0 {
		t.Error("Expected some indexes to be added")
	}
	if summary.Indexes.Removed == 0 {
		t.Error("Expected some indexes to be removed")
	}

	if summary.ForeignKeys.Added == 0 {
		t.Error("Expected some foreign keys to be added")
	}

	if !summary.TableOptionsChanged {
		t.Error("Expected table options to be changed (engine, charset, comment)")
	}

	// Verify specific changes
	columnNamesChanged := make([]string, len(diff.ColumnDiffs))
	for i, cd := range diff.ColumnDiffs {
		columnNamesChanged[i] = cd.Name
	}

	expectedColumns := []string{"id", "name", "phone", "updated_at"}
	for _, expectedCol := range expectedColumns {
		found := false
		for _, actualCol := range columnNamesChanged {
			if actualCol == expectedCol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column '%s' to be in changes, but not found", expectedCol)
		}
	}
}

// TestPrintTableDiffFunction tests the PrintTableDiff function doesn't crash
func TestPrintTableDiffFunction(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT)"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	diff := CompareTables(oldTable, newTable)

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintTableDiff panicked: %v", r)
		}
	}()

	// Capture output to avoid cluttering test output
	// In a real scenario you might want to capture and validate the output
	PrintTableDiff(diff, true)
	PrintTableDiff(diff, false)
	PrintDiffSummary(diff)
}

// TestEmptyTablesComparison tests comparison of tables with no columns (edge case)
func TestEmptyTablesComparison(t *testing.T) {
	sql1 := "CREATE TABLE test1 (id INT)"
	sql2 := "CREATE TABLE test2 (id INT)"

	tables1, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse SQL1: %v", err)
	}
	tables2, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse SQL2: %v", err)
	}

	table1 := tables1[0]
	table2 := tables2[0]

	// Artificially clear columns to test edge case
	table1.Columns = []parser.ColumnDefinition{}
	table2.Columns = []parser.ColumnDefinition{}

	diff := CompareTables(table1, table2)

	// Should not crash and should detect table name change
	if !diff.TableNameChanged {
		t.Error("Expected table name to be changed (test1 vs test2)")
	}
	if len(diff.ColumnDiffs) != 0 {
		t.Errorf("Expected 0 column diffs for empty tables, got %d", len(diff.ColumnDiffs))
	}
}

// TestNoneValuesHandling tests handling of nil/null values in comparisons
func TestNoneValuesHandling(t *testing.T) {
	sql := "CREATE TABLE test (id INT)"
	tables, err := parser.ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("Failed to parse SQL: %v", err)
	}

	table := tables[0]

	// Set some fields to nil to test edge cases
	table.PrimaryKey = nil
	table.TableOptions = nil
	table.PartitionOptions = nil
	table.Indexes = []parser.IndexDefinition{}
	table.ForeignKeys = []parser.ForeignKeyDefinition{}

	diff := CompareTables(table, table)

	// Should not crash
	if diff.HasChanges() {
		t.Error("Expected no changes for identical tables with nil values")
	}
}

// TestConvenienceFunction tests the convenience CompareTables function
func TestConvenienceFunction(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT)"
	sql2 := "CREATE TABLE test (id INT, name VARCHAR(255))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	// Test convenience function
	diff := CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes from convenience function")
	}
	if diff.ColumnsAdded != 1 {
		t.Errorf("Expected 1 column added, got %d", diff.ColumnsAdded)
	}
}

// TestGetSummaryMethod tests the GetSummary method
func TestGetSummaryMethod(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT)"
	sql2 := "CREATE TABLE test (id BIGINT, name VARCHAR(100))"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	diff := CompareTables(oldTable, newTable)
	summary := diff.GetSummary()

	// Check summary structure
	// Table name should not be changed in this test
	if summary.TableNameChanged {
		t.Error("Table name should not be changed in this test")
	}

	if summary.Columns.Added != 1 {
		t.Errorf("Expected 1 added column in summary, got %d", summary.Columns.Added)
	}
	if summary.Columns.Modified != 1 {
		t.Errorf("Expected 1 modified column in summary, got %d", summary.Columns.Modified)
	}

	// Test other summary fields - should have 0 index changes for this simple test
	if summary.Indexes.Added != 0 {
		t.Errorf("Expected 0 added indexes in summary, got %d", summary.Indexes.Added)
	}

	// Should have 0 FK changes for this simple test
	if summary.ForeignKeys.Added != 0 {
		t.Errorf("Expected 0 added foreign keys in summary, got %d", summary.ForeignKeys.Added)
	}

	// Test boolean fields - no need to check types as they're now compile-time checked
	// For this test, no primary key, table options, or partitioning changes expected
	if summary.PrimaryKeyChanged {
		t.Error("No primary key changes expected in this test")
	}
	if summary.TableOptionsChanged {
		t.Error("No table options changes expected in this test")
	}
	if summary.PartitioningChanged {
		t.Error("No partitioning changes expected in this test")
	}
}

// TestCommentChanges tests detection of comment changes
func TestCommentChanges(t *testing.T) {
	sql1 := "CREATE TABLE test (id INT COMMENT 'Old comment')"
	sql2 := "CREATE TABLE test (id INT COMMENT 'New comment')"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for comment change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.Comment == nil {
		t.Error("Expected comment change in column diff")
	}

	commentChange := colDiff.Changes.Comment
	if !strings.Contains(commentChange.Old.(string), "Old comment") {
		t.Errorf("Expected old comment to contain 'Old comment', got '%v'", commentChange.Old)
	}
	if !strings.Contains(commentChange.New.(string), "New comment") {
		t.Errorf("Expected new comment to contain 'New comment', got '%v'", commentChange.New)
	}
}

// TestUniqueConstraintChanges tests detection of unique constraint changes
func TestUniqueConstraintChanges(t *testing.T) {
	sql1 := "CREATE TABLE test (email VARCHAR(255))"
	sql2 := "CREATE TABLE test (email VARCHAR(255) UNIQUE)"

	oldTables, err := parser.ParseSQLDump(sql1)
	if err != nil {
		t.Fatalf("Failed to parse old SQL: %v", err)
	}
	newTables, err := parser.ParseSQLDump(sql2)
	if err != nil {
		t.Fatalf("Failed to parse new SQL: %v", err)
	}

	oldTable := oldTables[0]
	newTable := newTables[0]

	analyzer := NewTableDiffAnalyzer()
	diff := analyzer.CompareTables(oldTable, newTable)

	if !diff.HasChanges() {
		t.Error("Expected changes for unique constraint change")
	}
	if diff.ColumnsModified != 1 {
		t.Errorf("Expected 1 column modified, got %d", diff.ColumnsModified)
	}

	colDiff := diff.ColumnDiffs[0]
	if colDiff.Changes.Unique == nil {
		t.Error("Expected unique change in column diff")
	}

	uniqueChange := colDiff.Changes.Unique
	if uniqueChange.Old != false {
		t.Errorf("Expected old unique to be false, got %v", uniqueChange.Old)
	}
	if uniqueChange.New != true {
		t.Errorf("Expected new unique to be true, got %v", uniqueChange.New)
	}
}
