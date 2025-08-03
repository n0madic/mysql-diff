package parser

import (
	"strings"
)

// ParseSQLDump parses a SQL dump containing multiple CREATE TABLE statements
func ParseSQLDump(sql string) ([]*CreateTableStatement, error) {
	lexer := NewMySQLLexer(sql)
	tokens := lexer.Tokenize()

	var tables []*CreateTableStatement
	var currentTokens []Token

	// Process all tokens
	for _, token := range tokens {
		// Skip MySQL directives and comments
		if token.Type == MYSQL_DIRECTIVE || token.Type == SQL_COMMENT {
			continue
		}

		// Start new statement on CREATE
		if token.Type == CREATE {
			// Finish previous statement if exists
			if len(currentTokens) > 0 {
				if isCreateTable(currentTokens) {
					if table := parseTokens(currentTokens); table != nil {
						tables = append(tables, table)
					}
				}
				currentTokens = nil
			}
		}

		// Add non-EOF tokens to current statement
		if token.Type != EOF {
			currentTokens = append(currentTokens, token)
		}

		// End statement on semicolon or EOF
		if token.Type == SEMICOLON || token.Type == EOF {
			if len(currentTokens) > 0 {
				if isCreateTable(currentTokens) {
					if table := parseTokens(currentTokens); table != nil {
						tables = append(tables, table)
					}
				}
				currentTokens = nil
			}
		}
	}

	// Handle remaining tokens
	if len(currentTokens) > 0 && isCreateTable(currentTokens) {
		if table := parseTokens(currentTokens); table != nil {
			tables = append(tables, table)
		}
	}

	return tables, nil
}

func isCreateTable(tokens []Token) bool {
	if len(tokens) < 2 {
		return false
	}

	if tokens[0].Type == CREATE {
		if tokens[1].Type == TABLE {
			return true
		}
		if tokens[1].Type == TEMPORARY && len(tokens) >= 3 && tokens[2].Type == TABLE {
			return true
		}
	}

	return false
}

func parseTokens(tokens []Token) *CreateTableStatement {
	parser := NewMySQLCreateTableParser(tokens)
	table, err := parser.Parse()
	if err != nil {
		return nil
	}
	return table
}

// Helper function to clean table name (remove backticks)
func cleanTableName(name string) string {
	return strings.Trim(name, "`")
}

// Helper function to check if a string is quoted
func isQuoted(s string) bool {
	return (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "`") && strings.HasSuffix(s, "`"))
}

// Helper function to unquote a string
func unquote(s string) string {
	if isQuoted(s) {
		return s[1 : len(s)-1]
	}
	return s
}
