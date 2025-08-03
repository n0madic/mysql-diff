package diff

import (
	"fmt"
	"slices"
	"strings"

	"github.com/n0madic/mysql-diff/pkg/parser"
)

// comparePrimaryKeys compares primary key definitions
func (a *TableDiffAnalyzer) comparePrimaryKeys(oldPK, newPK *parser.PrimaryKeyDefinition) *PrimaryKeyDiff {
	if oldPK == nil && newPK == nil {
		return nil
	}

	if oldPK == nil {
		return &PrimaryKeyDiff{
			ChangeType: ChangeTypeAdded,
			NewPK:      newPK,
			Changes:    &PrimaryKeyChanges{},
		}
	}

	if newPK == nil {
		return &PrimaryKeyDiff{
			ChangeType: ChangeTypeRemoved,
			OldPK:      oldPK,
			Changes:    &PrimaryKeyChanges{},
		}
	}

	// Both exist, check for changes
	changes := &PrimaryKeyChanges{}

	// Compare column lists
	oldCols := make([]string, len(oldPK.Columns))
	newCols := make([]string, len(newPK.Columns))

	for i, col := range oldPK.Columns {
		oldCols[i] = col.Name
	}
	for i, col := range newPK.Columns {
		newCols[i] = col.Name
	}

	if !slices.Equal(oldCols, newCols) {
		changes.Columns = &FieldChange[[]string]{
			Old: oldCols,
			New: newCols,
		}
	}

	// Compare other attributes
	if !ptrEqual(oldPK.Name, newPK.Name) {
		changes.Name = &FieldChange[any]{
			Old: ptrToValue(oldPK.Name),
			New: ptrToValue(newPK.Name),
		}
	}

	if !ptrEqual(oldPK.Using, newPK.Using) {
		changes.Using = &FieldChange[any]{
			Old: ptrToValue(oldPK.Using),
			New: ptrToValue(newPK.Using),
		}
	}

	if !ptrEqual(oldPK.Comment, newPK.Comment) {
		changes.Comment = &FieldChange[any]{
			Old: ptrToValue(oldPK.Comment),
			New: ptrToValue(newPK.Comment),
		}
	}

	if changes.HasChanges() {
		return &PrimaryKeyDiff{
			ChangeType: ChangeTypeModified,
			OldPK:      oldPK,
			NewPK:      newPK,
			Changes:    changes,
		}
	}

	return nil
}

// compareIndexes compares index definitions
func (a *TableDiffAnalyzer) compareIndexes(oldIndexes, newIndexes []parser.IndexDefinition) []IndexDiff {
	var diffs []IndexDiff

	// Create maps for comparison (using name + columns as key since name can be nil)
	indexKey := func(idx parser.IndexDefinition) string {
		cols := make([]string, len(idx.Columns))
		for i, col := range idx.Columns {
			cols[i] = col.Name
		}
		name := ""
		if idx.Name != nil {
			name = *idx.Name
		}
		return fmt.Sprintf("%s:%s:%s", name, strings.Join(cols, ":"), idx.IndexType)
	}

	oldIndexesMap := make(map[string]parser.IndexDefinition)
	newIndexesMap := make(map[string]parser.IndexDefinition)

	for _, idx := range oldIndexes {
		oldIndexesMap[indexKey(idx)] = idx
	}
	for _, idx := range newIndexes {
		newIndexesMap[indexKey(idx)] = idx
	}

	// Find all index keys
	allIndexKeys := make(map[string]bool)
	for key := range oldIndexesMap {
		allIndexKeys[key] = true
	}
	for key := range newIndexesMap {
		allIndexKeys[key] = true
	}

	for idxKey := range allIndexKeys {
		oldIdx, hasOld := oldIndexesMap[idxKey]
		newIdx, hasNew := newIndexesMap[idxKey]

		if !hasOld {
			// Index added
			diffs = append(diffs, IndexDiff{
				Name:       newIdx.Name,
				ChangeType: ChangeTypeAdded,
				NewIndex:   &newIdx,
				Changes:    &IndexChanges{},
			})
		} else if !hasNew {
			// Index removed
			diffs = append(diffs, IndexDiff{
				Name:       oldIdx.Name,
				ChangeType: ChangeTypeRemoved,
				OldIndex:   &oldIdx,
				Changes:    &IndexChanges{},
			})
		} else {
			// Index exists in both, check for changes
			changes := a.compareIndexDefinitions(oldIdx, newIdx)
			if changes.HasChanges() {
				diffs = append(diffs, IndexDiff{
					Name:       oldIdx.Name,
					ChangeType: ChangeTypeModified,
					OldIndex:   &oldIdx,
					NewIndex:   &newIdx,
					Changes:    changes,
				})
			}
		}
	}

	return diffs
}

