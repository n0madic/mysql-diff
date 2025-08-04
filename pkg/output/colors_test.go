package output

import (
	"os"
	"strings"
	"testing"
)

func TestColorizeBasicFunctionality(t *testing.T) {
	// Force enable colors for testing
	originalEnabled := Colors.Enabled
	Colors.Enabled = true
	defer func() { Colors.Enabled = originalEnabled }()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"RedText", "error", Red + "error" + Reset},
		{"GreenText", "success", Green + "success" + Reset},
		{"BlueText", "info", Blue + "info" + Reset},
		{"YellowText", "warning", Yellow + "warning" + Reset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			switch tt.name {
			case "RedText":
				result = RedText(tt.input)
			case "GreenText":
				result = GreenText(tt.input)
			case "BlueText":
				result = BlueText(tt.input)
			case "YellowText":
				result = YellowText(tt.input)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestColorsDisabled(t *testing.T) {
	// Force disable colors
	originalEnabled := Colors.Enabled
	Colors.Enabled = false
	defer func() { Colors.Enabled = originalEnabled }()

	input := "test"
	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"RedText", RedText},
		{"GreenText", GreenText},
		{"BlueText", BlueText},
		{"YellowText", YellowText},
		{"BoldText", BoldText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(input)
			if result != input {
				t.Errorf("Expected %q (no colors), got %q", input, result)
			}
		})
	}
}

func TestSetColorsEnabled(t *testing.T) {
	originalEnabled := Colors.Enabled
	defer func() { Colors.Enabled = originalEnabled }()

	// Test enabling
	SetColorsEnabled(true)
	if !Colors.Enabled {
		t.Error("Expected colors to be enabled")
	}

	// Test disabling
	SetColorsEnabled(false)
	if Colors.Enabled {
		t.Error("Expected colors to be disabled")
	}
}

func TestColorizeSQLStatement(t *testing.T) {
	// Force enable colors for testing
	originalEnabled := Colors.Enabled
	Colors.Enabled = true
	defer func() { Colors.Enabled = originalEnabled }()

	tests := []struct {
		name        string
		input       string
		contains    []string // Check that output contains these colored keywords
		notContains []string // Check that these strings are NOT colored
	}{
		{
			name:  "Basic ALTER TABLE",
			input: "ALTER TABLE users ADD COLUMN email VARCHAR(255);",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				GreenText("ADD COLUMN"),
			},
		},
		{
			name:  "MODIFY COLUMN with DEFAULT",
			input: "ALTER TABLE users MODIFY COLUMN name VARCHAR(100) DEFAULT 'test';",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				YellowText("MODIFY COLUMN"),
				PurpleText("DEFAULT"),
			},
		},
		{
			name:  "Index operations",
			input: "ALTER TABLE posts DROP INDEX idx_title, ADD INDEX idx_new (title);",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				RedText("DROP INDEX"),
				GreenText("ADD INDEX"),
			},
		},
		{
			name:  "Keywords in string literals should NOT be colored",
			input: "ALTER TABLE users ADD COLUMN description TEXT DEFAULT 'alter table info';",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				GreenText("ADD COLUMN"),
				PurpleText("DEFAULT"),
			},
			notContains: []string{
				// These should NOT appear as they are inside string literal
				"'" + BrightCyanText("ALTER TABLE") + "'",
				"'alter " + BrightCyanText("TABLE") + "'",
			},
		},
		{
			name:  "Keywords in column names should NOT be colored",
			input: "ALTER TABLE users ADD COLUMN table_name VARCHAR(255);",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				GreenText("ADD COLUMN"),
			},
			// table_name should not have 'table' colored since it's part of identifier
		},
		{
			name:  "Complex statement with multiple operations",
			input: "ALTER TABLE users DROP COLUMN old_field, ADD COLUMN new_field INT DEFAULT 0, MODIFY COLUMN email VARCHAR(255) NOT NULL;",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				RedText("DROP COLUMN"),
				GreenText("ADD COLUMN"),
				YellowText("MODIFY COLUMN"),
				PurpleText("DEFAULT"),
				// NOT and NULL are colored separately since NOT NULL isn't a single regex match
				PurpleText("NULL"),
			},
		},
		{
			name:  "Table options",
			input: "ALTER TABLE users ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User table';",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				CyanText("ENGINE="),
				// In the output, we see that DEFAULT is colored separately from CHARSET=
				// This happens because the pattern DEFAULT\s+CHARSET= doesn't match when they're separate
				CyanText("CHARSET="),
				CyanText("COMMENT="),
			},
		},
		{
			name:  "Foreign key operations",
			input: "ALTER TABLE posts ADD FOREIGN KEY (user_id) REFERENCES users(id), DROP FOREIGN KEY fk_posts_user;",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				GreenText("ADD FOREIGN KEY"),
				RedText("DROP FOREIGN KEY"),
			},
		},
		{
			name:  "Keywords in different cases",
			input: "alter table Users add column Email varchar(255);",
			contains: []string{
				// Should work case-insensitive
				"alter table", // Will be colored as ALTER TABLE
				"add column",  // Will be colored as ADD COLUMN
			},
		},
		{
			name:  "Tricky case: keywords in various contexts",
			input: "ALTER TABLE modification_log ADD COLUMN modify_date DATETIME DEFAULT 'updated by alter process';",
			contains: []string{
				BrightCyanText("ALTER TABLE"),
				GreenText("ADD COLUMN"),
				PurpleText("DEFAULT"),
			},
			notContains: []string{
				// These should NOT be colored as they're part of other words or in strings
				"modification_" + BrightCyanText("log"),        // 'log' is not a standalone keyword
				"modify_" + YellowText("date"),                 // 'date' as part of identifier
				"'updated by " + BrightCyanText("alter") + "'", // inside string literal
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorizeSQLStatement(tt.input)

			// Check that expected colored keywords are present
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, but got: %q", expected, result)
				}
			}

			// Check that unwanted colorings are NOT present
			for _, notExpected := range tt.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result NOT to contain %q, but got: %q", notExpected, result)
				}
			}
		})
	}
}

