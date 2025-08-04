package parser

import (
	"testing"
)

func TestBasicTableCreation(t *testing.T) {
	sql := "CREATE TABLE users (id INT, name VARCHAR(255))"
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if table.TableName != "users" {
		t.Errorf("Expected table name 'users', got '%s'", table.TableName)
	}

	if len(table.Columns) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(table.Columns))
	}

	// Check first column
	if table.Columns[0].Name != "id" {
		t.Errorf("Expected first column name 'id', got '%s'", table.Columns[0].Name)
	}
	if table.Columns[0].DataType.Name != "INT" {
		t.Errorf("Expected first column type 'INT', got '%s'", table.Columns[0].DataType.Name)
	}

	// Check second column
	if table.Columns[1].Name != "name" {
		t.Errorf("Expected second column name 'name', got '%s'", table.Columns[1].Name)
	}
	if table.Columns[1].DataType.Name != "VARCHAR" {
		t.Errorf("Expected second column type 'VARCHAR', got '%s'", table.Columns[1].DataType.Name)
	}
	if len(table.Columns[1].DataType.Parameters) != 1 || table.Columns[1].DataType.Parameters[0] != "255" {
		t.Errorf("Expected VARCHAR parameter '255', got %v", table.Columns[1].DataType.Parameters)
	}
}

func TestDataTypesWithParameters(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id BIGINT,
		amount DECIMAL(10,2),
		description TEXT,
		status ENUM('active', 'inactive'),
		created_at TIMESTAMP
	)
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]

	// Check BIGINT
	if table.Columns[0].DataType.Name != "BIGINT" {
		t.Errorf("Expected BIGINT, got %s", table.Columns[0].DataType.Name)
	}

	// Check DECIMAL with parameters
	if table.Columns[1].DataType.Name != "DECIMAL" {
		t.Errorf("Expected DECIMAL, got %s", table.Columns[1].DataType.Name)
	}
	if len(table.Columns[1].DataType.Parameters) != 2 {
		t.Errorf("Expected 2 DECIMAL parameters, got %d", len(table.Columns[1].DataType.Parameters))
	}
	if table.Columns[1].DataType.Parameters[0] != "10" || table.Columns[1].DataType.Parameters[1] != "2" {
		t.Errorf("Expected DECIMAL parameters ['10', '2'], got %v", table.Columns[1].DataType.Parameters)
	}

	// Check TEXT
	if table.Columns[2].DataType.Name != "TEXT" {
		t.Errorf("Expected TEXT, got %s", table.Columns[2].DataType.Name)
	}

	// Check ENUM
	if table.Columns[3].DataType.Name != "ENUM" {
		t.Errorf("Expected ENUM, got %s", table.Columns[3].DataType.Name)
	}
	if len(table.Columns[3].DataType.Parameters) != 2 {
		t.Errorf("Expected 2 ENUM parameters, got %d", len(table.Columns[3].DataType.Parameters))
	}

	// Check TIMESTAMP
	if table.Columns[4].DataType.Name != "TIMESTAMP" {
		t.Errorf("Expected TIMESTAMP, got %s", table.Columns[4].DataType.Name)
	}
}

func TestColumnModifiers(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id INT NOT NULL AUTO_INCREMENT,
		email VARCHAR(255) UNIQUE,
		name VARCHAR(100) DEFAULT 'Unknown',
		count INT UNSIGNED
	)
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	// Check NOT NULL and AUTO_INCREMENT
	idCol := table.Columns[0]
	if idCol.Nullable == nil || *idCol.Nullable != false {
		t.Errorf("Expected id column to be NOT NULL")
	}
	if !idCol.AutoIncrement {
		t.Errorf("Expected id column to be AUTO_INCREMENT")
	}

	// Check UNIQUE
	emailCol := table.Columns[1]
	if !emailCol.Unique {
		t.Errorf("Expected email column to be UNIQUE")
	}

	// Check DEFAULT
	nameCol := table.Columns[2]
	if nameCol.DefaultValue == nil || *nameCol.DefaultValue != "'Unknown'" {
		t.Errorf("Expected name column default value to be 'Unknown', got %v", nameCol.DefaultValue)
	}

	// Check UNSIGNED
	countCol := table.Columns[3]
	if !countCol.DataType.Unsigned {
		t.Errorf("Expected count column to be UNSIGNED")
	}
}