// compareIndexDefinitions compares two index definitions
func (a *TableDiffAnalyzer) compareIndexDefinitions(oldIdx, newIdx parser.IndexDefinition) *IndexChanges {
	changes := &IndexChanges{}

	// Compare basic attributes
	if !ptrEqual(oldIdx.Name, newIdx.Name) {
		changes.Name = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Name),
			New: ptrToValue(newIdx.Name),
		}
	}

	if oldIdx.IndexType != newIdx.IndexType {
		changes.IndexType = &FieldChange[string]{
			Old: oldIdx.IndexType,
			New: newIdx.IndexType,
		}
	}

	if !ptrEqual(oldIdx.KeyBlockSize, newIdx.KeyBlockSize) {
		changes.KeyBlockSize = &FieldChange[any]{
			Old: ptrToValue(oldIdx.KeyBlockSize),
			New: ptrToValue(newIdx.KeyBlockSize),
		}
	}

	if !ptrEqual(oldIdx.Comment, newIdx.Comment) {
		changes.Comment = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Comment),
			New: ptrToValue(newIdx.Comment),
		}
	}

	if !ptrEqual(oldIdx.Using, newIdx.Using) {
		changes.Using = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Using),
			New: ptrToValue(newIdx.Using),
		}
	}

	if !ptrEqual(oldIdx.Visible, newIdx.Visible) {
		changes.Visible = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Visible),
			New: ptrToValue(newIdx.Visible),
		}
	}

	if !ptrEqual(oldIdx.Parser, newIdx.Parser) {
		changes.Parser = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Parser),
			New: ptrToValue(newIdx.Parser),
		}
	}

	if !ptrEqual(oldIdx.Algorithm, newIdx.Algorithm) {
		changes.Algorithm = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Algorithm),
			New: ptrToValue(newIdx.Algorithm),
		}
	}

	if !ptrEqual(oldIdx.Lock, newIdx.Lock) {
		changes.Lock = &FieldChange[any]{
			Old: ptrToValue(oldIdx.Lock),
			New: ptrToValue(newIdx.Lock),
		}
	}

	if !ptrEqual(oldIdx.EngineAttribute, newIdx.EngineAttribute) {
		changes.EngineAttribute = &FieldChange[any]{
			Old: ptrToValue(oldIdx.EngineAttribute),
			New: ptrToValue(newIdx.EngineAttribute),
		}
	}

	// Compare columns
	if !a.indexColumnsEqual(oldIdx.Columns, newIdx.Columns) {
		oldCols := a.indexColumnsToString(oldIdx.Columns)
		newCols := a.indexColumnsToString(newIdx.Columns)
		changes.Columns = &FieldChange[any]{
			Old: oldCols,
			New: newCols,
		}
	}

	return changes
}

// indexColumnsEqual checks if two index column lists are equal
func (a *TableDiffAnalyzer) indexColumnsEqual(oldCols, newCols []parser.IndexColumn) bool {
	if len(oldCols) != len(newCols) {
		return false
	}

	for i, oldCol := range oldCols {
		newCol := newCols[i]
		if oldCol.Name != newCol.Name ||
			!ptrEqual(oldCol.Length, newCol.Length) ||
			!ptrEqual(oldCol.Direction, newCol.Direction) {
			return false
		}
	}

	return true
}

// indexColumnsToString converts index columns to string representation
func (a *TableDiffAnalyzer) indexColumnsToString(columns []parser.IndexColumn) string {
	var colStrs []string
	for _, col := range columns {
		colStr := col.Name
		if col.Length != nil {
			colStr += fmt.Sprintf("(%d)", *col.Length)
		}
		if col.Direction != nil {
			colStr += fmt.Sprintf(" %s", *col.Direction)
		}
		colStrs = append(colStrs, colStr)
	}
	return fmt.Sprintf("(%s)", strings.Join(colStrs, ", "))
}

