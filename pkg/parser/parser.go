package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// MySQLCreateTableParser parses CREATE TABLE statements
type MySQLCreateTableParser struct {
	tokens       []Token
	pos          int
	currentToken Token
}

// NewMySQLCreateTableParser creates a new parser instance
func NewMySQLCreateTableParser(tokens []Token) *MySQLCreateTableParser {
	parser := &MySQLCreateTableParser{
		tokens: tokens,
		pos:    0,
	}

	if len(tokens) > 0 {
		parser.currentToken = tokens[0]
	} else {
		parser.currentToken = Token{Type: EOF, Value: "", Position: 0}
	}

	return parser
}

// advance moves to the next token
func (p *MySQLCreateTableParser) advance() {
	p.pos++
	if p.pos < len(p.tokens) {
		p.currentToken = p.tokens[p.pos]
	} else {
		p.currentToken = Token{Type: EOF, Value: "", Position: p.pos}
	}
}

// peek looks ahead at the next token(s)
func (p *MySQLCreateTableParser) peek(offset ...int) Token {
	off := 1
	if len(offset) > 0 {
		off = offset[0]
	}

	peekPos := p.pos + off
	if peekPos < len(p.tokens) {
		return p.tokens[peekPos]
	}
	return Token{Type: EOF, Value: "", Position: peekPos}
}

// match checks if current token matches any of the given types
func (p *MySQLCreateTableParser) match(tokenTypes ...TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.currentToken.Type == tokenType {
			return true
		}
	}
	return false
}

// consume expects and consumes a specific token type
func (p *MySQLCreateTableParser) consume(tokenType TokenType) (Token, error) {
	if p.currentToken.Type == tokenType {
		token := p.currentToken
		p.advance()
		return token, nil
	}
	return Token{}, fmt.Errorf("expected %s, got %s at line %d, column %d",
		tokenType.String(), p.currentToken.Type.String(), p.currentToken.Line, p.currentToken.Column)
}

// Parse parses the tokens into a CREATE TABLE statement
func (p *MySQLCreateTableParser) Parse() (*CreateTableStatement, error) {
	return p.parseCreateTable()
}

// parseCreateTable parses a CREATE TABLE statement
func (p *MySQLCreateTableParser) parseCreateTable() (*CreateTableStatement, error) {
	// CREATE [TEMPORARY] TABLE [IF NOT EXISTS] table_name
	if _, err := p.consume(CREATE); err != nil {
		return nil, err
	}

	temporary := false
	if p.match(TEMPORARY) {
		temporary = true
		p.advance()
	}

	if _, err := p.consume(TABLE); err != nil {
		return nil, err
	}

	ifNotExists := false
	if p.match(IF) {
		p.advance()
		if _, err := p.consume(NOT); err != nil {
			return nil, err
		}
		if _, err := p.consume(EXISTS); err != nil {
			return nil, err
		}
		ifNotExists = true
	}

	tableNameToken, err := p.consume(IDENTIFIER)
	if err != nil {
		return nil, err
	}

	stmt := &CreateTableStatement{
		TableName:   tableNameToken.Value,
		Temporary:   temporary,
		IfNotExists: ifNotExists,
	}

	// Parse column definitions and constraints
	if p.match(LPAREN) {
		p.advance()
		if err := p.parseTableElements(stmt); err != nil {
			return nil, err
		}
		if _, err := p.consume(RPAREN); err != nil {
			return nil, err
		}
	}

	// Parse table options
	if !p.match(EOF, SEMICOLON) {
		tableOptions, err := p.parseTableOptions()
		if err != nil {
			return nil, err
		}
		stmt.TableOptions = tableOptions
	}

	// Parse partition options
	if p.match(PARTITION) {
		partitionOptions, err := p.parsePartitionOptions()
		if err != nil {
			return nil, err
		}
		stmt.PartitionOptions = partitionOptions
	}

	return stmt, nil
}