func TestPrimaryKey(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id INT,
		name VARCHAR(100),
		PRIMARY KEY (id)
	)
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	if table.PrimaryKey == nil {
		t.Fatalf("Expected primary key to be defined")
	}

	if len(table.PrimaryKey.Columns) != 1 {
		t.Errorf("Expected 1 primary key column, got %d", len(table.PrimaryKey.Columns))
	}

	if table.PrimaryKey.Columns[0].Name != "id" {
		t.Errorf("Expected primary key column 'id', got '%s'", table.PrimaryKey.Columns[0].Name)
	}
}

func TestCompositePrimaryKey(t *testing.T) {
	sql := `
	CREATE TABLE test (
		user_id INT,
		role_id INT,
		name VARCHAR(100),
		PRIMARY KEY (user_id, role_id)
	)
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	if table.PrimaryKey == nil {
		t.Fatalf("Expected primary key to be defined")
	}

	if len(table.PrimaryKey.Columns) != 2 {
		t.Errorf("Expected 2 primary key columns, got %d", len(table.PrimaryKey.Columns))
	}

	if table.PrimaryKey.Columns[0].Name != "user_id" {
		t.Errorf("Expected first primary key column 'user_id', got '%s'", table.PrimaryKey.Columns[0].Name)
	}

	if table.PrimaryKey.Columns[1].Name != "role_id" {
		t.Errorf("Expected second primary key column 'role_id', got '%s'", table.PrimaryKey.Columns[1].Name)
	}
}

func TestIndexes(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id INT,
		email VARCHAR(255),
		name VARCHAR(100),
		INDEX idx_email (email),
		UNIQUE KEY unique_name (name)
	)
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	if len(table.Indexes) != 2 {
		t.Errorf("Expected 2 indexes, got %d", len(table.Indexes))
	}

	// Check regular index
	regularIndex := table.Indexes[0]
	if regularIndex.IndexType != "INDEX" {
		t.Errorf("Expected INDEX type, got %s", regularIndex.IndexType)
	}
	if regularIndex.Name == nil || *regularIndex.Name != "idx_email" {
		t.Errorf("Expected index name 'idx_email', got %v", regularIndex.Name)
	}
	if len(regularIndex.Columns) != 1 || regularIndex.Columns[0].Name != "email" {
		t.Errorf("Expected index on 'email' column")
	}

	// Check unique index
	uniqueIndex := table.Indexes[1]
	if uniqueIndex.IndexType != "UNIQUE" {
		t.Errorf("Expected UNIQUE type, got %s", uniqueIndex.IndexType)
	}
	if uniqueIndex.Name == nil || *uniqueIndex.Name != "unique_name" {
		t.Errorf("Expected index name 'unique_name', got %v", uniqueIndex.Name)
	}
}

func TestForeignKey(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id INT,
		user_id INT,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	)
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	if len(table.ForeignKeys) != 1 {
		t.Errorf("Expected 1 foreign key, got %d", len(table.ForeignKeys))
	}

	fk := table.ForeignKeys[0]
	if len(fk.Columns) != 1 || fk.Columns[0] != "user_id" {
		t.Errorf("Expected foreign key on 'user_id' column")
	}

	if fk.Reference.TableName != "users" {
		t.Errorf("Expected reference to 'users' table, got '%s'", fk.Reference.TableName)
	}

	if len(fk.Reference.Columns) != 1 || fk.Reference.Columns[0] != "id" {
		t.Errorf("Expected reference to 'id' column")
	}

	if fk.Reference.OnDelete == nil || *fk.Reference.OnDelete != "CASCADE" {
		t.Errorf("Expected ON DELETE CASCADE")
	}
}

func TestTableOptions(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id INT,
		name VARCHAR(100)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1000
	`
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	if table.TableOptions == nil {
		t.Fatalf("Expected table options to be defined")
	}

	if table.TableOptions.Engine == nil || *table.TableOptions.Engine != "InnoDB" {
		t.Errorf("Expected ENGINE=InnoDB")
	}

	if table.TableOptions.CharacterSet == nil || *table.TableOptions.CharacterSet != "utf8mb4" {
		t.Errorf("Expected DEFAULT CHARSET=utf8mb4")
	}

	if table.TableOptions.AutoIncrement == nil || *table.TableOptions.AutoIncrement != 1000 {
		t.Errorf("Expected AUTO_INCREMENT=1000")
	}
}