// compareForeignKeys compares foreign key definitions
func (a *TableDiffAnalyzer) compareForeignKeys(oldFKs, newFKs []parser.ForeignKeyDefinition) []ForeignKeyDiff {
	var diffs []ForeignKeyDiff

	// Create maps for comparison
	fkKey := func(fk parser.ForeignKeyDefinition) string {
		cols := strings.Join(fk.Columns, ":")
		ref := fmt.Sprintf("%s:%s", fk.Reference.TableName, strings.Join(fk.Reference.Columns, ":"))
		name := ""
		if fk.Name != nil {
			name = *fk.Name
		}
		return fmt.Sprintf("%s:%s:%s", name, cols, ref)
	}

	oldFKsMap := make(map[string]parser.ForeignKeyDefinition)
	newFKsMap := make(map[string]parser.ForeignKeyDefinition)

	for _, fk := range oldFKs {
		oldFKsMap[fkKey(fk)] = fk
	}
	for _, fk := range newFKs {
		newFKsMap[fkKey(fk)] = fk
	}

	// Find all FK keys
	allFKKeys := make(map[string]bool)
	for key := range oldFKsMap {
		allFKKeys[key] = true
	}
	for key := range newFKsMap {
		allFKKeys[key] = true
	}

	for fkKeyStr := range allFKKeys {
		oldFK, hasOld := oldFKsMap[fkKeyStr]
		newFK, hasNew := newFKsMap[fkKeyStr]

		if !hasOld {
			// Foreign key added
			diffs = append(diffs, ForeignKeyDiff{
				Name:       newFK.Name,
				ChangeType: ChangeTypeAdded,
				NewFK:      &newFK,
				Changes:    &ForeignKeyChanges{},
			})
		} else if !hasNew {
			// Foreign key removed
			diffs = append(diffs, ForeignKeyDiff{
				Name:       oldFK.Name,
				ChangeType: ChangeTypeRemoved,
				OldFK:      &oldFK,
				Changes:    &ForeignKeyChanges{},
			})
		} else {
			// Foreign key exists in both, check for changes
			changes := a.compareForeignKeyDefinitions(oldFK, newFK)
			if changes.HasChanges() {
				diffs = append(diffs, ForeignKeyDiff{
					Name:       oldFK.Name,
					ChangeType: ChangeTypeModified,
					OldFK:      &oldFK,
					NewFK:      &newFK,
					Changes:    changes,
				})
			}
		}
	}

	return diffs
}

// compareForeignKeyDefinitions compares two foreign key definitions
func (a *TableDiffAnalyzer) compareForeignKeyDefinitions(oldFK, newFK parser.ForeignKeyDefinition) *ForeignKeyChanges {
	changes := &ForeignKeyChanges{}

	// Compare basic attributes
	if !ptrEqual(oldFK.Name, newFK.Name) {
		changes.Name = &FieldChange[any]{
			Old: ptrToValue(oldFK.Name),
			New: ptrToValue(newFK.Name),
		}
	}

	if !slices.Equal(oldFK.Columns, newFK.Columns) {
		changes.Columns = &FieldChange[[]string]{
			Old: oldFK.Columns,
			New: newFK.Columns,
		}
	}

	// Compare reference
	if oldFK.Reference.TableName != newFK.Reference.TableName {
		changes.ReferenceTable = &FieldChange[string]{
			Old: oldFK.Reference.TableName,
			New: newFK.Reference.TableName,
		}
	}

	if !slices.Equal(oldFK.Reference.Columns, newFK.Reference.Columns) {
		changes.ReferenceColumns = &FieldChange[[]string]{
			Old: oldFK.Reference.Columns,
			New: newFK.Reference.Columns,
		}
	}

	if !ptrEqual(oldFK.Reference.OnDelete, newFK.Reference.OnDelete) {
		changes.OnDelete = &FieldChange[any]{
			Old: ptrToValue(oldFK.Reference.OnDelete),
			New: ptrToValue(newFK.Reference.OnDelete),
		}
	}

	if !ptrEqual(oldFK.Reference.OnUpdate, newFK.Reference.OnUpdate) {
		changes.OnUpdate = &FieldChange[any]{
			Old: ptrToValue(oldFK.Reference.OnUpdate),
			New: ptrToValue(newFK.Reference.OnUpdate),
		}
	}

	return changes
}

