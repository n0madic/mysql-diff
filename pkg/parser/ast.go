package parser

// ASTNode is the base interface for all AST nodes
type ASTNode interface {
	*CreateTableStatement | *ColumnDefinition | *IndexDefinition | *PrimaryKeyDefinition | *ForeignKeyDefinition | *CheckConstraint | *TableOptions | *PartitionOptions | *PartitionDefinition
}

// DataType represents a MySQL data type
type DataType struct {
	Name       string
	Parameters []string
	Unsigned   bool
	Zerofill   bool
}

// GeneratedColumn represents a generated column definition
type GeneratedColumn struct {
	Expression string `json:"expression"` // SQL expression for generation
	Type       string `json:"type"`       // VIRTUAL or STORED
}

// ColumnDefinition represents a column definition in a CREATE TABLE statement
type ColumnDefinition struct {
	Name          string
	DataType      DataType
	Nullable      *bool // nil = not specified, true = NULL, false = NOT NULL
	DefaultValue  *string
	AutoIncrement bool
	Unique        bool
	PrimaryKey    bool
	Comment       *string
	Collation     *string
	CharacterSet  *string
	Visible       *bool
	Generated     *GeneratedColumn // generated column definition
	ColumnFormat  *string
	Storage       *string
	Reference     *ForeignKeyReference
}

// IndexColumn represents a column reference in an index
type IndexColumn struct {
	Name      string
	Length    *int
	Direction *string // ASC, DESC
}

// IndexDefinition represents an index definition
type IndexDefinition struct {
	Name            *string
	IndexType       string // INDEX, UNIQUE, FULLTEXT, SPATIAL
	Columns         []IndexColumn
	KeyBlockSize    *int
	Using           *string // BTREE, HASH
	Comment         *string
	Visible         *bool
	Parser          *string // For FULLTEXT WITH PARSER
	Algorithm       *string // INPLACE, etc.
	Lock            *string // NONE, etc.
	EngineAttribute *string
}

// PrimaryKeyDefinition represents a primary key definition
type PrimaryKeyDefinition struct {
	Columns []IndexColumn
	Name    *string
	Using   *string
	Comment *string
}

// ForeignKeyReference represents a foreign key reference
type ForeignKeyReference struct {
	TableName string
	Columns   []string
	OnDelete  *string
	OnUpdate  *string
}

// ForeignKeyDefinition represents a foreign key constraint
type ForeignKeyDefinition struct {
	Name      *string
	Columns   []string
	Reference ForeignKeyReference
}

// CheckConstraint represents a check constraint
type CheckConstraint struct {
	Name       *string
	Expression string
	Enforced   *bool
}

// TableOptions represents table-level options
type TableOptions struct {
	Engine           *string
	AutoIncrement    *int
	CharacterSet     *string
	Collate          *string
	Comment          *string
	RowFormat        *string
	KeyBlockSize     *int
	MaxRows          *int
	MinRows          *int
	Tablespace       *string
	DataDirectory    *string
	IndexDirectory   *string
	Encryption       *string
	Compression      *string
	StatsPersistent  *int
	StatsAutoRecalc  *int
	StatsSamplePages *int
	PackKeys         *int
	Checksum         *int
	DelayKeyWrite    *int
	Union            []string
	InsertMethod     *string
}

// PartitionDefinition represents a single partition
type PartitionDefinition struct {
	Name           string
	Type           string // RANGE, LIST, HASH, KEY
	Expression     *string
	Values         []string // For RANGE/LIST partitions
	Comment        *string
	DataDirectory  *string
	IndexDirectory *string
	MaxRows        *int
	MinRows        *int
	Tablespace     *string
}

// PartitionOptions represents partitioning options
type PartitionOptions struct {
	Type           string // RANGE, LIST, HASH, KEY
	Expression     *string
	Columns        []string // For RANGE/LIST COLUMNS
	Linear         bool
	Partitions     []PartitionDefinition
	PartitionCount *int // For HASH/KEY without explicit partition definitions
}

// CreateTableStatement represents a complete CREATE TABLE statement
type CreateTableStatement struct {
	TableName        string
	Temporary        bool
	IfNotExists      bool
	Columns          []ColumnDefinition
	Indexes          []IndexDefinition
	PrimaryKey       *PrimaryKeyDefinition
	ForeignKeys      []ForeignKeyDefinition
	CheckConstraints []CheckConstraint
	TableOptions     *TableOptions
	PartitionOptions *PartitionOptions
}