// parseTableElements parses the elements inside the CREATE TABLE parentheses
func (p *MySQLCreateTableParser) parseTableElements(stmt *CreateTableStatement) error {
	for !p.match(RPAREN) {
		if p.match(CONSTRAINT) {
			// Handle named constraints
			p.advance() // CONSTRAINT
			// For now, skip to the actual constraint type
			if p.match(IDENTIFIER) {
				p.advance() // constraint name
			}
		}

		if p.match(PRIMARY) {
			primaryKey, err := p.parsePrimaryKey()
			if err != nil {
				return err
			}
			stmt.PrimaryKey = primaryKey
		} else if p.match(UNIQUE) {
			index, err := p.parseUniqueIndex()
			if err != nil {
				return err
			}
			stmt.Indexes = append(stmt.Indexes, index)
		} else if p.match(INDEX, KEY) {
			index, err := p.parseIndex()
			if err != nil {
				return err
			}
			stmt.Indexes = append(stmt.Indexes, index)
		} else if p.match(FULLTEXT) {
			index, err := p.parseFulltextIndex()
			if err != nil {
				return err
			}
			stmt.Indexes = append(stmt.Indexes, index)
		} else if p.match(SPATIAL) {
			index, err := p.parseSpatialIndex()
			if err != nil {
				return err
			}
			stmt.Indexes = append(stmt.Indexes, index)
		} else if p.match(FOREIGN) {
			foreignKey, err := p.parseForeignKey()
			if err != nil {
				return err
			}
			stmt.ForeignKeys = append(stmt.ForeignKeys, foreignKey)
		} else if p.match(CHECK) {
			checkConstraint, err := p.parseCheckConstraint()
			if err != nil {
				return err
			}
			stmt.CheckConstraints = append(stmt.CheckConstraints, checkConstraint)
		} else {
			// Column definition
			column, err := p.parseColumnDefinition()
			if err != nil {
				return err
			}
			stmt.Columns = append(stmt.Columns, column)
		}

		if p.match(COMMA) {
			p.advance()
		} else if !p.match(RPAREN) {
			break
		}
	}

	return nil
}

// parseColumnDefinition parses a column definition
func (p *MySQLCreateTableParser) parseColumnDefinition() (ColumnDefinition, error) {
	// Accept both IDENTIFIER and certain keywords as column names
	var nameToken Token
	if p.match(IDENTIFIER) {
		nameToken = p.currentToken
		p.advance()
	} else if p.isKeywordUsableAsIdentifier() {
		nameToken = p.currentToken
		p.advance()
	} else {
		return ColumnDefinition{}, fmt.Errorf("expected column name, got %s at line %d, column %d",
			p.currentToken.Type.String(), p.currentToken.Line, p.currentToken.Column)
	}

	dataType, err := p.parseDataType()
	if err != nil {
		return ColumnDefinition{}, err
	}

	column := ColumnDefinition{
		Name:     nameToken.Value,
		DataType: dataType,
	}

	// Parse column attributes
	for !p.match(COMMA, RPAREN, EOF) {
		if p.match(NOT) {
			p.advance()
			if p.match(NULL) {
				p.advance()
				nullable := false
				column.Nullable = &nullable
			}
		} else if p.match(NULL) {
			p.advance()
			nullable := true
			column.Nullable = &nullable
		} else if p.match(DEFAULT) {
			p.advance()
			// Parse default value expression (can be multiple tokens)
			defaultValue := ""
			if p.match(STRING, NUMBER, NULL, TRUE, FALSE, IDENTIFIER) {
				defaultValue = p.currentToken.Value
				p.advance()
			}
			column.DefaultValue = &defaultValue
		} else if p.match(AUTO_INCREMENT) {
			p.advance()
			column.AutoIncrement = true
		} else if p.match(UNIQUE) {
			p.advance()
			column.Unique = true
		} else if p.match(PRIMARY) {
			p.advance()
			if p.match(KEY) {
				p.advance()
			}
			column.PrimaryKey = true
		} else if p.match(COMMENT) {
			p.advance()
			if p.match(STRING) {
				comment := p.currentToken.Value
				column.Comment = &comment
				p.advance()
			}
		} else if p.match(COLLATE) {
			p.advance()
			if p.match(IDENTIFIER) {
				collation := p.currentToken.Value
				column.Collation = &collation
				p.advance()
			}
		} else if p.match(CHARACTER) {
			p.advance()
			if p.match(SET) {
				p.advance()
				if p.match(IDENTIFIER) {
					charSet := p.currentToken.Value
					column.CharacterSet = &charSet
					p.advance()
				}
			}
		} else if p.match(GENERATED) {
			p.advance()
			generated := &GeneratedColumn{
				Type: "VIRTUAL", // default
			}

			if p.match(ALWAYS) {
				p.advance()
			}

			if p.match(AS) {
				p.advance()
				if p.match(LPAREN) {
					p.advance()
					// Parse expression (simplified)
					expr := ""
					parenCount := 1
					for parenCount > 0 && !p.match(EOF) {
						if p.match(LPAREN) {
							parenCount++
						} else if p.match(RPAREN) {
							parenCount--
						}
						if parenCount > 0 {
							expr += p.currentToken.Value + " "
						}
						p.advance()
					}
					generated.Expression = strings.TrimSpace(expr)
				}
			}

			if p.match(VIRTUAL) {
				p.advance()
				generated.Type = "VIRTUAL"
			} else if p.match(STORED) {
				p.advance()
				generated.Type = "STORED"
			}

			column.Generated = generated
		} else if p.match(VISIBLE) {
			p.advance()
			visible := true
			column.Visible = &visible
		} else if p.match(INVISIBLE) {
			p.advance()
			visible := false
			column.Visible = &visible
		} else if p.match(ON) {
			p.advance()
			if p.match(UPDATE) {
				p.advance()
				// Skip the ON UPDATE expression for now
				// This is typically CURRENT_TIMESTAMP
				if p.match(IDENTIFIER) {
					p.advance()
				}
			}
		} else {
			// Skip unknown attributes
			p.advance()
		}
	}

	return column, nil
}