func TestTemporaryTable(t *testing.T) {
	sql := "CREATE TEMPORARY TABLE temp_users (id INT, name VARCHAR(255))"
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if !table.Temporary {
		t.Errorf("Expected temporary table")
	}
}

func TestIfNotExists(t *testing.T) {
	sql := "CREATE TABLE IF NOT EXISTS users (id INT, name VARCHAR(255))"
	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if !table.IfNotExists {
		t.Errorf("Expected IF NOT EXISTS")
	}
}

func TestMultipleTables(t *testing.T) {
	sql := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT,
		name VARCHAR(255),
		PRIMARY KEY (id)
	);
	
	CREATE TABLE posts (
		id INT AUTO_INCREMENT,
		user_id INT,
		title VARCHAR(255),
		PRIMARY KEY (id),
		FOREIGN KEY (user_id) REFERENCES users (id)
	);
	`

	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	if len(tables) != 2 {
		t.Errorf("Expected 2 tables, got %d", len(tables))
	}

	if tables[0].TableName != "users" {
		t.Errorf("Expected first table to be 'users', got '%s'", tables[0].TableName)
	}

	if tables[1].TableName != "posts" {
		t.Errorf("Expected second table to be 'posts', got '%s'", tables[1].TableName)
	}
}

func TestComplexDataTypes(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		price DECIMAL(10,2) UNSIGNED,
		flags BIT(8),
		data JSON,
		location POINT,
		status ENUM('active', 'inactive', 'pending'),
		tags SET('tag1', 'tag2', 'tag3')
	)
	`

	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	// Check BIGINT UNSIGNED
	if !table.Columns[0].DataType.Unsigned {
		t.Errorf("Expected id column to be UNSIGNED")
	}

	// Check DECIMAL UNSIGNED
	if !table.Columns[1].DataType.Unsigned {
		t.Errorf("Expected price column to be UNSIGNED")
	}

	// Check BIT with parameter
	if table.Columns[2].DataType.Name != "BIT" {
		t.Errorf("Expected BIT type")
	}
	if len(table.Columns[2].DataType.Parameters) != 1 || table.Columns[2].DataType.Parameters[0] != "8" {
		t.Errorf("Expected BIT(8)")
	}

	// Check JSON
	if table.Columns[3].DataType.Name != "JSON" {
		t.Errorf("Expected JSON type")
	}

	// Check POINT (spatial type)
	if table.Columns[4].DataType.Name != "POINT" {
		t.Errorf("Expected POINT type")
	}

	// Check ENUM
	if table.Columns[5].DataType.Name != "ENUM" {
		t.Errorf("Expected ENUM type")
	}

	// Check SET
	if table.Columns[6].DataType.Name != "SET" {
		t.Errorf("Expected SET type")
	}
}

func TestFulltextAndSpatialIndexes(t *testing.T) {
	sql := `
	CREATE TABLE test (
		id INT,
		content TEXT,
		location POINT,
		FULLTEXT KEY ft_content (content),
		SPATIAL KEY sp_location (location)
	)
	`

	tables, err := ParseSQLDump(sql)

	if err != nil {
		t.Fatalf("ParseSQLDump failed: %v", err)
	}

	table := tables[0]

	if len(table.Indexes) != 2 {
		t.Errorf("Expected 2 indexes, got %d", len(table.Indexes))
	}

	// Check FULLTEXT index
	ftIndex := table.Indexes[0]
	if ftIndex.IndexType != "FULLTEXT" {
		t.Errorf("Expected FULLTEXT index type, got %s", ftIndex.IndexType)
	}

	// Check SPATIAL index
	spIndex := table.Indexes[1]
	if spIndex.IndexType != "SPATIAL" {
		t.Errorf("Expected SPATIAL index type, got %s", spIndex.IndexType)
	}
}
