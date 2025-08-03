package parser

// TokenType represents the type of a SQL token
type TokenType int

const (
	// Keywords
	CREATE TokenType = iota
	TABLE
	TEMPORARY
	IF
	NOT
	EXISTS
	LIKE
	AS
	SELECT
	IGNORE
	REPLACE

	// Column types
	INT
	TINYINT
	SMALLINT
	MEDIUMINT
	BIGINT
	VARCHAR
	CHAR
	TEXT
	DECIMAL
	FLOAT
	DOUBLE
	DATE
	DATETIME
	TIMESTAMP
	TIME
	YEAR
	BLOB
	JSON
	ENUM
	SET
	BINARY
	VARBINARY
	BIT
	// Geometric types
	GEOMETRY
	POINT
	LINESTRING
	POLYGON

	// Column attributes
	NULL
	DEFAULT
	AUTO_INCREMENT
	UNIQUE
	PRIMARY
	KEY
	COMMENT
	COLLATE
	CHARACTER
	CHARSET
	VISIBLE
	INVISIBLE
	GENERATED
	ALWAYS
	VIRTUAL
	STORED
	UNSIGNED
	ZEROFILL

	// Index types
	INDEX
	FULLTEXT
	SPATIAL
	FOREIGN
	REFERENCES
	CHECK
	CONSTRAINT

	// Table options
	ENGINE
	AUTO_INCREMENT_OPT
	COMMENT_OPT
	ROW_FORMAT
	TABLESPACE
	DATA
	DIRECTORY
	COMPRESSION
	ENCRYPTION
	KEY_BLOCK_SIZE
	MAX_ROWS
	MIN_ROWS
	STATS_PERSISTENT
	STATS_AUTO_RECALC
	STATS_SAMPLE_PAGES
	PACK_KEYS
	CHECKSUM
	DELAY_KEY_WRITE
	UNION
	INSERT_METHOD

	// Partition options
	PARTITION
	BY
	HASH
	RANGE
	LIST
	COLUMNS
	VALUES
	LESS
	THAN
	IN
	MAXVALUE
	LINEAR

	// Reference options
	ON
	DELETE
	UPDATE
	CASCADE
	RESTRICT
	SET_NULL
	NO
	ACTION

	// Index options
	ASC
	DESC
	WITH
	PARSER
	ALGORITHM
	LOCK
	ENGINE_ATTRIBUTE
	INPLACE
	NONE
	FIRST
	LAST

	// Column format and storage options
	COLUMN_FORMAT
	FIXED
	DYNAMIC
	STORAGE
	DISK
	MEMORY
	COMPRESSED

	// Symbols
	LPAREN
	RPAREN
	COMMA
	SEMICOLON
	EQUALS
	DOT

	// Literals
	IDENTIFIER
	STRING
	NUMBER

	// Special
	EOF
	WHITESPACE
	SQL_COMMENT
	MYSQL_DIRECTIVE

	// Additional keywords for dumps
	DROP
	USE
	DATABASE
)