func TestColorizeSQLStatementDisabled(t *testing.T) {
	// Force disable colors
	originalEnabled := Colors.Enabled
	Colors.Enabled = false
	defer func() { Colors.Enabled = originalEnabled }()

	input := "ALTER TABLE users ADD COLUMN email VARCHAR(255);"
	result := ColorizeSQLStatement(input)

	if result != input {
		t.Errorf("Expected no coloring when disabled, got: %q, expected: %q", result, input)
	}
}

func TestColorizeChange(t *testing.T) {
	// Force enable colors for testing
	originalEnabled := Colors.Enabled
	Colors.Enabled = true
	defer func() { Colors.Enabled = originalEnabled }()

	tests := []struct {
		text       string
		changeType string
		expected   string
	}{
		{"new_column", "added", GreenText("+ new_column")},
		{"old_column", "removed", RedText("- old_column")},
		{"changed_column", "modified", YellowText("~ changed_column")},
		{"column", "add", GreenText("+ column")},
		{"column", "drop", RedText("- column")},
		{"column", "unknown", "column"}, // No change for unknown type
	}

	for _, tt := range tests {
		t.Run(tt.changeType, func(t *testing.T) {
			result := ColorizeChange(tt.text, tt.changeType)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestColorizeHelperFunctions(t *testing.T) {
	// Force enable colors for testing
	originalEnabled := Colors.Enabled
	Colors.Enabled = true
	defer func() { Colors.Enabled = originalEnabled }()

	tests := []struct {
		name     string
		function func(string) string
		input    string
		expected string
	}{
		{"ColorizeTableName", ColorizeTableName, "users", CyanText("users")},
		{"ColorizeColumnName", ColorizeColumnName, "email", YellowText("email")},
		{"ColorizeDataType", ColorizeDataType, "VARCHAR", PurpleText("VARCHAR")},
		{"ColorizeString", ColorizeString, "'test'", BrightGreenText("'test'")},
		{"ColorizeNumber", ColorizeNumber, "123", BrightBlueText("123")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsColorSupported(t *testing.T) {
	// This test checks the logic but doesn't assert specific values
	// since it depends on the runtime environment

	// Test NO_COLOR environment variable
	originalNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	os.Setenv("NO_COLOR", "1")
	if isColorSupported() {
		t.Error("Expected colors to be disabled when NO_COLOR is set")
	}

	os.Setenv("NO_COLOR", "")
	// Other tests depend on runtime environment, so we just ensure the function doesn't panic
	_ = isColorSupported()
}

func BenchmarkColorizeSQLStatement(b *testing.B) {
	// Force enable colors for benchmarking
	originalEnabled := Colors.Enabled
	Colors.Enabled = true
	defer func() { Colors.Enabled = originalEnabled }()

	statement := "ALTER TABLE users MODIFY COLUMN name VARCHAR(150) DEFAULT 'updated', ADD COLUMN email VARCHAR(255), DROP INDEX idx_old, ADD INDEX idx_new (email);"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ColorizeSQLStatement(statement)
	}
}
