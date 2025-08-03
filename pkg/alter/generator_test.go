package alter

import (
	"strings"
	"testing"

	"github.com/n0madic/mysql-diff/pkg/diff"
	"github.com/n0madic/mysql-diff/pkg/parser"
)

func TestStatementGenerator_GenerateAlterStatements(t *testing.T) {
	generator := NewStatementGenerator()

	// Test basic column addition
	oldTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "users",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
			{Name: "name", DataType: parser.DataType{Name: "VARCHAR", Parameters: []string{"255"}}},
		},
	}

	analyzer := diff.NewTableDiffAnalyzer()
	tableDiff := analyzer.CompareTables(oldTable, newTable)

	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(statements))
	}

	expected := "ADD COLUMN `name` VARCHAR(255)"
	if !strings.Contains(statements[0], expected) {
		t.Errorf("Expected statement to contain '%s', got: %s", expected, statements[0])
	}
}

func TestStatementGenerator_TableRename(t *testing.T) {
	generator := NewStatementGenerator()

	oldTable := &parser.CreateTableStatement{
		TableName: "old_users",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
		},
	}

	newTable := &parser.CreateTableStatement{
		TableName: "new_users",
		Columns: []parser.ColumnDefinition{
			{Name: "id", DataType: parser.DataType{Name: "INT"}},
		},
	}

	analyzer := diff.NewTableDiffAnalyzer()
	tableDiff := analyzer.CompareTables(oldTable, newTable)

	statements := generator.GenerateAlterStatements(tableDiff)

	if len(statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(statements))
	}

	expected := "ALTER TABLE `old_users` RENAME TO `new_users`;"
	if statements[0] != expected {
		t.Errorf("Expected '%s', got: %s", expected, statements[0])
	}
}

