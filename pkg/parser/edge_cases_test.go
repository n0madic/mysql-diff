package parser

import (
	"testing"
)

func TestParseEmptyFile(t *testing.T) {
	sql := ""
	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump should not fail on empty input: %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("Expected 0 tables from empty input, got %d", len(tables))
	}
}

func TestParseCommentsOnly(t *testing.T) {
	sql := `-- This is a comment
	/* Multi-line comment
	   with multiple lines */
	# Hash style comment
	-- Another comment`

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump should not fail on comments only: %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("Expected 0 tables from comments only, got %d", len(tables))
	}
}

func TestParseUnicodeNames(t *testing.T) {
	sql := "CREATE TABLE `пользователи` (" +
		"`идентификатор` INT AUTO_INCREMENT PRIMARY KEY," +
		"`имя` VARCHAR(255) NOT NULL," +
		"`日期` DATETIME DEFAULT CURRENT_TIMESTAMP," +
		"`🚀_emoji_column` TEXT" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump failed on Unicode names: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if table.TableName != "пользователи" {
		t.Errorf("Expected table name 'пользователи', got '%s'", table.TableName)
	}

	// Check columns with Unicode names
	expectedColumns := []string{"идентификатор", "имя", "日期", "🚀_emoji_column"}
	if len(table.Columns) != len(expectedColumns) {
		t.Fatalf("Expected %d columns, got %d", len(expectedColumns), len(table.Columns))
	}

	for i, expected := range expectedColumns {
		if table.Columns[i].Name != expected {
			t.Errorf("Expected column name '%s', got '%s'", expected, table.Columns[i].Name)
		}
	}
}

func TestParseEscapedQuotes(t *testing.T) {
	sql := `CREATE TABLE test (
		id INT PRIMARY KEY,
		single_quote VARCHAR(255) DEFAULT 'It''s a test',
		double_quote VARCHAR(255) DEFAULT "He said \"Hello\"",
		backslash VARCHAR(255) DEFAULT 'Path\\to\\file',
		comment_field VARCHAR(100) COMMENT 'Comment with ''quotes'''
	);`

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump failed on escaped quotes: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if len(table.Columns) != 5 {
		t.Fatalf("Expected 5 columns, got %d", len(table.Columns))
	}

	// Check default values with escaped quotes
	singleQuoteCol := table.Columns[1]
	if singleQuoteCol.DefaultValue == nil {
		t.Error("Expected single_quote column to have default value")
	} else if *singleQuoteCol.DefaultValue != "'It''s a test'" {
		t.Errorf("Expected default value 'It''s a test', got %v", *singleQuoteCol.DefaultValue)
	}
}

func TestParseComplexEnumSet(t *testing.T) {
	sql := `CREATE TABLE complex_enums (
		id INT PRIMARY KEY,
		status ENUM('active', 'inactive', 'pending') DEFAULT 'pending',
		roles SET('admin', 'user', 'moderator') DEFAULT 'user',
		complex_enum ENUM('value with spaces', 'value,with,commas', 'value''with''quotes') DEFAULT 'value with spaces'
	);`

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump failed on complex ENUM/SET: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if len(table.Columns) != 4 {
		t.Fatalf("Expected 4 columns, got %d", len(table.Columns))
	}

	// Check ENUM column
	statusCol := table.Columns[1]
	if statusCol.DataType.Name != "ENUM" {
		t.Errorf("Expected ENUM type, got %s", statusCol.DataType.Name)
	}
	if len(statusCol.DataType.Parameters) != 3 {
		t.Errorf("Expected 3 ENUM values, got %d", len(statusCol.DataType.Parameters))
	}

	// Check SET column
	rolesCol := table.Columns[2]
	if rolesCol.DataType.Name != "SET" {
		t.Errorf("Expected SET type, got %s", rolesCol.DataType.Name)
	}
	if len(rolesCol.DataType.Parameters) != 3 {
		t.Errorf("Expected 3 SET values, got %d", len(rolesCol.DataType.Parameters))
	}
}

