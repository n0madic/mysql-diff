package diff

import (
	"github.com/n0madic/mysql-diff/pkg/parser"
)

// ChangeType represents the type of change that occurred
type ChangeType string

const (
	ChangeTypeAdded     ChangeType = "added"
	ChangeTypeRemoved   ChangeType = "removed"
	ChangeTypeModified  ChangeType = "modified"
	ChangeTypeUnchanged ChangeType = "unchanged"
)

// FieldChange represents a change in a specific field
type FieldChange[T any] struct {
	Old T `json:"old"`
	New T `json:"new"`
}

// ColumnChanges represents specific field changes for columns
type ColumnChanges struct {
	DataType      *FieldChange[string]                  `json:"data_type,omitempty"`
	Nullable      *FieldChange[any]                     `json:"nullable,omitempty"`
	DefaultValue  *FieldChange[any]                     `json:"default_value,omitempty"`
	AutoIncrement *FieldChange[bool]                    `json:"auto_increment,omitempty"`
	Unique        *FieldChange[bool]                    `json:"unique,omitempty"`
	PrimaryKey    *FieldChange[bool]                    `json:"primary_key,omitempty"`
	Comment       *FieldChange[any]                     `json:"comment,omitempty"`
	Collation     *FieldChange[any]                     `json:"collation,omitempty"`
	CharacterSet  *FieldChange[any]                     `json:"character_set,omitempty"`
	Visible       *FieldChange[any]                     `json:"visible,omitempty"`
	ColumnFormat  *FieldChange[any]                     `json:"column_format,omitempty"`
	Storage       *FieldChange[any]                     `json:"storage,omitempty"`
	Generated     *FieldChange[*parser.GeneratedColumn] `json:"generated,omitempty"`
}

// HasChanges returns true if there are any changes in the column
func (c *ColumnChanges) HasChanges() bool {
	return c.DataType != nil || c.Nullable != nil || c.DefaultValue != nil ||
		c.AutoIncrement != nil || c.Unique != nil || c.PrimaryKey != nil ||
		c.Comment != nil || c.Collation != nil || c.CharacterSet != nil ||
		c.Visible != nil || c.ColumnFormat != nil || c.Storage != nil ||
		c.Generated != nil
}

// IndexChanges represents specific field changes for indexes
type IndexChanges struct {
	Name            *FieldChange[any]    `json:"name,omitempty"`
	IndexType       *FieldChange[string] `json:"index_type,omitempty"`
	Columns         *FieldChange[any]    `json:"columns,omitempty"`
	KeyBlockSize    *FieldChange[any]    `json:"key_block_size,omitempty"`
	Using           *FieldChange[any]    `json:"using,omitempty"`
	Comment         *FieldChange[any]    `json:"comment,omitempty"`
	Visible         *FieldChange[any]    `json:"visible,omitempty"`
	Parser          *FieldChange[any]    `json:"parser,omitempty"`
	Algorithm       *FieldChange[any]    `json:"algorithm,omitempty"`
	Lock            *FieldChange[any]    `json:"lock,omitempty"`
	EngineAttribute *FieldChange[any]    `json:"engine_attribute,omitempty"`
}

// HasChanges returns true if there are any changes in the index
func (c *IndexChanges) HasChanges() bool {
	return c.Name != nil || c.IndexType != nil || c.Columns != nil ||
		c.KeyBlockSize != nil || c.Using != nil || c.Comment != nil ||
		c.Visible != nil || c.Parser != nil || c.Algorithm != nil ||
		c.Lock != nil || c.EngineAttribute != nil
}

// PrimaryKeyChanges represents specific field changes for primary keys
type PrimaryKeyChanges struct {
	Columns *FieldChange[[]string] `json:"columns,omitempty"`
	Name    *FieldChange[any]      `json:"name,omitempty"`
	Using   *FieldChange[any]      `json:"using,omitempty"`
	Comment *FieldChange[any]      `json:"comment,omitempty"`
}

// HasChanges returns true if there are any changes in the primary key
func (c *PrimaryKeyChanges) HasChanges() bool {
	return c.Columns != nil || c.Name != nil || c.Using != nil || c.Comment != nil
}