// parseDataType parses a data type definition
func (p *MySQLCreateTableParser) parseDataType() (DataType, error) {
	dataType := DataType{}

	// Data type name
	if !p.match(INT, TINYINT, SMALLINT, MEDIUMINT, BIGINT, VARCHAR, CHAR, TEXT,
		DECIMAL, FLOAT, DOUBLE, DATE, DATETIME, TIMESTAMP, TIME, YEAR, BLOB,
		JSON, ENUM, SET, BINARY, VARBINARY, BIT, BOOLEAN, GEOMETRY, POINT, LINESTRING, POLYGON) {
		return dataType, fmt.Errorf("expected data type, got %s", p.currentToken.Type.String())
	}

	dataType.Name = p.currentToken.Value
	p.advance()

	// Parse parameters if present
	if p.match(LPAREN) {
		p.advance()

		for !p.match(RPAREN) {
			if p.match(NUMBER, STRING, IDENTIFIER) {
				dataType.Parameters = append(dataType.Parameters, p.currentToken.Value)
				p.advance()
			}

			if p.match(COMMA) {
				p.advance()
			} else {
				break
			}
		}

		if _, err := p.consume(RPAREN); err != nil {
			return dataType, err
		}
	}

	// Parse UNSIGNED and ZEROFILL
	if p.match(UNSIGNED) {
		p.advance()
		dataType.Unsigned = true
	}

	if p.match(ZEROFILL) {
		p.advance()
		dataType.Zerofill = true
	}

	return dataType, nil
}

// parsePrimaryKey parses a primary key definition
func (p *MySQLCreateTableParser) parsePrimaryKey() (*PrimaryKeyDefinition, error) {
	if _, err := p.consume(PRIMARY); err != nil {
		return nil, err
	}
	if _, err := p.consume(KEY); err != nil {
		return nil, err
	}

	pk := &PrimaryKeyDefinition{}

	if _, err := p.consume(LPAREN); err != nil {
		return nil, err
	}

	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return nil, err
		}

		indexCol := IndexColumn{
			Name: columnToken.Value,
		}

		// Parse optional length
		if p.match(LPAREN) {
			p.advance()
			if p.match(NUMBER) {
				if length, err := strconv.Atoi(p.currentToken.Value); err == nil {
					indexCol.Length = &length
				}
				p.advance()
			}
			if _, err := p.consume(RPAREN); err != nil {
				return nil, err
			}
		}

		// Parse optional direction
		if p.match(ASC) {
			direction := "ASC"
			indexCol.Direction = &direction
			p.advance()
		} else if p.match(DESC) {
			direction := "DESC"
			indexCol.Direction = &direction
			p.advance()
		}

		pk.Columns = append(pk.Columns, indexCol)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return nil, err
	}

	return pk, nil
}