func TestFormatColumnDefinition(t *testing.T) {
	generator := NewStatementGenerator()

	tests := []struct {
		name     string
		column   *parser.ColumnDefinition
		expected string
	}{
		{
			name: "Basic column",
			column: &parser.ColumnDefinition{
				Name:     "id",
				DataType: parser.DataType{Name: "INT"},
			},
			expected: "`id` INT",
		},
		{
			name: "Column with parameters",
			column: &parser.ColumnDefinition{
				Name:     "name",
				DataType: parser.DataType{Name: "VARCHAR", Parameters: []string{"255"}},
			},
			expected: "`name` VARCHAR(255)",
		},
		{
			name: "Column with NOT NULL",
			column: &parser.ColumnDefinition{
				Name:     "email",
				DataType: parser.DataType{Name: "VARCHAR", Parameters: []string{"255"}},
				Nullable: boolPtr(false),
			},
			expected: "`email` VARCHAR(255) NOT NULL",
		},
		{
			name: "Column with AUTO_INCREMENT",
			column: &parser.ColumnDefinition{
				Name:          "id",
				DataType:      parser.DataType{Name: "INT"},
				AutoIncrement: true,
			},
			expected: "`id` INT AUTO_INCREMENT",
		},
		{
			name: "Column with default value",
			column: &parser.ColumnDefinition{
				Name:         "status",
				DataType:     parser.DataType{Name: "VARCHAR", Parameters: []string{"20"}},
				DefaultValue: stringPtr("active"),
			},
			expected: "`status` VARCHAR(20) DEFAULT 'active'",
		},
		{
			name: "Column with CURRENT_TIMESTAMP default",
			column: &parser.ColumnDefinition{
				Name:         "created_at",
				DataType:     parser.DataType{Name: "TIMESTAMP"},
				DefaultValue: stringPtr("CURRENT_TIMESTAMP"),
			},
			expected: "`created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		},
		{
			name: "Column with UNSIGNED",
			column: &parser.ColumnDefinition{
				Name:     "age",
				DataType: parser.DataType{Name: "INT", Unsigned: true},
			},
			expected: "`age` INT UNSIGNED",
		},
		{
			name: "Column with character set and collation",
			column: &parser.ColumnDefinition{
				Name:         "description",
				DataType:     parser.DataType{Name: "TEXT"},
				CharacterSet: stringPtr("utf8mb4"),
				Collation:    stringPtr("utf8mb4_unicode_ci"),
			},
			expected: "`description` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		},
		{
			name: "Column with comment",
			column: &parser.ColumnDefinition{
				Name:     "notes",
				DataType: parser.DataType{Name: "TEXT"},
				Comment:  stringPtr("User notes"),
			},
			expected: "`notes` TEXT COMMENT 'User notes'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.formatColumnDefinition(tt.column)
			if result != tt.expected {
				t.Errorf("Expected '%s', got: '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatPrimaryKeyDefinition(t *testing.T) {
	generator := NewStatementGenerator()

	tests := []struct {
		name     string
		pk       *parser.PrimaryKeyDefinition
		expected string
	}{
		{
			name: "Single column primary key",
			pk: &parser.PrimaryKeyDefinition{
				Columns: []parser.IndexColumn{
					{Name: "id"},
				},
			},
			expected: "PRIMARY KEY (`id`)",
		},
		{
			name: "Composite primary key",
			pk: &parser.PrimaryKeyDefinition{
				Columns: []parser.IndexColumn{
					{Name: "user_id"},
					{Name: "role_id"},
				},
			},
			expected: "PRIMARY KEY (`user_id`, `role_id`)",
		},
		{
			name: "Named primary key",
			pk: &parser.PrimaryKeyDefinition{
				Name: stringPtr("pk_users"),
				Columns: []parser.IndexColumn{
					{Name: "id"},
				},
			},
			expected: "CONSTRAINT `pk_users` PRIMARY KEY (`id`)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.formatPrimaryKeyDefinition(tt.pk)
			if result != tt.expected {
				t.Errorf("Expected '%s', got: '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatIndexDefinition(t *testing.T) {
	generator := NewStatementGenerator()

	tests := []struct {
		name     string
		index    *parser.IndexDefinition
		expected string
	}{
		{
			name: "Basic index",
			index: &parser.IndexDefinition{
				Name: stringPtr("idx_name"),
				Columns: []parser.IndexColumn{
					{Name: "name"},
				},
			},
			expected: "INDEX `idx_name` (`name`)",
		},
		{
			name: "Unique index",
			index: &parser.IndexDefinition{
				Name:      stringPtr("idx_email"),
				IndexType: "UNIQUE",
				Columns: []parser.IndexColumn{
					{Name: "email"},
				},
			},
			expected: "UNIQUE INDEX `idx_email` (`email`)",
		},
		{
			name: "Composite index",
			index: &parser.IndexDefinition{
				Name: stringPtr("idx_name_email"),
				Columns: []parser.IndexColumn{
					{Name: "name"},
					{Name: "email"},
				},
			},
			expected: "INDEX `idx_name_email` (`name`, `email`)",
		},
		{
			name: "Index with length",
			index: &parser.IndexDefinition{
				Name: stringPtr("idx_description"),
				Columns: []parser.IndexColumn{
					{Name: "description", Length: intPtr(100)},
				},
			},
			expected: "INDEX `idx_description` (`description`(100))",
		},
		{
			name: "Fulltext index",
			index: &parser.IndexDefinition{
				Name:      stringPtr("idx_fulltext"),
				IndexType: "FULLTEXT",
				Columns: []parser.IndexColumn{
					{Name: "content"},
				},
			},
			expected: "FULLTEXT INDEX `idx_fulltext` (`content`)",
		},
		{
			name: "Index with options",
			index: &parser.IndexDefinition{
				Name: stringPtr("idx_complex"),
				Columns: []parser.IndexColumn{
					{Name: "name"},
				},
				Using:   stringPtr("BTREE"),
				Comment: stringPtr("Index comment"),
			},
			expected: "INDEX `idx_complex` (`name`) USING BTREE COMMENT 'Index comment'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.formatIndexDefinition(tt.index)
			if result != tt.expected {
				t.Errorf("Expected '%s', got: '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatForeignKeyDefinition(t *testing.T) {
	generator := NewStatementGenerator()

	tests := []struct {
		name     string
		fk       *parser.ForeignKeyDefinition
		expected string
	}{
		{
			name: "Basic foreign key",
			fk: &parser.ForeignKeyDefinition{
				Columns: []string{"user_id"},
				Reference: parser.ForeignKeyReference{
					TableName: "users",
					Columns:   []string{"id"},
				},
			},
			expected: "FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)",
		},
		{
			name: "Named foreign key",
			fk: &parser.ForeignKeyDefinition{
				Name:    stringPtr("fk_user"),
				Columns: []string{"user_id"},
				Reference: parser.ForeignKeyReference{
					TableName: "users",
					Columns:   []string{"id"},
				},
			},
			expected: "CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)",
		},
		{
			name: "Foreign key with referential actions",
			fk: &parser.ForeignKeyDefinition{
				Columns: []string{"user_id"},
				Reference: parser.ForeignKeyReference{
					TableName: "users",
					Columns:   []string{"id"},
					OnDelete:  stringPtr("CASCADE"),
					OnUpdate:  stringPtr("RESTRICT"),
				},
			},
			expected: "FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT",
		},
		{
			name: "Composite foreign key",
			fk: &parser.ForeignKeyDefinition{
				Columns: []string{"tenant_id", "user_id"},
				Reference: parser.ForeignKeyReference{
					TableName: "tenant_users",
					Columns:   []string{"tenant_id", "user_id"},
				},
			},
			expected: "FOREIGN KEY (`tenant_id`, `user_id`) REFERENCES `tenant_users` (`tenant_id`, `user_id`)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.formatForeignKeyDefinition(tt.fk)
			if result != tt.expected {
				t.Errorf("Expected '%s', got: '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateTableOptionsChanges(t *testing.T) {
	generator := NewStatementGenerator()

	tests := []struct {
		name        string
		tableName   string
		optionsDiff *diff.TableOptionsDiff
		expected    string
	}{
		{
			name:      "Engine change",
			tableName: "users",
			optionsDiff: &diff.TableOptionsDiff{
				ChangeType: diff.ChangeTypeModified,
				NewOptions: &parser.TableOptions{
					Engine: stringPtr("InnoDB"),
				},
			},
			expected: "ALTER TABLE `users` ENGINE=InnoDB;",
		},
		{
			name:      "Multiple options change",
			tableName: "products",
			optionsDiff: &diff.TableOptionsDiff{
				ChangeType: diff.ChangeTypeModified,
				NewOptions: &parser.TableOptions{
					Engine:       stringPtr("InnoDB"),
					CharacterSet: stringPtr("utf8mb4"),
					Comment:      stringPtr("Product table"),
				},
			},
			expected: "ALTER TABLE `products` ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Product table';",
		},
		{
			name:      "No options",
			tableName: "empty",
			optionsDiff: &diff.TableOptionsDiff{
				ChangeType: diff.ChangeTypeModified,
				NewOptions: &parser.TableOptions{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.generateTableOptionsChanges(tt.tableName, tt.optionsDiff)
			if result != tt.expected {
				t.Errorf("Expected '%s', got: '%s'", tt.expected, result)
			}
		})
	}
}

func TestMatchTablesByName(t *testing.T) {
	oldTables := []*parser.CreateTableStatement{
		{TableName: "users"},
		{TableName: "products"},
	}

	newTables := []*parser.CreateTableStatement{
		{TableName: "users"},
		{TableName: "orders"},
	}

	matches := MatchTablesByName(oldTables, newTables)

	// Should have 3 entries: users (both), products (old only), orders (new only)
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	// Check users table (exists in both)
	if match, ok := matches["users"]; !ok {
		t.Error("Expected users table in matches")
	} else if match.Old == nil || match.New == nil {
		t.Error("Users table should exist in both old and new")
	}

	// Check products table (old only)
	if match, ok := matches["products"]; !ok {
		t.Error("Expected products table in matches")
	} else if match.Old == nil || match.New != nil {
		t.Error("Products table should exist only in old schema")
	}

	// Check orders table (new only)
	if match, ok := matches["orders"]; !ok {
		t.Error("Expected orders table in matches")
	} else if match.Old != nil || match.New == nil {
		t.Error("Orders table should exist only in new schema")
	}
}

func TestGenerateDropTableStatements(t *testing.T) {
	oldTables := []*parser.CreateTableStatement{
		{TableName: "users"},
		{TableName: "products"},
		{TableName: "orders"},
	}

	existingNames := map[string]bool{
		"users":  true,
		"orders": true,
	}

	statements := GenerateDropTableStatements(oldTables, existingNames)

	if len(statements) != 1 {
		t.Errorf("Expected 1 drop statement, got %d", len(statements))
	}

	expected := "DROP TABLE IF EXISTS `products`;"
	if statements[0] != expected {
		t.Errorf("Expected '%s', got: '%s'", expected, statements[0])
	}
}

func TestGenerateCreateTableStatements(t *testing.T) {
	newTables := []*parser.CreateTableStatement{
		{TableName: "users"},
		{TableName: "products"},
		{TableName: "orders"},
	}

	existingNames := map[string]bool{
		"users":    true,
		"products": true,
	}

	statements := GenerateCreateTableStatements(newTables, existingNames)

	if len(statements) != 1 {
		t.Errorf("Expected 1 create statement, got %d", len(statements))
	}

	expected := "-- CREATE TABLE `orders` (...); -- New table, full definition needed"
	if statements[0] != expected {
		t.Errorf("Expected '%s', got: '%s'", expected, statements[0])
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