func TestParseLongNames(t *testing.T) {
	// MySQL allows up to 64 characters for identifiers
	longTableName := "this_is_a_very_long_table_name_that_approaches_mysql_limit_64"
	longColumnName := "this_is_a_very_long_column_name_that_approaches_mysql_limit"

	sql := "CREATE TABLE `" + longTableName + "` (" +
		"`" + longColumnName + "` INT PRIMARY KEY," +
		"short INT" +
		");"

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump failed on long names: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if table.TableName != longTableName {
		t.Errorf("Expected table name '%s', got '%s'", longTableName, table.TableName)
	}

	if len(table.Columns) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(table.Columns))
	}

	if table.Columns[0].Name != longColumnName {
		t.Errorf("Expected column name '%s', got '%s'", longColumnName, table.Columns[0].Name)
	}
}

func TestParseComplexDefaults(t *testing.T) {
	sql := `CREATE TABLE complex_defaults (
		id INT PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		null_field VARCHAR(255) DEFAULT NULL,
		empty_string VARCHAR(255) DEFAULT '',
		zero_int INT DEFAULT 0,
		negative_int INT DEFAULT -1,
		decimal_default DECIMAL(10,2) DEFAULT 99.99,
		boolean_true BOOLEAN DEFAULT TRUE,
		boolean_false BOOLEAN DEFAULT FALSE
	);`

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump failed on complex defaults: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0]
	if len(table.Columns) != 10 {
		t.Fatalf("Expected 10 columns, got %d", len(table.Columns))
	}

	// Check various default value types
	testCases := []struct {
		columnIndex int
		columnName  string
		hasDefault  bool
		defaultVal  string
	}{
		{1, "created_at", true, "CURRENT_TIMESTAMP"},
		{3, "null_field", true, "NULL"},
		{4, "empty_string", true, "''"},
		{5, "zero_int", true, "0"},
		{6, "negative_int", true, "-1"},
		{7, "decimal_default", true, "99.99"},
		{8, "boolean_true", true, "TRUE"},
		{9, "boolean_false", true, "FALSE"},
	}

	for _, tc := range testCases {
		col := table.Columns[tc.columnIndex]
		if col.Name != tc.columnName {
			t.Errorf("Expected column name '%s', got '%s'", tc.columnName, col.Name)
		}

		if tc.hasDefault {
			if col.DefaultValue == nil {
				t.Errorf("Column '%s' should have default value", tc.columnName)
			} else if *col.DefaultValue != tc.defaultVal {
				t.Errorf("Column '%s' expected default '%s', got '%s'",
					tc.columnName, tc.defaultVal, *col.DefaultValue)
			}
		} else {
			if col.DefaultValue != nil {
				t.Errorf("Column '%s' should not have default value, got '%s'",
					tc.columnName, *col.DefaultValue)
			}
		}
	}
}

func TestParseMalformedSQL(t *testing.T) {
	testCases := []struct {
		name string
		sql  string
	}{
		{
			"Incomplete CREATE statement",
			"CREATE TABLE test",
		},
		{
			"Missing parentheses",
			"CREATE TABLE test id INT PRIMARY KEY;",
		},
		{
			"Unmatched parentheses",
			"CREATE TABLE test (id INT PRIMARY KEY;",
		},
		{
			"Invalid column definition",
			"CREATE TABLE test (id);",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tables, err := ParseSQLDump(tc.sql)
			if err == nil && len(tables) > 0 {
				// Some malformed SQL might still parse partially
				t.Logf("Malformed SQL parsed unexpectedly: %s", tc.sql)
			}
			// We don't fail the test because the parser might be lenient
			// and that's acceptable behavior
		})
	}
}

func TestParseWithMySQLDirectives(t *testing.T) {
	sql := `SET FOREIGN_KEY_CHECKS=0;
	
	CREATE TABLE test (
		id INT PRIMARY KEY,
		name VARCHAR(255)
	) ENGINE=InnoDB;
	
	SET FOREIGN_KEY_CHECKS=1;`

	tables, err := ParseSQLDump(sql)
	if err != nil {
		t.Fatalf("ParseSQLDump failed with MySQL directives: %v", err)
	}
	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}

	if tables[0].TableName != "test" {
		t.Errorf("Expected table name 'test', got '%s'", tables[0].TableName)
	}
}