// String returns the string representation of a TokenType
func (t TokenType) String() string {
	tokens := map[TokenType]string{
		CREATE:             "CREATE",
		TABLE:              "TABLE",
		TEMPORARY:          "TEMPORARY",
		IF:                 "IF",
		NOT:                "NOT",
		EXISTS:             "EXISTS",
		LIKE:               "LIKE",
		AS:                 "AS",
		SELECT:             "SELECT",
		IGNORE:             "IGNORE",
		REPLACE:            "REPLACE",
		INT:                "INT",
		TINYINT:            "TINYINT",
		SMALLINT:           "SMALLINT",
		MEDIUMINT:          "MEDIUMINT",
		BIGINT:             "BIGINT",
		VARCHAR:            "VARCHAR",
		CHAR:               "CHAR",
		TEXT:               "TEXT",
		DECIMAL:            "DECIMAL",
		FLOAT:              "FLOAT",
		DOUBLE:             "DOUBLE",
		DATE:               "DATE",
		DATETIME:           "DATETIME",
		TIMESTAMP:          "TIMESTAMP",
		TIME:               "TIME",
		YEAR:               "YEAR",
		BLOB:               "BLOB",
		JSON:               "JSON",
		ENUM:               "ENUM",
		SET:                "SET",
		BINARY:             "BINARY",
		VARBINARY:          "VARBINARY",
		BIT:                "BIT",
		GEOMETRY:           "GEOMETRY",
		POINT:              "POINT",
		LINESTRING:         "LINESTRING",
		POLYGON:            "POLYGON",
		NULL:               "NULL",
		DEFAULT:            "DEFAULT",
		AUTO_INCREMENT:     "AUTO_INCREMENT",
		UNIQUE:             "UNIQUE",
		PRIMARY:            "PRIMARY",
		KEY:                "KEY",
		COMMENT:            "COMMENT",
		COLLATE:            "COLLATE",
		CHARACTER:          "CHARACTER",
		CHARSET:            "CHARSET",
		VISIBLE:            "VISIBLE",
		INVISIBLE:          "INVISIBLE",
		GENERATED:          "GENERATED",
		ALWAYS:             "ALWAYS",
		VIRTUAL:            "VIRTUAL",
		STORED:             "STORED",
		UNSIGNED:           "UNSIGNED",
		ZEROFILL:           "ZEROFILL",
		INDEX:              "INDEX",
		FULLTEXT:           "FULLTEXT",
		SPATIAL:            "SPATIAL",
		FOREIGN:            "FOREIGN",
		REFERENCES:         "REFERENCES",
		CHECK:              "CHECK",
		CONSTRAINT:         "CONSTRAINT",
		ENGINE:             "ENGINE",
		AUTO_INCREMENT_OPT: "AUTO_INCREMENT",
		COMMENT_OPT:        "COMMENT",
		ROW_FORMAT:         "ROW_FORMAT",
		TABLESPACE:         "TABLESPACE",
		DATA:               "DATA",
		DIRECTORY:          "DIRECTORY",
		COMPRESSION:        "COMPRESSION",
		ENCRYPTION:         "ENCRYPTION",
		KEY_BLOCK_SIZE:     "KEY_BLOCK_SIZE",
		MAX_ROWS:           "MAX_ROWS",
		MIN_ROWS:           "MIN_ROWS",
		STATS_PERSISTENT:   "STATS_PERSISTENT",
		STATS_AUTO_RECALC:  "STATS_AUTO_RECALC",
		STATS_SAMPLE_PAGES: "STATS_SAMPLE_PAGES",
		PACK_KEYS:          "PACK_KEYS",
		CHECKSUM:           "CHECKSUM",
		DELAY_KEY_WRITE:    "DELAY_KEY_WRITE",
		UNION:              "UNION",
		INSERT_METHOD:      "INSERT_METHOD",
		PARTITION:          "PARTITION",
		BY:                 "BY",
		HASH:               "HASH",
		RANGE:              "RANGE",
		LIST:               "LIST",
		COLUMNS:            "COLUMNS",
		VALUES:             "VALUES",
		LESS:               "LESS",
		THAN:               "THAN",
		IN:                 "IN",
		MAXVALUE:           "MAXVALUE",
		LINEAR:             "LINEAR",
		ON:                 "ON",
		DELETE:             "DELETE",
		UPDATE:             "UPDATE",
		CASCADE:            "CASCADE",
		RESTRICT:           "RESTRICT",
		SET_NULL:           "SET_NULL",
		NO:                 "NO",
		ACTION:             "ACTION",
		ASC:                "ASC",
		DESC:               "DESC",
		WITH:               "WITH",
		PARSER:             "PARSER",
		ALGORITHM:          "ALGORITHM",
		LOCK:               "LOCK",
		ENGINE_ATTRIBUTE:   "ENGINE_ATTRIBUTE",
		INPLACE:            "INPLACE",
		NONE:               "NONE",
		FIRST:              "FIRST",
		LAST:               "LAST",
		COLUMN_FORMAT:      "COLUMN_FORMAT",
		FIXED:              "FIXED",
		DYNAMIC:            "DYNAMIC",
		STORAGE:            "STORAGE",
		DISK:               "DISK",
		MEMORY:             "MEMORY",
		COMPRESSED:         "COMPRESSED",
		LPAREN:             "(",
		RPAREN:             ")",
		COMMA:              ",",
		SEMICOLON:          ";",
		EQUALS:             "=",
		DOT:                ".",
		IDENTIFIER:         "IDENTIFIER",
		STRING:             "STRING",
		NUMBER:             "NUMBER",
		EOF:                "EOF",
		WHITESPACE:         "WHITESPACE",
		SQL_COMMENT:        "SQL_COMMENT",
		MYSQL_DIRECTIVE:    "MYSQL_DIRECTIVE",
		DROP:               "DROP",
		USE:                "USE",
		DATABASE:           "DATABASE",
	}
	if name, ok := tokens[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Token represents a single token in the SQL
type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}