// compareTableOptions compares table options
func (a *TableDiffAnalyzer) compareTableOptions(oldOpts, newOpts *parser.TableOptions) *TableOptionsDiff {
	if oldOpts == nil && newOpts == nil {
		return nil
	}

	if oldOpts == nil {
		return &TableOptionsDiff{
			ChangeType: ChangeTypeAdded,
			NewOptions: newOpts,
			Changes:    &TableOptionsChanges{},
		}
	}

	if newOpts == nil {
		return &TableOptionsDiff{
			ChangeType: ChangeTypeRemoved,
			OldOptions: oldOpts,
			Changes:    &TableOptionsChanges{},
		}
	}

	// Both exist, check for changes
	changes := &TableOptionsChanges{}

	// Compare all table option attributes
	if !ptrEqual(oldOpts.Engine, newOpts.Engine) {
		changes.Engine = &FieldChange[any]{
			Old: ptrToValue(oldOpts.Engine),
			New: ptrToValue(newOpts.Engine),
		}
	}

	if !ptrEqual(oldOpts.AutoIncrement, newOpts.AutoIncrement) {
		changes.AutoIncrement = &FieldChange[any]{
			Old: ptrToValue(oldOpts.AutoIncrement),
			New: ptrToValue(newOpts.AutoIncrement),
		}
	}

	if !ptrEqual(oldOpts.CharacterSet, newOpts.CharacterSet) {
		changes.CharacterSet = &FieldChange[any]{
			Old: ptrToValue(oldOpts.CharacterSet),
			New: ptrToValue(newOpts.CharacterSet),
		}
	}

	if !ptrEqual(oldOpts.Collate, newOpts.Collate) {
		changes.Collate = &FieldChange[any]{
			Old: ptrToValue(oldOpts.Collate),
			New: ptrToValue(newOpts.Collate),
		}
	}

	if !ptrEqual(oldOpts.Comment, newOpts.Comment) {
		changes.Comment = &FieldChange[any]{
			Old: ptrToValue(oldOpts.Comment),
			New: ptrToValue(newOpts.Comment),
		}
	}

	// Add more table options comparisons as needed...

	if changes.HasChanges() {
		return &TableOptionsDiff{
			ChangeType: ChangeTypeModified,
			OldOptions: oldOpts,
			NewOptions: newOpts,
			Changes:    changes,
		}
	}

	return nil
}

// comparePartitions compares partition options
func (a *TableDiffAnalyzer) comparePartitions(oldPart, newPart *parser.PartitionOptions) *PartitionDiff {
	if oldPart == nil && newPart == nil {
		return nil
	}

	if oldPart == nil {
		return &PartitionDiff{
			ChangeType:   ChangeTypeAdded,
			NewPartition: newPart,
			Changes:      &PartitionChanges{},
		}
	}

	if newPart == nil {
		return &PartitionDiff{
			ChangeType:   ChangeTypeRemoved,
			OldPartition: oldPart,
			Changes:      &PartitionChanges{},
		}
	}

	// Both exist, check for changes
	changes := &PartitionChanges{}

	// Compare partition attributes
	if oldPart.Type != newPart.Type {
		changes.Type = &FieldChange[string]{
			Old: oldPart.Type,
			New: newPart.Type,
		}
	}

	if oldPart.Linear != newPart.Linear {
		changes.Linear = &FieldChange[bool]{
			Old: oldPart.Linear,
			New: newPart.Linear,
		}
	}

	if !ptrEqual(oldPart.Expression, newPart.Expression) {
		changes.Expression = &FieldChange[any]{
			Old: ptrToValue(oldPart.Expression),
			New: ptrToValue(newPart.Expression),
		}
	}

	if !slices.Equal(oldPart.Columns, newPart.Columns) {
		changes.Columns = &FieldChange[[]string]{
			Old: oldPart.Columns,
			New: newPart.Columns,
		}
	}

	if !ptrEqual(oldPart.PartitionCount, newPart.PartitionCount) {
		changes.PartitionsCount = &FieldChange[any]{
			Old: ptrToValue(oldPart.PartitionCount),
			New: ptrToValue(newPart.PartitionCount),
		}
	}

	// Compare partition definitions (simplified)
	oldPartCount := len(oldPart.Partitions)
	newPartCount := len(newPart.Partitions)
	if oldPartCount != newPartCount {
		changes.PartitionDefinitions = &FieldChange[any]{
			Old: oldPartCount,
			New: newPartCount,
		}
	}

	if changes.HasChanges() {
		return &PartitionDiff{
			ChangeType:   ChangeTypeModified,
			OldPartition: oldPart,
			NewPartition: newPart,
			Changes:      changes,
		}
	}

	return nil
}
