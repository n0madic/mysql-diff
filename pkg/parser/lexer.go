package parser

import (
	"strings"
	"unicode"
)

// MySQLLexer tokenizes SQL input
type MySQLLexer struct {
	text        []rune
	pos         int
	line        int
	column      int
	currentChar *rune
	keywords    map[string]TokenType
}

// NewMySQLLexer creates a new lexer instance
func NewMySQLLexer(text string) *MySQLLexer {
	runes := []rune(text)
	lexer := &MySQLLexer{
		text:   runes,
		pos:    0,
		line:   1,
		column: 1,
	}

	if len(runes) > 0 {
		lexer.currentChar = &runes[0]
	}

	lexer.keywords = map[string]TokenType{
		"CREATE":             CREATE,
		"TABLE":              TABLE,
		"TEMPORARY":          TEMPORARY,
		"IF":                 IF,
		"NOT":                NOT,
		"EXISTS":             EXISTS,
		"LIKE":               LIKE,
		"AS":                 AS,
		"SELECT":             SELECT,
		"IGNORE":             IGNORE,
		"REPLACE":            REPLACE,
		"INT":                INT,
		"TINYINT":            TINYINT,
		"SMALLINT":           SMALLINT,
		"MEDIUMINT":          MEDIUMINT,
		"BIGINT":             BIGINT,
		"VARCHAR":            VARCHAR,
		"CHAR":               CHAR,
		"TEXT":               TEXT,
		"DECIMAL":            DECIMAL,
		"FLOAT":              FLOAT,
		"DOUBLE":             DOUBLE,
		"DATE":               DATE,
		"DATETIME":           DATETIME,
		"TIMESTAMP":          TIMESTAMP,
		"TIME":               TIME,
		"YEAR":               YEAR,
		"BLOB":               BLOB,
		"JSON":               JSON,
		"ENUM":               ENUM,
		"SET":                SET,
		"BINARY":             BINARY,
		"VARBINARY":          VARBINARY,
		"BIT":                BIT,
		"GEOMETRY":           GEOMETRY,
		"POINT":              POINT,
		"LINESTRING":         LINESTRING,
		"POLYGON":            POLYGON,
		"NULL":               NULL,
		"DEFAULT":            DEFAULT,
		"AUTO_INCREMENT":     AUTO_INCREMENT,
		"UNIQUE":             UNIQUE,
		"PRIMARY":            PRIMARY,
		"KEY":                KEY,
		"COMMENT":            COMMENT,
		"COLLATE":            COLLATE,
		"CHARACTER":          CHARACTER,
		"CHARSET":            CHARSET,
		"VISIBLE":            VISIBLE,
		"INVISIBLE":          INVISIBLE,
		"GENERATED":          GENERATED,
		"ALWAYS":             ALWAYS,
		"VIRTUAL":            VIRTUAL,
		"STORED":             STORED,
		"UNSIGNED":           UNSIGNED,
		"ZEROFILL":           ZEROFILL,
		"INDEX":              INDEX,
		"FULLTEXT":           FULLTEXT,
		"SPATIAL":            SPATIAL,
		"FOREIGN":            FOREIGN,
		"REFERENCES":         REFERENCES,
		"CHECK":              CHECK,
		"CONSTRAINT":         CONSTRAINT,
		"ENGINE":             ENGINE,
		"ROW_FORMAT":         ROW_FORMAT,
		"TABLESPACE":         TABLESPACE,
		"DATA":               DATA,
		"DIRECTORY":          DIRECTORY,
		"COMPRESSION":        COMPRESSION,
		"ENCRYPTION":         ENCRYPTION,
		"KEY_BLOCK_SIZE":     KEY_BLOCK_SIZE,
		"MAX_ROWS":           MAX_ROWS,
		"MIN_ROWS":           MIN_ROWS,
		"STATS_PERSISTENT":   STATS_PERSISTENT,
		"STATS_AUTO_RECALC":  STATS_AUTO_RECALC,
		"STATS_SAMPLE_PAGES": STATS_SAMPLE_PAGES,
		"PACK_KEYS":          PACK_KEYS,
		"CHECKSUM":           CHECKSUM,
		"DELAY_KEY_WRITE":    DELAY_KEY_WRITE,
		"UNION":              UNION,
		"INSERT_METHOD":      INSERT_METHOD,
		"PARTITION":          PARTITION,
		"BY":                 BY,
		"HASH":               HASH,
		"RANGE":              RANGE,
		"LIST":               LIST,
		"COLUMNS":            COLUMNS,
		"VALUES":             VALUES,
		"LESS":               LESS,
		"THAN":               THAN,
		"IN":                 IN,
		"MAXVALUE":           MAXVALUE,
		"LINEAR":             LINEAR,
		"ON":                 ON,
		"DELETE":             DELETE,
		"UPDATE":             UPDATE,
		"CASCADE":            CASCADE,
		"RESTRICT":           RESTRICT,
		"NO":                 NO,
		"ACTION":             ACTION,
		"ASC":                ASC,
		"DESC":               DESC,
		"WITH":               WITH,
		"PARSER":             PARSER,
		"ALGORITHM":          ALGORITHM,
		"LOCK":               LOCK,
		"ENGINE_ATTRIBUTE":   ENGINE_ATTRIBUTE,
		"INPLACE":            INPLACE,
		"NONE":               NONE,
		"FIRST":              FIRST,
		"LAST":               LAST,
		"COLUMN_FORMAT":      COLUMN_FORMAT,
		"FIXED":              FIXED,
		"DYNAMIC":            DYNAMIC,
		"STORAGE":            STORAGE,
		"DISK":               DISK,
		"MEMORY":             MEMORY,
		"COMPRESSED":         COMPRESSED,
		"DROP":               DROP,
		"USE":                USE,
		"DATABASE":           DATABASE,
	}

	return lexer
}