// parseIndex parses a regular index definition
func (p *MySQLCreateTableParser) parseIndex() (IndexDefinition, error) {
	index := IndexDefinition{
		IndexType: "INDEX",
	}

	if _, err := p.consume(INDEX); err != nil {
		if _, err := p.consume(KEY); err != nil {
			return index, err
		}
	}

	// Optional index name
	if p.match(IDENTIFIER) {
		name := p.currentToken.Value
		index.Name = &name
		p.advance()
	}

	if _, err := p.consume(LPAREN); err != nil {
		return index, err
	}

	// Parse index columns
	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return index, err
		}

		indexCol := IndexColumn{
			Name: columnToken.Value,
		}

		// Parse optional length and direction (similar to primary key)
		if p.match(LPAREN) {
			p.advance()
			if p.match(NUMBER) {
				if length, err := strconv.Atoi(p.currentToken.Value); err == nil {
					indexCol.Length = &length
				}
				p.advance()
			}
			if _, err := p.consume(RPAREN); err != nil {
				return index, err
			}
		}

		if p.match(ASC) {
			direction := "ASC"
			indexCol.Direction = &direction
			p.advance()
		} else if p.match(DESC) {
			direction := "DESC"
			indexCol.Direction = &direction
			p.advance()
		}

		index.Columns = append(index.Columns, indexCol)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return index, err
	}

	return index, nil
}

// parseUniqueIndex parses a unique index definition
func (p *MySQLCreateTableParser) parseUniqueIndex() (IndexDefinition, error) {
	if _, err := p.consume(UNIQUE); err != nil {
		return IndexDefinition{}, err
	}

	index := IndexDefinition{
		IndexType: "UNIQUE",
	}

	// Optional KEY or INDEX keyword
	if p.match(KEY, INDEX) {
		p.advance()
	}

	// Optional index name
	if p.match(IDENTIFIER) {
		name := p.currentToken.Value
		index.Name = &name
		p.advance()
	}

	// Parse columns (similar to regular index)
	if _, err := p.consume(LPAREN); err != nil {
		return index, err
	}

	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return index, err
		}

		indexCol := IndexColumn{
			Name: columnToken.Value,
		}

		if p.match(LPAREN) {
			p.advance()
			if p.match(NUMBER) {
				if length, err := strconv.Atoi(p.currentToken.Value); err == nil {
					indexCol.Length = &length
				}
				p.advance()
			}
			if _, err := p.consume(RPAREN); err != nil {
				return index, err
			}
		}

		if p.match(ASC) {
			direction := "ASC"
			indexCol.Direction = &direction
			p.advance()
		} else if p.match(DESC) {
			direction := "DESC"
			indexCol.Direction = &direction
			p.advance()
		}

		index.Columns = append(index.Columns, indexCol)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return index, err
	}

	return index, nil
}

// parseFulltextIndex parses a fulltext index definition
func (p *MySQLCreateTableParser) parseFulltextIndex() (IndexDefinition, error) {
	if _, err := p.consume(FULLTEXT); err != nil {
		return IndexDefinition{}, err
	}

	index := IndexDefinition{
		IndexType: "FULLTEXT",
	}

	// Optional KEY or INDEX keyword
	if p.match(KEY, INDEX) {
		p.advance()
	}

	// Optional index name
	if p.match(IDENTIFIER) {
		name := p.currentToken.Value
		index.Name = &name
		p.advance()
	}

	// Parse columns
	if _, err := p.consume(LPAREN); err != nil {
		return index, err
	}

	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return index, err
		}

		indexCol := IndexColumn{
			Name: columnToken.Value,
		}

		index.Columns = append(index.Columns, indexCol)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return index, err
	}

	return index, nil
}