// ForeignKeyChanges represents specific field changes for foreign keys
type ForeignKeyChanges struct {
	Name             *FieldChange[any]      `json:"name,omitempty"`
	Columns          *FieldChange[[]string] `json:"columns,omitempty"`
	ReferenceTable   *FieldChange[string]   `json:"reference_table,omitempty"`
	ReferenceColumns *FieldChange[[]string] `json:"reference_columns,omitempty"`
	OnDelete         *FieldChange[any]      `json:"on_delete,omitempty"`
	OnUpdate         *FieldChange[any]      `json:"on_update,omitempty"`
}

// HasChanges returns true if there are any changes in the foreign key
func (c *ForeignKeyChanges) HasChanges() bool {
	return c.Name != nil || c.Columns != nil || c.ReferenceTable != nil ||
		c.ReferenceColumns != nil || c.OnDelete != nil || c.OnUpdate != nil
}

// TableOptionsChanges represents specific field changes for table options
type TableOptionsChanges struct {
	Engine        *FieldChange[any] `json:"engine,omitempty"`
	AutoIncrement *FieldChange[any] `json:"auto_increment,omitempty"`
	CharacterSet  *FieldChange[any] `json:"character_set,omitempty"`
	Collate       *FieldChange[any] `json:"collate,omitempty"`
	Comment       *FieldChange[any] `json:"comment,omitempty"`
}

// HasChanges returns true if there are any changes in the table options
func (c *TableOptionsChanges) HasChanges() bool {
	return c.Engine != nil || c.AutoIncrement != nil || c.CharacterSet != nil ||
		c.Collate != nil || c.Comment != nil
}

// PartitionChanges represents specific field changes for partitions
type PartitionChanges struct {
	Type                 *FieldChange[string]   `json:"type,omitempty"`
	Linear               *FieldChange[bool]     `json:"linear,omitempty"`
	Expression           *FieldChange[any]      `json:"expression,omitempty"`
	Columns              *FieldChange[[]string] `json:"columns,omitempty"`
	PartitionsCount      *FieldChange[any]      `json:"partitions_count,omitempty"`
	PartitionDefinitions *FieldChange[any]      `json:"partition_definitions,omitempty"`
}

// HasChanges returns true if there are any changes in the partitions
func (c *PartitionChanges) HasChanges() bool {
	return c.Type != nil || c.Linear != nil || c.Expression != nil ||
		c.Columns != nil || c.PartitionsCount != nil || c.PartitionDefinitions != nil
}

// ColumnDiff represents differences in a column definition
type ColumnDiff struct {
	Name       string                   `json:"name"`
	ChangeType ChangeType               `json:"change_type"`
	OldColumn  *parser.ColumnDefinition `json:"old_column,omitempty"`
	NewColumn  *parser.ColumnDefinition `json:"new_column,omitempty"`
	Changes    *ColumnChanges           `json:"changes,omitempty"`
}

// IndexDiff represents differences in an index definition
type IndexDiff struct {
	Name       *string                 `json:"name"`
	ChangeType ChangeType              `json:"change_type"`
	OldIndex   *parser.IndexDefinition `json:"old_index,omitempty"`
	NewIndex   *parser.IndexDefinition `json:"new_index,omitempty"`
	Changes    *IndexChanges           `json:"changes,omitempty"`
}

// ForeignKeyDiff represents differences in a foreign key definition
type ForeignKeyDiff struct {
	Name       *string                      `json:"name"`
	ChangeType ChangeType                   `json:"change_type"`
	OldFK      *parser.ForeignKeyDefinition `json:"old_fk,omitempty"`
	NewFK      *parser.ForeignKeyDefinition `json:"new_fk,omitempty"`
	Changes    *ForeignKeyChanges           `json:"changes,omitempty"`
}

// PrimaryKeyDiff represents differences in primary key definition
type PrimaryKeyDiff struct {
	ChangeType ChangeType                   `json:"change_type"`
	OldPK      *parser.PrimaryKeyDefinition `json:"old_pk,omitempty"`
	NewPK      *parser.PrimaryKeyDefinition `json:"new_pk,omitempty"`
	Changes    *PrimaryKeyChanges           `json:"changes,omitempty"`
}