// advance moves to the next character
func (l *MySQLLexer) advance() {
	if l.currentChar != nil && *l.currentChar == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	l.pos++
	if l.pos >= len(l.text) {
		l.currentChar = nil
	} else {
		l.currentChar = &l.text[l.pos]
	}
}

// peek looks ahead at the next character(s) without advancing
func (l *MySQLLexer) peek(offset ...int) *rune {
	off := 1
	if len(offset) > 0 {
		off = offset[0]
	}

	peekPos := l.pos + off
	if peekPos >= len(l.text) {
		return nil
	}
	return &l.text[peekPos]
}

// skipWhitespace skips whitespace characters
func (l *MySQLLexer) skipWhitespace() {
	for l.currentChar != nil && unicode.IsSpace(*l.currentChar) {
		l.advance()
	}
}

// skipComment skips SQL comments
func (l *MySQLLexer) skipComment() bool {
	if l.currentChar == nil {
		return false
	}

	// Single line comment --
	if *l.currentChar == '-' {
		next := l.peek()
		if next != nil && *next == '-' {
			l.advance() // Skip first -
			l.advance() // Skip second -
			for l.currentChar != nil && *l.currentChar != '\n' {
				l.advance()
			}
			return true
		}
	}

	// Single line comment #
	if *l.currentChar == '#' {
		for l.currentChar != nil && *l.currentChar != '\n' {
			l.advance()
		}
		return true
	}

	// Multi-line comment /* */
	if *l.currentChar == '/' {
		next := l.peek()
		if next != nil && *next == '*' {
			l.advance() // Skip /
			l.advance() // Skip *
			for l.currentChar != nil {
				if *l.currentChar == '*' {
					next := l.peek()
					if next != nil && *next == '/' {
						l.advance() // Skip *
						l.advance() // Skip /
						break
					}
				}
				l.advance()
			}
			return true
		}
	}

	return false
}

// readMySQLDirective reads MySQL-specific directives like /*!40101 ... */
func (l *MySQLLexer) readMySQLDirective() string {
	value := ""
	for l.currentChar != nil {
		if *l.currentChar == '*' {
			next := l.peek()
			if next != nil && *next == '/' {
				l.advance() // Skip *
				l.advance() // Skip /
				break
			}
		}
		value += string(*l.currentChar)
		l.advance()
	}
	return value
}