// parseSpatialIndex parses a spatial index definition
func (p *MySQLCreateTableParser) parseSpatialIndex() (IndexDefinition, error) {
	if _, err := p.consume(SPATIAL); err != nil {
		return IndexDefinition{}, err
	}

	index := IndexDefinition{
		IndexType: "SPATIAL",
	}

	// Optional KEY or INDEX keyword
	if p.match(KEY, INDEX) {
		p.advance()
	}

	// Optional index name
	if p.match(IDENTIFIER) {
		name := p.currentToken.Value
		index.Name = &name
		p.advance()
	}

	// Parse columns
	if _, err := p.consume(LPAREN); err != nil {
		return index, err
	}

	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return index, err
		}

		indexCol := IndexColumn{
			Name: columnToken.Value,
		}

		index.Columns = append(index.Columns, indexCol)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return index, err
	}

	return index, nil
}

// parseForeignKey parses a foreign key definition
func (p *MySQLCreateTableParser) parseForeignKey() (ForeignKeyDefinition, error) {
	if _, err := p.consume(FOREIGN); err != nil {
		return ForeignKeyDefinition{}, err
	}
	if _, err := p.consume(KEY); err != nil {
		return ForeignKeyDefinition{}, err
	}

	fk := ForeignKeyDefinition{}

	// Optional constraint name
	if p.match(IDENTIFIER) && p.peek().Type == LPAREN {
		name := p.currentToken.Value
		fk.Name = &name
		p.advance()
	}

	// Parse local columns
	if _, err := p.consume(LPAREN); err != nil {
		return fk, err
	}

	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return fk, err
		}

		fk.Columns = append(fk.Columns, columnToken.Value)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return fk, err
	}

	// Parse REFERENCES
	if _, err := p.consume(REFERENCES); err != nil {
		return fk, err
	}

	tableToken, err := p.consume(IDENTIFIER)
	if err != nil {
		return fk, err
	}

	fk.Reference.TableName = tableToken.Value

	// Parse referenced columns
	if _, err := p.consume(LPAREN); err != nil {
		return fk, err
	}

	for !p.match(RPAREN) {
		columnToken, err := p.consume(IDENTIFIER)
		if err != nil {
			return fk, err
		}

		fk.Reference.Columns = append(fk.Reference.Columns, columnToken.Value)

		if p.match(COMMA) {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.consume(RPAREN); err != nil {
		return fk, err
	}

	// Parse ON DELETE and ON UPDATE clauses
	for p.match(ON) {
		p.advance()

		if p.match(DELETE) {
			p.advance()
			if p.match(CASCADE) {
				onDelete := "CASCADE"
				fk.Reference.OnDelete = &onDelete
				p.advance()
			} else if p.match(RESTRICT) {
				onDelete := "RESTRICT"
				fk.Reference.OnDelete = &onDelete
				p.advance()
			} else if p.match(SET_NULL) {
				onDelete := "SET NULL"
				fk.Reference.OnDelete = &onDelete
				p.advance()
			}
		} else if p.match(UPDATE) {
			p.advance()
			if p.match(CASCADE) {
				onUpdate := "CASCADE"
				fk.Reference.OnUpdate = &onUpdate
				p.advance()
			} else if p.match(RESTRICT) {
				onUpdate := "RESTRICT"
				fk.Reference.OnUpdate = &onUpdate
				p.advance()
			} else if p.match(SET_NULL) {
				onUpdate := "SET NULL"
				fk.Reference.OnUpdate = &onUpdate
				p.advance()
			}
		}
	}

	return fk, nil
}

// parseCheckConstraint parses a check constraint
func (p *MySQLCreateTableParser) parseCheckConstraint() (CheckConstraint, error) {
	if _, err := p.consume(CHECK); err != nil {
		return CheckConstraint{}, err
	}

	check := CheckConstraint{}

	// Parse check expression (simplified)
	if _, err := p.consume(LPAREN); err != nil {
		return check, err
	}

	expression := ""
	parenCount := 1
	for parenCount > 0 && !p.match(EOF) {
		if p.match(LPAREN) {
			parenCount++
		} else if p.match(RPAREN) {
			parenCount--
		}
		if parenCount > 0 {
			expression += p.currentToken.Value + " "
		}
		p.advance()
	}

	check.Expression = strings.TrimSpace(expression)

	return check, nil
}

// parseTableOptions parses table options
func (p *MySQLCreateTableParser) parseTableOptions() (*TableOptions, error) {
	options := &TableOptions{}

	for !p.match(EOF, SEMICOLON, PARTITION) {
		if p.match(ENGINE) {
			p.advance()
			if p.match(EQUALS) {
				p.advance()
			}
			if p.match(IDENTIFIER) {
				engine := p.currentToken.Value
				options.Engine = &engine
				p.advance()
			}
		} else if p.match(AUTO_INCREMENT) {
			p.advance()
			if p.match(EQUALS) {
				p.advance()
			}
			if p.match(NUMBER) {
				if autoInc, err := strconv.Atoi(p.currentToken.Value); err == nil {
					options.AutoIncrement = &autoInc
				}
				p.advance()
			}
		} else if p.match(DEFAULT) {
			p.advance()
			if p.match(CHARSET) {
				p.advance()
				if p.match(EQUALS) {
					p.advance()
				}
				if p.match(IDENTIFIER) {
					charset := p.currentToken.Value
					options.CharacterSet = &charset
					p.advance()
				}
			}
		} else if p.match(CHARACTER) {
			p.advance()
			if p.match(SET) {
				p.advance()
				if p.match(EQUALS) {
					p.advance()
				}
				if p.match(IDENTIFIER) {
					charset := p.currentToken.Value
					options.CharacterSet = &charset
					p.advance()
				}
			}
		} else if p.match(COLLATE) {
			p.advance()
			if p.match(EQUALS) {
				p.advance()
			}
			if p.match(IDENTIFIER) {
				collate := p.currentToken.Value
				options.Collate = &collate
				p.advance()
			}
		} else if p.match(COMMENT) {
			p.advance()
			if p.match(EQUALS) {
				p.advance()
			}
			if p.match(STRING) {
				comment := p.currentToken.Value
				options.Comment = &comment
				p.advance()
			}
		} else {
			// Skip unknown options
			p.advance()
		}
	}

	return options, nil
}

// parsePartitionOptions parses partition options (simplified)
func (p *MySQLCreateTableParser) parsePartitionOptions() (*PartitionOptions, error) {
	if _, err := p.consume(PARTITION); err != nil {
		return nil, err
	}
	if _, err := p.consume(BY); err != nil {
		return nil, err
	}

	partOptions := &PartitionOptions{}

	if p.match(HASH) {
		p.advance()
		partOptions.Type = "HASH"
		// Parse hash expression (simplified)
		if p.match(LPAREN) {
			p.advance()
			// Skip to closing paren
			for !p.match(RPAREN) && !p.match(EOF) {
				p.advance()
			}
			if _, err := p.consume(RPAREN); err != nil {
				return nil, err
			}
		}
	} else if p.match(RANGE) {
		p.advance()
		partOptions.Type = "RANGE"
		// Similar simplified parsing for other partition types
	}

	// Skip remaining partition details for now
	for !p.match(EOF, SEMICOLON) {
		p.advance()
	}

	return partOptions, nil
}

// isKeywordUsableAsIdentifier checks if the current token is a keyword that can be used as an identifier
func (p *MySQLCreateTableParser) isKeywordUsableAsIdentifier() bool {
	// List of keywords that can be used as column names in MySQL
	allowedKeywords := []TokenType{
		DATA, DIRECTORY, COMPRESSION, ENCRYPTION, TABLESPACE,
		STATS_PERSISTENT, STATS_AUTO_RECALC, STATS_SAMPLE_PAGES,
		PACK_KEYS, CHECKSUM, DELAY_KEY_WRITE, MEMORY, DISK,
		FIXED, DYNAMIC, COMPRESSED, FIRST, LAST, ACTION,
	}

	for _, keyword := range allowedKeywords {
		if p.match(keyword) {
			return true
		}
	}

	return false
}