// TableOptionsDiff represents differences in table options
type TableOptionsDiff struct {
	ChangeType ChangeType           `json:"change_type"`
	OldOptions *parser.TableOptions `json:"old_options,omitempty"`
	NewOptions *parser.TableOptions `json:"new_options,omitempty"`
	Changes    *TableOptionsChanges `json:"changes,omitempty"`
}

// PartitionDiff represents differences in partition options
type PartitionDiff struct {
	ChangeType   ChangeType               `json:"change_type"`
	OldPartition *parser.PartitionOptions `json:"old_partition,omitempty"`
	NewPartition *parser.PartitionOptions `json:"new_partition,omitempty"`
	Changes      *PartitionChanges        `json:"changes,omitempty"`
}

// TableDiff represents complete difference analysis between two tables
type TableDiff struct {
	OldTable *parser.CreateTableStatement `json:"old_table"`
	NewTable *parser.CreateTableStatement `json:"new_table"`

	// Table-level changes
	TableNameChanged    bool `json:"table_name_changed"`
	TableOptionsChanged bool `json:"table_options_changed"`

	// Component differences
	ColumnDiffs      []ColumnDiff      `json:"column_diffs"`
	PrimaryKeyDiff   *PrimaryKeyDiff   `json:"primary_key_diff,omitempty"`
	IndexDiffs       []IndexDiff       `json:"index_diffs"`
	ForeignKeyDiffs  []ForeignKeyDiff  `json:"foreign_key_diffs"`
	TableOptionsDiff *TableOptionsDiff `json:"table_options_diff,omitempty"`
	PartitionDiff    *PartitionDiff    `json:"partition_diff,omitempty"`

	// Summary counters
	ColumnsAdded        int `json:"columns_added"`
	ColumnsRemoved      int `json:"columns_removed"`
	ColumnsModified     int `json:"columns_modified"`
	IndexesAdded        int `json:"indexes_added"`
	IndexesRemoved      int `json:"indexes_removed"`
	IndexesModified     int `json:"indexes_modified"`
	ForeignKeysAdded    int `json:"foreign_keys_added"`
	ForeignKeysRemoved  int `json:"foreign_keys_removed"`
	ForeignKeysModified int `json:"foreign_keys_modified"`
}

// HasChanges returns true if there are any changes between the tables
func (td *TableDiff) HasChanges() bool {
	return td.TableNameChanged ||
		td.TableOptionsChanged ||
		len(td.ColumnDiffs) > 0 ||
		td.PrimaryKeyDiff != nil ||
		len(td.IndexDiffs) > 0 ||
		len(td.ForeignKeyDiffs) > 0 ||
		td.TableOptionsDiff != nil ||
		td.PartitionDiff != nil
}

// TableSummary represents a typed summary of table changes
type TableSummary struct {
	TableNameChanged    bool           `json:"table_name_changed"`
	Columns             ChangesSummary `json:"columns"`
	Indexes             ChangesSummary `json:"indexes"`
	ForeignKeys         ChangesSummary `json:"foreign_keys"`
	PrimaryKeyChanged   bool           `json:"primary_key_changed"`
	TableOptionsChanged bool           `json:"table_options_changed"`
	PartitioningChanged bool           `json:"partitioning_changed"`
}

// ChangesSummary represents a summary of changes for a specific component
type ChangesSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Modified int `json:"modified"`
}

// GetSummary returns a typed summary of all changes
func (td *TableDiff) GetSummary() TableSummary {
	return TableSummary{
		TableNameChanged: td.TableNameChanged,
		Columns: ChangesSummary{
			Added:    td.ColumnsAdded,
			Removed:  td.ColumnsRemoved,
			Modified: td.ColumnsModified,
		},
		Indexes: ChangesSummary{
			Added:    td.IndexesAdded,
			Removed:  td.IndexesRemoved,
			Modified: td.IndexesModified,
		},
		ForeignKeys: ChangesSummary{
			Added:    td.ForeignKeysAdded,
			Removed:  td.ForeignKeysRemoved,
			Modified: td.ForeignKeysModified,
		},
		PrimaryKeyChanged:   td.PrimaryKeyDiff != nil,
		TableOptionsChanged: td.TableOptionsDiff != nil,
		PartitioningChanged: td.PartitionDiff != nil,
	}
}