// readString reads quoted strings
func (l *MySQLLexer) readString() string {
	quote := *l.currentChar
	l.advance() // Skip opening quote

	value := ""
	for l.currentChar != nil && *l.currentChar != quote {
		if *l.currentChar == '\\' {
			l.advance()
			if l.currentChar != nil {
				value += string(*l.currentChar)
				l.advance()
			}
		} else {
			value += string(*l.currentChar)
			l.advance()
		}
	}

	if l.currentChar != nil && *l.currentChar == quote {
		l.advance() // Skip closing quote
	}

	return value
}

// readNumber reads numeric literals
func (l *MySQLLexer) readNumber() string {
	value := ""
	for l.currentChar != nil && (unicode.IsDigit(*l.currentChar) || *l.currentChar == '.') {
		value += string(*l.currentChar)
		l.advance()
	}
	return value
}

// readIdentifier reads identifiers
func (l *MySQLLexer) readIdentifier() string {
	value := ""
	for l.currentChar != nil && (unicode.IsLetter(*l.currentChar) || unicode.IsDigit(*l.currentChar) || *l.currentChar == '_' || *l.currentChar == '$') {
		value += string(*l.currentChar)
		l.advance()
	}
	return value
}

// readQuotedIdentifier reads backtick-quoted identifiers
func (l *MySQLLexer) readQuotedIdentifier() string {
	l.advance() // Skip opening backtick

	value := ""
	for l.currentChar != nil && *l.currentChar != '`' {
		value += string(*l.currentChar)
		l.advance()
	}

	if l.currentChar != nil && *l.currentChar == '`' {
		l.advance() // Skip closing backtick
	}

	return value
}

// GetNextToken returns the next token from the input
func (l *MySQLLexer) GetNextToken() Token {
	for l.currentChar != nil {
		if unicode.IsSpace(*l.currentChar) {
			l.skipWhitespace()
			continue
		}

		// Skip comments
		if l.skipComment() {
			continue
		}

		// Handle MySQL directives
		if *l.currentChar == '/' {
			next1 := l.peek(1)
			next2 := l.peek(2)
			if next1 != nil && *next1 == '*' && next2 != nil && *next2 == '!' {
				return Token{
					Type:     MYSQL_DIRECTIVE,
					Value:    l.readMySQLDirective(),
					Position: l.pos,
					Line:     l.line,
					Column:   l.column,
				}
			}
		}

		if *l.currentChar == '"' || *l.currentChar == '\'' {
			return Token{
				Type:     STRING,
				Value:    l.readString(),
				Position: l.pos,
				Line:     l.line,
				Column:   l.column,
			}
		}

		if *l.currentChar == '`' {
			return Token{
				Type:     IDENTIFIER,
				Value:    l.readQuotedIdentifier(),
				Position: l.pos,
				Line:     l.line,
				Column:   l.column,
			}
		}

		if unicode.IsDigit(*l.currentChar) {
			return Token{
				Type:     NUMBER,
				Value:    l.readNumber(),
				Position: l.pos,
				Line:     l.line,
				Column:   l.column,
			}
		}

		if unicode.IsLetter(*l.currentChar) || *l.currentChar == '_' {
			value := l.readIdentifier()
			tokenType := IDENTIFIER
			if kw, exists := l.keywords[strings.ToUpper(value)]; exists {
				tokenType = kw
			}
			return Token{
				Type:     tokenType,
				Value:    value,
				Position: l.pos,
				Line:     l.line,
				Column:   l.column,
			}
		}

		// Single character tokens
		charTokens := map[rune]TokenType{
			'(': LPAREN,
			')': RPAREN,
			',': COMMA,
			';': SEMICOLON,
			'=': EQUALS,
			'.': DOT,
		}

		if tokenType, exists := charTokens[*l.currentChar]; exists {
			token := Token{
				Type:     tokenType,
				Value:    string(*l.currentChar),
				Position: l.pos,
				Line:     l.line,
				Column:   l.column,
			}
			l.advance()
			return token
		}

		// Skip unknown characters
		l.advance()
	}

	return Token{
		Type:     EOF,
		Value:    "",
		Position: l.pos,
		Line:     l.line,
		Column:   l.column,
	}
}

// Tokenize returns all tokens from the input
func (l *MySQLLexer) Tokenize() []Token {
	var tokens []Token
	for {
		token := l.GetNextToken()
		tokens = append(tokens, token)
		if token.Type == EOF {
			break
		}
	}
	return tokens
}
