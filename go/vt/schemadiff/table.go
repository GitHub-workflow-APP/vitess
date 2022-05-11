/*
Copyright 2022 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schemadiff

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	golcs "github.com/yudai/golcs"

	"vitess.io/vitess/go/mysql/collations"
	"vitess.io/vitess/go/vt/sqlparser"
)

//
type AlterTableEntityDiff struct {
	from       *CreateTableEntity
	to         *CreateTableEntity
	alterTable *sqlparser.AlterTable

	subsequentDiff *AlterTableEntityDiff
}

// IsEmpty implements EntityDiff
func (d *AlterTableEntityDiff) IsEmpty() bool {
	return d.Statement() == nil
}

// IsEmpty implements EntityDiff
func (d *AlterTableEntityDiff) Entities() (from Entity, to Entity) {
	return d.from, d.to
}

// Statement implements EntityDiff
func (d *AlterTableEntityDiff) Statement() sqlparser.Statement {
	if d == nil {
		return nil
	}
	return d.alterTable
}

// AlterTable returns the underlying sqlparser.AlterTable that was generated for the diff.
func (d *AlterTableEntityDiff) AlterTable() *sqlparser.AlterTable {
	if d == nil {
		return nil
	}
	return d.alterTable
}

// StatementString implements EntityDiff
func (d *AlterTableEntityDiff) StatementString() (s string) {
	if stmt := d.Statement(); stmt != nil {
		s = sqlparser.String(stmt)
	}
	return s
}

// CanonicalStatementString implements EntityDiff
func (d *AlterTableEntityDiff) CanonicalStatementString() (s string) {
	if stmt := d.Statement(); stmt != nil {
		s = sqlparser.CanonicalString(stmt)
	}
	return s
}

// SubsequentDiff implements EntityDiff
func (d *AlterTableEntityDiff) SubsequentDiff() EntityDiff {
	if d == nil {
		return nil
	}
	return d.subsequentDiff
}

// addSubsequentDiff adds a subsequent diff to the tail of the diff sequence
func (d *AlterTableEntityDiff) addSubsequentDiff(diff *AlterTableEntityDiff) {
	if d.subsequentDiff == nil {
		d.subsequentDiff = diff
	} else {
		d.subsequentDiff.addSubsequentDiff(diff)
	}
}

//
type CreateTableEntityDiff struct {
	createTable *sqlparser.CreateTable
}

// IsEmpty implements EntityDiff
func (d *CreateTableEntityDiff) IsEmpty() bool {
	return d.Statement() == nil
}

// IsEmpty implements EntityDiff
func (d *CreateTableEntityDiff) Entities() (from Entity, to Entity) {
	return nil, NewCreateTableEntity(d.createTable)
}

// Statement implements EntityDiff
func (d *CreateTableEntityDiff) Statement() sqlparser.Statement {
	if d == nil {
		return nil
	}
	return d.createTable
}

// CreateTable returns the underlying sqlparser.CreateTable that was generated for the diff.
func (d *CreateTableEntityDiff) CreateTable() *sqlparser.CreateTable {
	if d == nil {
		return nil
	}
	return d.createTable
}

// StatementString implements EntityDiff
func (d *CreateTableEntityDiff) StatementString() (s string) {
	if stmt := d.Statement(); stmt != nil {
		s = sqlparser.String(stmt)
	}
	return s
}

// CanonicalStatementString implements EntityDiff
func (d *CreateTableEntityDiff) CanonicalStatementString() (s string) {
	if stmt := d.Statement(); stmt != nil {
		s = sqlparser.CanonicalString(stmt)
	}
	return s
}

// SubsequentDiff implements EntityDiff
func (d *CreateTableEntityDiff) SubsequentDiff() EntityDiff {
	return nil
}

//
type DropTableEntityDiff struct {
	from      *CreateTableEntity
	dropTable *sqlparser.DropTable
}

// IsEmpty implements EntityDiff
func (d *DropTableEntityDiff) IsEmpty() bool {
	return d.Statement() == nil
}

// IsEmpty implements EntityDiff
func (d *DropTableEntityDiff) Entities() (from Entity, to Entity) {
	return d.from, nil
}

// Statement implements EntityDiff
func (d *DropTableEntityDiff) Statement() sqlparser.Statement {
	if d == nil {
		return nil
	}
	return d.dropTable
}

// DropTable returns the underlying sqlparser.DropTable that was generated for the diff.
func (d *DropTableEntityDiff) DropTable() *sqlparser.DropTable {
	if d == nil {
		return nil
	}
	return d.dropTable
}

// StatementString implements EntityDiff
func (d *DropTableEntityDiff) StatementString() (s string) {
	if stmt := d.Statement(); stmt != nil {
		s = sqlparser.String(stmt)
	}
	return s
}

// CanonicalStatementString implements EntityDiff
func (d *DropTableEntityDiff) CanonicalStatementString() (s string) {
	if stmt := d.Statement(); stmt != nil {
		s = sqlparser.CanonicalString(stmt)
	}
	return s
}

// SubsequentDiff implements EntityDiff
func (d *DropTableEntityDiff) SubsequentDiff() EntityDiff {
	return nil
}

// CreateTableEntity stands for a TABLE construct. It contains the table's CREATE statement.
type CreateTableEntity struct {
	sqlparser.CreateTable
}

func NewCreateTableEntity(c *sqlparser.CreateTable) *CreateTableEntity {
	entity := &CreateTableEntity{CreateTable: *c}
	entity.normalize()
	return entity
}

// normalize normalizes table definition:
// - setting names to all keys
// - table option case (upper/lower/special)
// The function returns this receiver as courtesy
func (c *CreateTableEntity) normalize() *CreateTableEntity {
	c.normalizeUnnamedKeys()
	c.normalizeUnnamedConstraints()
	c.normalizeTableOptions()
	c.normalizeColumnOptions()
	c.normalizePartitionOptions()
	return c
}

func (c *CreateTableEntity) normalizeTableOptions() {
	for _, opt := range c.CreateTable.TableSpec.Options {
		switch strings.ToUpper(opt.Name) {
		case "CHARSET", "COLLATE":
			opt.String = strings.ToLower(opt.String)
			if charset, ok := charsetAliases[opt.String]; ok {
				opt.String = charset
			}
		case "ENGINE":
			opt.String = strings.ToUpper(opt.String)
			if engineName, ok := engineCasing[opt.String]; ok {
				opt.String = engineName
			}
		case "ROW_FORMAT":
			opt.String = strings.ToUpper(opt.String)
		}
	}
}

// Right now we assume MySQL 8.0 for the collation normalization handling.
const mysqlCollationVersion = "8.0.0"

var collationEnv = collations.NewEnvironment(mysqlCollationVersion)

func defaultCharset() string {
	collation := collationEnv.LookupByID(collations.ID(collationEnv.DefaultConnectionCharset()))
	if collation == nil {
		return ""
	}
	return collation.Charset().Name()
}

func defaultCharsetCollation(charset string) string {
	// The collation tables are based on utf8, not the utf8mb3 alias.
	// We already normalize to utf8mb3 to be explicit, so we have to
	// map it back here to find the default collation for utf8mb3.
	if charset == "utf8mb3" {
		charset = "utf8"
	}
	collation := collationEnv.DefaultCollationForCharset(charset)
	if collation == nil {
		return ""
	}
	return collation.Name()
}

func (c *CreateTableEntity) normalizeColumnOptions() {
	tableCharset := defaultCharset()
	tableCollation := ""
	for _, option := range c.CreateTable.TableSpec.Options {
		switch strings.ToUpper(option.Name) {
		case "CHARSET":
			tableCharset = option.String
		case "COLLATE":
			tableCollation = option.String
		}
	}
	defaultCollation := defaultCharsetCollation(tableCharset)
	if tableCollation == "" {
		tableCollation = defaultCollation
	}

	for _, col := range c.CreateTable.TableSpec.Columns {
		if col.Type.Options == nil {
			col.Type.Options = &sqlparser.ColumnTypeOptions{}
		}

		// Map known lowercase fields to always be lowercase
		col.Type.Type = strings.ToLower(col.Type.Type)
		col.Type.Charset = strings.ToLower(col.Type.Charset)
		col.Type.Options.Collate = strings.ToLower(col.Type.Options.Collate)

		// See https://dev.mysql.com/doc/refman/8.0/en/create-table.html
		// If neither NULL nor NOT NULL is specified, the column is treated as though NULL had been specified.
		// That documentation though is not 100% true. There's an exception, and that is
		// the `explicit_defaults_for_timestamp` flag. When that is disabled (the default on 5.7),
		// a timestamp defaults to `NOT NULL`.
		//
		// We opt here to instead remove that difference and always then add `NULL` and treat
		// `explicit_defaults_for_timestamp` as always enabled in the context of DDL for diffing.
		if col.Type.Type == "timestamp" {
			if col.Type.Options.Null == nil || *col.Type.Options.Null {
				timestampNull := true
				col.Type.Options.Null = &timestampNull
			}
		} else {
			if col.Type.Options.Null != nil && *col.Type.Options.Null {
				col.Type.Options.Null = nil
			}
		}
		if col.Type.Options.Null == nil || *col.Type.Options.Null {
			// If `DEFAULT NULL` is specified and the column allows NULL,
			// we drop that in the normalized form since that is equivalent to the default value.
			// See also https://dev.mysql.com/doc/refman/8.0/en/data-type-defaults.html
			if _, ok := col.Type.Options.Default.(*sqlparser.NullVal); ok {
				col.Type.Options.Default = nil
			}
		}

		// Map any charset aliases to the real charset. This applies mainly right
		// now to utf8 being an alias for utf8mb3.
		if charset, ok := charsetAliases[col.Type.Charset]; ok {
			col.Type.Charset = charset
		}

		// Remove any lengths for integral types since it is deprecated there and
		// doesn't mean anything anymore.
		if _, ok := integralTypes[col.Type.Type]; ok {
			col.Type.Length = nil
			// Remove zerofill for integral types but keep the behavior that this marks the value
			// as unsigned
			if col.Type.Zerofill {
				col.Type.Zerofill = false
				col.Type.Unsigned = true
			}
		}

		if _, ok := charsetTypes[col.Type.Type]; ok {
			// If the charset is explicitly configured and it mismatches, we don't normalize
			// anything for charsets or collations and move on.
			if col.Type.Charset != "" && col.Type.Charset != tableCharset {
				continue
			}

			// Alright, first check if both charset and collation are the same as
			// the table level options, in that case we can remove both since that's equivalent.
			if col.Type.Charset == tableCharset && col.Type.Options.Collate == tableCollation {
				col.Type.Charset = ""
				col.Type.Options.Collate = ""
			}
			// If we have no charset or collation defined, we inherit the table defaults
			// and don't need to do anything here and can continue to the next column.
			// It doesn't matter if that's because it's not defined, or if it was because
			// it was explicitly set to the same values.
			if col.Type.Charset == "" && col.Type.Options.Collate == "" {
				continue
			}

			// We have a matching charset as the default, but it is explicitly set. In that
			// case we still want to clear it, but set the default collation for the given charset
			// if no collation is defined yet. We set then the collation to the default collation.
			if col.Type.Charset != "" {
				col.Type.Charset = ""
				if col.Type.Options.Collate == "" {
					col.Type.Options.Collate = defaultCollation
				}
			}

			// We now have one case left, which is when we have set a collation but it's the same
			// as the table level. In that case, we can clear it since that is equivalent.
			if col.Type.Options.Collate == tableCollation {
				col.Type.Options.Collate = ""
			}
		}
	}
}

func (c *CreateTableEntity) normalizePartitionOptions() {
	if c.CreateTable.TableSpec.PartitionOption == nil {
		return
	}

	for _, def := range c.CreateTable.TableSpec.PartitionOption.Definitions {
		if def.Options == nil || def.Options.Engine == nil {
			continue
		}

		def.Options.Engine.Name = strings.ToUpper(def.Options.Engine.Name)
		if engineName, ok := engineCasing[def.Options.Engine.Name]; ok {
			def.Options.Engine.Name = engineName
		}
	}
}

func (c *CreateTableEntity) normalizeUnnamedKeys() {
	// let's ensure all keys have names
	keyNameExists := map[string]bool{}
	// first, we iterate and take note for all keys that do already have names
	for _, key := range c.CreateTable.TableSpec.Indexes {
		if name := key.Info.Name.String(); name != "" {
			keyNameExists[name] = true
		}
	}
	// now, let's look at keys that do not have names, and assign them new names
	for _, key := range c.CreateTable.TableSpec.Indexes {
		if name := key.Info.Name.String(); name == "" {
			// we know there must be at least one column covered by this key
			var colName string
			if len(key.Columns) > 0 {
				expressionFound := false
				for _, col := range key.Columns {
					if col.Expression != nil {
						expressionFound = true
					}
				}
				if expressionFound {
					// that's the name MySQL picks for an unnamed key when there's at least one functional index expression
					colName = "functional_index"
				} else {
					// like MySQL, we first try to call our index by the name of the first column:
					colName = key.Columns[0].Column.String()
				}
			}
			suggestedKeyName := colName
			// now let's see if that name is taken; if it is, enumerate new news until we find a free name
			for enumerate := 2; keyNameExists[suggestedKeyName]; enumerate++ {
				suggestedKeyName = fmt.Sprintf("%s_%d", colName, enumerate)
			}
			// OK we found a free slot!
			key.Info.Name = sqlparser.NewColIdent(suggestedKeyName)
			keyNameExists[suggestedKeyName] = true
		}
	}
}

func (c *CreateTableEntity) normalizeUnnamedConstraints() {
	// let's ensure all keys have names
	constraintNameExists := map[string]bool{}
	// first, we iterate and take note for all keys that do already have names
	for _, constraint := range c.CreateTable.TableSpec.Constraints {
		if name := constraint.Name.String(); name != "" {
			constraintNameExists[name] = true
		}
	}

	// now, let's look at keys that do not have names, and assign them new names
	for _, constraint := range c.CreateTable.TableSpec.Constraints {
		if name := constraint.Name.String(); name == "" {
			nameFormat := "%s_chk_%d"
			if _, fk := constraint.Details.(*sqlparser.ForeignKeyDefinition); fk {
				nameFormat = "%s_ibfk_%d"
			}
			suggestedCheckName := fmt.Sprintf(nameFormat, c.CreateTable.Table.Name.String(), 1)
			// now let's see if that name is taken; if it is, enumerate new news until we find a free name
			for enumerate := 2; constraintNameExists[suggestedCheckName]; enumerate++ {
				suggestedCheckName = fmt.Sprintf(nameFormat, c.CreateTable.Table.Name.String(), enumerate)
			}
			// OK we found a free slot!
			constraint.Name = sqlparser.NewColIdent(suggestedCheckName)
			constraintNameExists[suggestedCheckName] = true
		}
	}
}

// Name implements Entity interface
func (c *CreateTableEntity) Name() string {
	return c.CreateTable.GetTable().Name.String()
}

// Diff implements Entity interface function
func (c *CreateTableEntity) Diff(other Entity, hints *DiffHints) (EntityDiff, error) {
	otherCreateTable, ok := other.(*CreateTableEntity)
	if !ok {
		return nil, ErrEntityTypeMismatch
	}
	if hints.StrictIndexOrdering {
		return nil, ErrStrictIndexOrderingUnsupported
	}
	if c.CreateTable.TableSpec == nil {
		return nil, ErrUnexpectedTableSpec
	}

	d, err := c.TableDiff(otherCreateTable, hints)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Diff compares this table statement with another table statement, and sees what it takes to
// change this table to look like the other table.
// It returns an AlterTable statement if changes are found, or nil if not.
// the other table may be of different name; its name is ignored.
func (c *CreateTableEntity) TableDiff(other *CreateTableEntity, hints *DiffHints) (*AlterTableEntityDiff, error) {
	otherStmt := other.CreateTable
	otherStmt.Table = c.CreateTable.Table

	if !c.CreateTable.IsFullyParsed() {
		return nil, ErrNotFullyParsed
	}
	if !otherStmt.IsFullyParsed() {
		return nil, ErrNotFullyParsed
	}

	format := sqlparser.CanonicalString(&c.CreateTable)
	otherFormat := sqlparser.CanonicalString(&otherStmt)
	if format == otherFormat {
		return nil, nil
	}

	alterTable := &sqlparser.AlterTable{
		Table: otherStmt.Table,
	}
	diffedTableCharset := ""
	var partitionSpecs []*sqlparser.PartitionSpec
	{
		t1Options := c.CreateTable.TableSpec.Options
		t2Options := other.CreateTable.TableSpec.Options
		diffedTableCharset = c.diffTableCharset(t1Options, t2Options)
	}
	{
		// diff columns
		// ordered columns for both tables:
		t1Columns := c.CreateTable.TableSpec.Columns
		t2Columns := other.CreateTable.TableSpec.Columns
		c.diffColumns(alterTable, t1Columns, t2Columns, hints, (diffedTableCharset != ""))
	}
	{
		// diff keys
		// ordered keys for both tables:
		t1Keys := c.CreateTable.TableSpec.Indexes
		t2Keys := other.CreateTable.TableSpec.Indexes
		c.diffKeys(alterTable, t1Keys, t2Keys, hints)
	}
	{
		// diff constraints
		// ordered constraints for both tables:
		t1Constraints := c.CreateTable.TableSpec.Constraints
		t2Constraints := other.CreateTable.TableSpec.Constraints
		c.diffConstraints(alterTable, t1Constraints, t2Constraints, hints)
	}
	{
		// diff partitions
		// ordered keys for both tables:
		t1Partitions := c.CreateTable.TableSpec.PartitionOption
		t2Partitions := other.CreateTable.TableSpec.PartitionOption
		var err error
		partitionSpecs, err = c.diffPartitions(alterTable, t1Partitions, t2Partitions, hints)
		if err != nil {
			return nil, err
		}
	}
	{
		// diff table options
		// ordered keys for both tables:
		t1Options := c.CreateTable.TableSpec.Options
		t2Options := other.CreateTable.TableSpec.Options
		if err := c.diffOptions(alterTable, t1Options, t2Options, hints); err != nil {
			return nil, err
		}
	}
	if len(alterTable.AlterOptions) == 0 && alterTable.PartitionOption == nil && alterTable.PartitionSpec == nil && len(partitionSpecs) == 0 {
		// it's possible that the table definitions are different, and still there's no
		// "real" difference. Reasons could be:
		// - reordered keys -- we treat that as non-diff
		return nil, nil
	}
	if len(partitionSpecs) == 0 {
		return &AlterTableEntityDiff{alterTable: alterTable, from: c, to: other}, nil
	}
	// partitionSpecs has multiple entries
	if len(alterTable.AlterOptions) > 0 ||
		alterTable.PartitionOption != nil ||
		alterTable.PartitionSpec != nil {
		return nil, ErrMixedPartitionAndNonPartitionChanges
	}
	var parentAlterTableEntityDiff *AlterTableEntityDiff
	for _, partitionSpec := range partitionSpecs {
		alterTable := &sqlparser.AlterTable{
			Table:         otherStmt.Table,
			PartitionSpec: partitionSpec,
		}
		diff := &AlterTableEntityDiff{alterTable: alterTable, from: c, to: other}
		if parentAlterTableEntityDiff == nil {
			parentAlterTableEntityDiff = diff
		} else {
			parentAlterTableEntityDiff.addSubsequentDiff(diff)
		}
	}
	return parentAlterTableEntityDiff, nil
}

func (c *CreateTableEntity) diffTableCharset(
	t1Options sqlparser.TableOptions,
	t2Options sqlparser.TableOptions,
) string {
	getcharset := func(options sqlparser.TableOptions) string {
		for _, option := range options {
			if strings.ToUpper(option.Name) == "CHARSET" {
				return option.String
			}
		}
		return ""
	}
	t1Charset := getcharset(t1Options)
	t2Charset := getcharset(t2Options)
	if t1Charset != t2Charset {
		return t2Charset
	}
	return ""
}

// isDefaultTableOptionValue sees if the value for a TableOption is also its default value
func isDefaultTableOptionValue(option *sqlparser.TableOption) bool {
	switch strings.ToUpper(option.Name) {
	case "CHECKSUM":
		return sqlparser.CanonicalString(option.Value) == "0"
	case "COMMENT":
		return option.String == ""
	case "COMPRESSION":
		return sqlparser.CanonicalString(option.Value) == "" || sqlparser.CanonicalString(option.Value) == "''"
	case "CONNECTION":
		return sqlparser.CanonicalString(option.Value) == "" || sqlparser.CanonicalString(option.Value) == "''"
	case "DATA DIRECTORY":
		return sqlparser.CanonicalString(option.Value) == "" || sqlparser.CanonicalString(option.Value) == "''"
	case "DELAY_KEY_WRITE":
		return sqlparser.CanonicalString(option.Value) == "0"
	case "ENCRYPTION":
		return sqlparser.CanonicalString(option.Value) == "N"
	case "INDEX DIRECTORY":
		return sqlparser.CanonicalString(option.Value) == "" || sqlparser.CanonicalString(option.Value) == "''"
	case "KEY_BLOCK_SIZE":
		return sqlparser.CanonicalString(option.Value) == "0"
	case "MAX_ROWS":
		return sqlparser.CanonicalString(option.Value) == "0"
	case "MIN_ROWS":
		return sqlparser.CanonicalString(option.Value) == "0"
	case "PACK_KEYS":
		return strings.EqualFold(option.String, "DEFAULT")
	case "ROW_FORMAT":
		return strings.EqualFold(option.String, "DEFAULT")
	case "STATS_AUTO_RECALC":
		return strings.EqualFold(option.String, "DEFAULT")
	case "STATS_PERSISTENT":
		return strings.EqualFold(option.String, "DEFAULT")
	case "STATS_SAMPLE_PAGES":
		return strings.EqualFold(option.String, "DEFAULT")
	default:
		return false
	}
}

func (c *CreateTableEntity) diffOptions(alterTable *sqlparser.AlterTable,
	t1Options sqlparser.TableOptions,
	t2Options sqlparser.TableOptions,
	hints *DiffHints,
) error {
	t1OptionsMap := map[string]*sqlparser.TableOption{}
	t2OptionsMap := map[string]*sqlparser.TableOption{}
	for _, option := range t1Options {
		t1OptionsMap[option.Name] = option
	}
	for _, option := range t2Options {
		t2OptionsMap[option.Name] = option
	}
	alterTableOptions := sqlparser.TableOptions{}
	// dropped options
	for _, t1Option := range t1Options {
		if _, ok := t2OptionsMap[t1Option.Name]; !ok {
			// option exists in t1 but not in t2, hence it is dropped
			var tableOption *sqlparser.TableOption
			switch strings.ToUpper(t1Option.Name) {
			case "AUTO_INCREMENT":
				// skip
			case "AVG_ROW_LENGTH":
				// skip. MyISAM only, not interesting
			case "CHECKSUM":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewIntLiteral("0")}
			case "COLLATE":
				// skip. the default collation is applied per CHARSET
			case "COMMENT":
				tableOption = &sqlparser.TableOption{String: ""}
			case "COMPRESSION":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewStrLiteral("")}
			case "CONNECTION":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewStrLiteral("")}
			case "DATA DIRECTORY":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewStrLiteral("")}
			case "DELAY_KEY_WRITE":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewIntLiteral("0")}
			case "ENCRYPTION":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewStrLiteral("N")}
			case "ENGINE":
				// skip
			case "INDEX DIRECTORY":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewStrLiteral("")}
			case "INSERT_METHOD":
				// MyISAM only. skip
			case "KEY_BLOCK_SIZE":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewIntLiteral("0")}
			case "MAX_ROWS":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewIntLiteral("0")}
			case "MIN_ROWS":
				tableOption = &sqlparser.TableOption{Value: sqlparser.NewIntLiteral("0")}
			case "PACK_KEYS":
				tableOption = &sqlparser.TableOption{String: "DEFAULT"}
			case "PASSWORD":
				// unused option. skip
			case "ROW_FORMAT":
				tableOption = &sqlparser.TableOption{String: "DEFAULT"}
			case "STATS_AUTO_RECALC":
				tableOption = &sqlparser.TableOption{String: "DEFAULT"}
			case "STATS_PERSISTENT":
				tableOption = &sqlparser.TableOption{String: "DEFAULT"}
			case "STATS_SAMPLE_PAGES":
				tableOption = &sqlparser.TableOption{String: "DEFAULT"}
			case "TABLESPACE":
				// not supporting the change, skip
			case "UNION":
				// MyISAM/MERGE only. Skip
			default:
				return ErrUnsupportedTableOption
			}
			if tableOption != nil {
				tableOption.Name = t1Option.Name
				alterTableOptions = append(alterTableOptions, tableOption)
			}
		}

	}
	// changed options
	for _, t2Option := range t2Options {
		if t1Option, ok := t1OptionsMap[t2Option.Name]; ok {
			options1 := sqlparser.TableOptions{t1Option}
			options2 := sqlparser.TableOptions{t2Option}
			if sqlparser.CanonicalString(options1) != sqlparser.CanonicalString(options2) {
				// options are different.
				// However, we don't automatically apply these changes. It depends on the option!
				switch strings.ToUpper(t1Option.Name) {
				case "AUTO_INCREMENT":
					switch hints.AutoIncrementStrategy {
					case AutoIncrementApplyAlways:
						alterTableOptions = append(alterTableOptions, t2Option)
					case AutoIncrementApplyHigher:
						option1AutoIncrement, err := strconv.ParseInt(t1Option.Value.Val, 10, 64)
						if err != nil {
							return err
						}
						option2AutoIncrement, err := strconv.ParseInt(t2Option.Value.Val, 10, 64)
						if err != nil {
							return err
						}
						if option2AutoIncrement > option1AutoIncrement {
							// never decrease AUTO_INCREMENT. Only increase
							alterTableOptions = append(alterTableOptions, t2Option)
						}
					case AutoIncrementIgnore:
						// do not apply
					}
				default:
					// Apply the new options
					alterTableOptions = append(alterTableOptions, t2Option)
				}
			}
		}
	}
	// added options
	for _, t2Option := range t2Options {
		if _, ok := t1OptionsMap[t2Option.Name]; !ok {
			switch strings.ToUpper(t2Option.Name) {
			case "AUTO_INCREMENT":
				switch hints.AutoIncrementStrategy {
				case AutoIncrementApplyAlways, AutoIncrementApplyHigher:
					alterTableOptions = append(alterTableOptions, t2Option)
				case AutoIncrementIgnore:
					// do not apply
				}
			default:
				alterTableOptions = append(alterTableOptions, t2Option)
			}
		}
	}

	if len(alterTableOptions) > 0 {
		alterTable.AlterOptions = append(alterTable.AlterOptions, alterTableOptions)
	}
	return nil
}

// rangePartitionsAddedRemoved returns true when:
// - both table partitions are RANGE type
// - there is exactly one consequitive non-empty shared sequence of partitions (same names, same range values, in same order)
// - table1 may have non-empty list of partitions _preceding_ this sequence, and table2 may not
// - table2 may have non-empty list of partitions _following_ this sequence, and table1 may not
func (c *CreateTableEntity) isRangePartitionsRotation(
	t1Partitions *sqlparser.PartitionOption,
	t2Partitions *sqlparser.PartitionOption,
) (bool, []*sqlparser.PartitionSpec, error) {
	// Validate that both tables have range partitioning
	if t1Partitions.Type != t2Partitions.Type {
		return false, nil, nil
	}
	if t1Partitions.Type != sqlparser.RangeType {
		return false, nil, nil
	}
	definitions1 := t1Partitions.Definitions
	definitions2 := t2Partitions.Definitions
	// there has to be a non-empty shared list, therefore both definitions must be non-empty:
	if len(definitions1) == 0 {
		return false, nil, nil
	}
	if len(definitions2) == 0 {
		return false, nil, nil
	}
	droppedPartitions1 := []*sqlparser.PartitionDefinition{}
	// It's OK for prefix of t1 partitions to be nonexistent in t2 (as they may have been rotated away in t2)
	for len(definitions1) > 0 && sqlparser.CanonicalString(definitions1[0]) != sqlparser.CanonicalString(definitions2[0]) {
		droppedPartitions1 = append(droppedPartitions1, definitions1[0])
		definitions1 = definitions1[1:]
	}
	if len(definitions1) == 0 {
		// We've exhaused definition1 trying to find a shared partition with definitions2. Nothing found.
		// so there is no shared sequence between the two tables.
		return false, nil, nil
	}
	if len(definitions1) > len(definitions2) {
		return false, nil, nil
	}
	// To save computation, and because we've already shown that sqlparser.CanonicalString(definitions1[0]) == sqlparser.CanonicalString(definitions2[0]),
	// we can skip one element
	definitions1 = definitions1[1:]
	definitions2 = definitions2[1:]
	// Now let's ensure that whatever is remaining in definitions1 is an exact match for a prefix of definitions2
	// It's ok if we end up with leftover elements in definition2
	for len(definitions1) > 0 {
		if sqlparser.CanonicalString(definitions1[0]) != sqlparser.CanonicalString(definitions2[0]) {
			return false, nil, nil
		}
		definitions1 = definitions1[1:]
		definitions2 = definitions2[1:]
	}
	partitionSpecs := []*sqlparser.PartitionSpec{}
	addedPartitions2 := definitions2
	for _, p := range droppedPartitions1 {
		partitionSpec := &sqlparser.PartitionSpec{
			Action: sqlparser.DropAction,
			Names:  []sqlparser.ColIdent{p.Name},
		}
		partitionSpecs = append(partitionSpecs, partitionSpec)
	}
	for _, p := range addedPartitions2 {
		partitionSpec := &sqlparser.PartitionSpec{
			Action:      sqlparser.AddAction,
			Definitions: []*sqlparser.PartitionDefinition{p},
		}
		partitionSpecs = append(partitionSpecs, partitionSpec)
	}
	return true, partitionSpecs, nil
}

func (c *CreateTableEntity) diffPartitions(alterTable *sqlparser.AlterTable,
	t1Partitions *sqlparser.PartitionOption,
	t2Partitions *sqlparser.PartitionOption,
	hints *DiffHints,
) (partitionSpecs []*sqlparser.PartitionSpec, err error) {
	switch {
	case t1Partitions == nil && t2Partitions == nil:
		return nil, nil
	case t1Partitions == nil:
		// add partitioning
		alterTable.PartitionOption = t2Partitions
	case t2Partitions == nil:
		// remove partitioning
		partitionSpec := &sqlparser.PartitionSpec{
			Action: sqlparser.RemoveAction,
			IsAll:  true,
		}
		alterTable.PartitionSpec = partitionSpec
	case sqlparser.CanonicalString(t1Partitions) == sqlparser.CanonicalString(t2Partitions):
		// identical partitioning
		return nil, nil
	default:
		// partitioning was changed
		// For most cases, we produce a complete re-partitioing schema: we don't try and figure out the minimal
		// needed change. For example, maybe the minimal change is to REORGANIZE a specific partition and split
		// into two, thus unaffecting the rest of the partitions. But we don't evaluate that, we just set a
		// complete new ALTER TABLE ... PARTITION BY statement.
		// The idea is that it doesn't matter: we're not looking to do optimal in-place ALTERs, we run
		// Online DDL alters, where we create a new table anyway. Thus, the optimization is meaningless.

		// Having said that, we _do_ analyze the scenario of a RANGE partitioning rotation of partitions:
		// where zero or more partitions may have been dropped from the earlier range, and zero or more
		// partitions have been added with a later range:
		isRotation, partitionSpecs, err := c.isRangePartitionsRotation(t1Partitions, t2Partitions)
		if err != nil {
			return nil, err
		}
		if isRotation {
			switch hints.RangeRotationStrategy {
			case RangeRotationIgnore:
				return nil, nil
			case RangeRotationStatements:
				if len(partitionSpecs) == 1 {
					alterTable.PartitionSpec = partitionSpecs[0]
					partitionSpecs = nil
				}
				return partitionSpecs, nil
			case RangeRotationFullSpec:
				// proceed to return a full rebuild
			}
		}
		alterTable.PartitionOption = t2Partitions
	}
	return nil, nil
}

func (c *CreateTableEntity) diffConstraints(alterTable *sqlparser.AlterTable,
	t1Constraints []*sqlparser.ConstraintDefinition,
	t2Constraints []*sqlparser.ConstraintDefinition,
	hints *DiffHints,
) {
	t1ConstraintsMap := map[string]*sqlparser.ConstraintDefinition{}
	t2ConstraintsMap := map[string]*sqlparser.ConstraintDefinition{}
	for _, constraint := range t1Constraints {
		t1ConstraintsMap[constraint.Name.String()] = constraint
	}
	for _, constraint := range t2Constraints {
		t2ConstraintsMap[constraint.Name.String()] = constraint
	}

	dropConstraintStatement := func(constraint *sqlparser.ConstraintDefinition) *sqlparser.DropKey {
		if _, fk := constraint.Details.(*sqlparser.ForeignKeyDefinition); fk {
			return &sqlparser.DropKey{Name: constraint.Name, Type: sqlparser.ForeignKeyType}
		}
		return &sqlparser.DropKey{Name: constraint.Name, Type: sqlparser.CheckKeyType}
	}

	// evaluate dropped constraints
	//
	for _, t1Constraint := range t1Constraints {
		if _, ok := t2ConstraintsMap[t1Constraint.Name.String()]; !ok {
			// column exists in t1 but not in t2, hence it is dropped
			dropConstraint := dropConstraintStatement(t1Constraint)
			alterTable.AlterOptions = append(alterTable.AlterOptions, dropConstraint)
		}
	}

	for _, t2Constraint := range t2Constraints {
		t2ConstraintName := t2Constraint.Name.String()
		// evaluate modified & added constraints:
		//
		if t1Constraint, ok := t1ConstraintsMap[t2ConstraintName]; ok {
			// constraint exists in both tables
			// check diff between before/after columns:
			if sqlparser.CanonicalString(t2Constraint) != sqlparser.CanonicalString(t1Constraint) {
				// constraints with same name have different definition.
				// First we check if this is only the enforced setting that changed which can
				// be directly altered.
				check1Details, ok1 := t1Constraint.Details.(*sqlparser.CheckConstraintDefinition)
				check2Details, ok2 := t2Constraint.Details.(*sqlparser.CheckConstraintDefinition)
				if ok1 && ok2 && sqlparser.CanonicalString(check1Details.Expr) == sqlparser.CanonicalString(check2Details.Expr) {
					// We have the same expression, so we have a different Enforced here
					alterConstraint := &sqlparser.AlterCheck{
						Name:     t2Constraint.Name,
						Enforced: check2Details.Enforced,
					}
					alterTable.AlterOptions = append(alterTable.AlterOptions, alterConstraint)
					continue
				}

				// There's another change, so we need to drop and add.
				dropConstraint := dropConstraintStatement(t1Constraint)
				addConstraint := &sqlparser.AddConstraintDefinition{
					ConstraintDefinition: t2Constraint,
				}
				alterTable.AlterOptions = append(alterTable.AlterOptions, dropConstraint)
				alterTable.AlterOptions = append(alterTable.AlterOptions, addConstraint)
			}
		} else {
			// constraint exists in t2 but not in t1, hence it is added
			addConstraint := &sqlparser.AddConstraintDefinition{
				ConstraintDefinition: t2Constraint,
			}
			alterTable.AlterOptions = append(alterTable.AlterOptions, addConstraint)
		}
	}
}

func (c *CreateTableEntity) diffKeys(alterTable *sqlparser.AlterTable,
	t1Keys []*sqlparser.IndexDefinition,
	t2Keys []*sqlparser.IndexDefinition,
	hints *DiffHints,
) {
	t1KeysMap := map[string]*sqlparser.IndexDefinition{}
	t2KeysMap := map[string]*sqlparser.IndexDefinition{}
	for _, key := range t1Keys {
		t1KeysMap[key.Info.Name.String()] = key
	}
	for _, key := range t2Keys {
		t2KeysMap[key.Info.Name.String()] = key
	}

	dropKeyStatement := func(name sqlparser.ColIdent) *sqlparser.DropKey {
		dropKey := &sqlparser.DropKey{}
		if dropKey.Name.String() == "PRIMARY" {
			dropKey.Type = sqlparser.PrimaryKeyType
		} else {
			dropKey.Type = sqlparser.NormalKeyType
			dropKey.Name = name
		}
		return dropKey
	}

	// evaluate dropped keys
	//
	for _, t1Key := range t1Keys {
		if _, ok := t2KeysMap[t1Key.Info.Name.String()]; !ok {
			// column exists in t1 but not in t2, hence it is dropped
			dropKey := dropKeyStatement(t1Key.Info.Name)
			alterTable.AlterOptions = append(alterTable.AlterOptions, dropKey)
		}
	}

	for _, t2Key := range t2Keys {
		t2KeyName := t2Key.Info.Name.String()
		// evaluate modified & added keys:
		//
		if t1Key, ok := t1KeysMap[t2KeyName]; ok {
			// key exists in both tables
			// check diff between before/after columns:
			if sqlparser.CanonicalString(t2Key) != sqlparser.CanonicalString(t1Key) {
				// keys with same name have different definition. There is no ALTER INDEX statement,
				// we're gonna drop and create.
				dropKey := dropKeyStatement(t1Key.Info.Name)
				addKey := &sqlparser.AddIndexDefinition{
					IndexDefinition: t2Key,
				}
				alterTable.AlterOptions = append(alterTable.AlterOptions, dropKey)
				alterTable.AlterOptions = append(alterTable.AlterOptions, addKey)
			}
		} else {
			// key exists in t2 but not in t1, hence it is added
			addKey := &sqlparser.AddIndexDefinition{
				IndexDefinition: t2Key,
			}
			alterTable.AlterOptions = append(alterTable.AlterOptions, addKey)
		}
	}
}

// evaluateColumnReordering produces a minimal reordering set of columns. To elaborate:
// The function receives two sets of columns. the two must be permutations of one another. Specifically,
// these are the columns shared between the from&to tables.
// The function uses longest-common-subsequence (lcs) algorithm to compute which columns should not be moved.
// any column not in the lcs need to be reordered.
// The function a map of column names that need to be reordered, and the index into which they are reordered.
func evaluateColumnReordering(t1SharedColumns, t2SharedColumns []*sqlparser.ColumnDefinition) map[string]int {
	minimalColumnReordering := map[string]int{}

	t1SharedColNames := []interface{}{}
	for _, col := range t1SharedColumns {
		t1SharedColNames = append(t1SharedColNames, col.Name.String())
	}
	t2SharedColNames := []interface{}{}
	for _, col := range t2SharedColumns {
		t2SharedColNames = append(t2SharedColNames, col.Name.String())
	}

	lcs := golcs.New(t1SharedColNames, t2SharedColNames)
	lcsNames := map[string]bool{}
	for _, v := range lcs.Values() {
		lcsNames[v.(string)] = true
	}
	for i, t2Col := range t2SharedColumns {
		t2ColName := t2Col.Name.String()
		// see if this column is in longest common subsequence. If so, no need to reorder it. If not, it must be reordered.
		if _, ok := lcsNames[t2ColName]; !ok {
			minimalColumnReordering[t2ColName] = i
		}
	}

	return minimalColumnReordering
}

// Diff compares this table statement with another table statement, and sees what it takes to
// change this table to look like the other table.
// It returns an AlterTable statement if changes are found, or nil if not.
// the other table may be of different name; its name is ignored.
func (c *CreateTableEntity) diffColumns(alterTable *sqlparser.AlterTable,
	t1Columns []*sqlparser.ColumnDefinition,
	t2Columns []*sqlparser.ColumnDefinition,
	hints *DiffHints,
	tableCharsetChanged bool,
) {
	// map columns by names for easy access
	t1ColumnsMap := map[string]*sqlparser.ColumnDefinition{}
	t2ColumnsMap := map[string]*sqlparser.ColumnDefinition{}
	for _, col := range t1Columns {
		t1ColumnsMap[col.Name.String()] = col
	}
	for _, col := range t2Columns {
		t2ColumnsMap[col.Name.String()] = col
	}

	// For purpose of column reordering detection, we maintain a list of
	// shared columns, by order of appearance in t1
	t1SharedColumns := []*sqlparser.ColumnDefinition{}

	// evaluate dropped columns
	//
	for _, t1Col := range t1Columns {
		if _, ok := t2ColumnsMap[t1Col.Name.String()]; ok {
			t1SharedColumns = append(t1SharedColumns, t1Col)
		} else {
			// column exists in t1 but not in t2, hence it is dropped
			dropColumn := &sqlparser.DropColumn{
				Name: getColName(&t1Col.Name),
			}
			alterTable.AlterOptions = append(alterTable.AlterOptions, dropColumn)
		}
	}

	// For purpose of column reordering detection, we maintain a list of
	// shared columns, by order of appearance in t2
	t2SharedColumns := []*sqlparser.ColumnDefinition{}
	for _, t2Col := range t2Columns {
		t2ColName := t2Col.Name.String()
		if _, ok := t1ColumnsMap[t2ColName]; ok {
			// column exists in both tables
			t2SharedColumns = append(t2SharedColumns, t2Col)
		}
	}

	// evaluate modified columns
	//
	columnReordering := evaluateColumnReordering(t1SharedColumns, t2SharedColumns)
	for _, t2Col := range t2SharedColumns {
		t2ColName := t2Col.Name.String()
		// we know that column exists in both tables
		t1Col := t1ColumnsMap[t2ColName]
		t1ColEntity := NewColumnDefinitionEntity(t1Col)
		t2ColEntity := NewColumnDefinitionEntity(t2Col)

		// check diff between before/after columns:
		modifyColumnDiff := t1ColEntity.ColumnDiff(t2ColEntity, hints)
		if modifyColumnDiff == nil {
			// even if there's no apparent change, there can still be implciit changes
			// it is possible that the table charset is changed. the column may be some col1 TEXT NOT NULL, possibly in both varsions 1 and 2,
			// but implicitly the column has changed its characters set. So we need to explicitly ass a MODIFY COLUMN statement, so that
			// MySQL rebuilds it.
			if tableCharsetChanged && t2ColEntity.IsTextual() && t2Col.Type.Charset == "" {
				modifyColumnDiff = NewModifyColumnDiffByDefinition(t2Col)
			}
		}
		// It is also possible that a column is reordered. Whether the column definition has
		// or hasn't changed, if a column is reordered then that's a change of its own!
		if columnReorderIndex, ok := columnReordering[t2ColName]; ok {
			// seems like we previously evaluated that this column should be reordered
			if modifyColumnDiff == nil {
				// create column change
				modifyColumnDiff = NewModifyColumnDiffByDefinition(t2Col)
			}
			if columnReorderIndex == 0 {
				modifyColumnDiff.modifyColumn.First = true
			} else {
				modifyColumnDiff.modifyColumn.After = getColName(&t2SharedColumns[columnReorderIndex-1].Name)
			}
		}
		if modifyColumnDiff != nil {
			// column definition or ordering has changed
			alterTable.AlterOptions = append(alterTable.AlterOptions, modifyColumnDiff.modifyColumn)
		}
	}
	// Evaluate added columns
	//
	// Every added column is obviously a diff. But on top of that, we are also interested to know
	// if the column is added somewhere in between existing columns rather than appended to the
	// end of existing columns list.
	expectAppendIndex := len(t2SharedColumns)
	for t2ColIndex, t2Col := range t2Columns {
		t2ColName := t2Col.Name.String()
		if _, ok := t1ColumnsMap[t2ColName]; !ok {
			// column exists in t2 but not in t1, hence it is added
			addColumn := &sqlparser.AddColumns{
				Columns: []*sqlparser.ColumnDefinition{t2Col},
			}
			if t2ColIndex < expectAppendIndex {
				// This column is added somewhere in between existing columns, not appended at end of column list
				if t2ColIndex == 0 {
					addColumn.First = true
				} else {
					addColumn.After = getColName(&t2Columns[t2ColIndex-1].Name)
				}
			}
			expectAppendIndex++
			alterTable.AlterOptions = append(alterTable.AlterOptions, addColumn)
		}
	}
}

// Create implements Entity interface
func (c *CreateTableEntity) Create() EntityDiff {
	return &CreateTableEntityDiff{createTable: &c.CreateTable}
}

// Drop implements Entity interface
func (c *CreateTableEntity) Drop() EntityDiff {
	dropTable := &sqlparser.DropTable{
		FromTables: []sqlparser.TableName{c.Table},
	}
	return &DropTableEntityDiff{from: c, dropTable: dropTable}
}

func sortAlterOptions(diff *AlterTableEntityDiff) {
	optionOrder := func(opt sqlparser.AlterOption) int {
		switch opt.(type) {
		case *sqlparser.DropKey:
			return 1
		case *sqlparser.DropColumn:
			return 2
		case *sqlparser.ModifyColumn:
			return 3
		case *sqlparser.AddColumns:
			return 4
		case *sqlparser.AddIndexDefinition:
			return 5
		case *sqlparser.AddConstraintDefinition:
			return 6
		case sqlparser.TableOptions, *sqlparser.TableOptions:
			return 7
		default:
			return math.MaxInt
		}
	}
	opts := diff.alterTable.AlterOptions
	sort.SliceStable(opts, func(i, j int) bool {
		return optionOrder(opts[i]) < optionOrder(opts[j])
	})
}

// apply attempts to apply an ALTER TABLE diff onto this entity's table definition.
// supported modifications are only those created by schemadiff's Diff() function.
func (c *CreateTableEntity) apply(diff *AlterTableEntityDiff) error {
	sortAlterOptions(diff)
	// Apply partitioning changes:
	if spec := diff.alterTable.PartitionSpec; spec != nil {
		switch {
		case spec.Action == sqlparser.RemoveAction && spec.IsAll:
			// Remove partitioning
			c.TableSpec.PartitionOption = nil
		case spec.Action == sqlparser.DropAction && len(spec.Names) > 0:
			for _, dropPartitionName := range spec.Names {
				// Drop partitions
				partitionName := dropPartitionName.String()
				if c.TableSpec.PartitionOption == nil {
					return errors.Wrap(ErrApplyPartitionNotFound, partitionName)
				}
				partitionFound := false
				for i, p := range c.TableSpec.PartitionOption.Definitions {
					if p.Name.String() == partitionName {
						c.TableSpec.PartitionOption.Definitions = append(
							c.TableSpec.PartitionOption.Definitions[0:i],
							c.TableSpec.PartitionOption.Definitions[i+1:]...,
						)
						partitionFound = true
						break
					}
				}
				if !partitionFound {
					return errors.Wrap(ErrApplyPartitionNotFound, partitionName)
				}
			}
		case spec.Action == sqlparser.AddAction && len(spec.Definitions) == 1:
			// Add one partition
			partitionName := spec.Definitions[0].Name.String()
			if c.TableSpec.PartitionOption == nil {
				return ErrApplyNoPartitions
			}
			if len(c.TableSpec.PartitionOption.Definitions) == 0 {
				return ErrApplyNoPartitions
			}
			for _, p := range c.TableSpec.PartitionOption.Definitions {
				if p.Name.String() == partitionName {
					return errors.Wrap(ErrApplyDuplicatePartition, partitionName)
				}
			}
			c.TableSpec.PartitionOption.Definitions = append(
				c.TableSpec.PartitionOption.Definitions,
				spec.Definitions[0],
			)
		default:
			return errors.Wrap(ErrUnsupportedApplyOperation, sqlparser.CanonicalString(spec))
		}
	}
	if diff.alterTable.PartitionOption != nil {
		// Specify new spec:
		c.CreateTable.TableSpec.PartitionOption = diff.alterTable.PartitionOption
	}
	// reorderColumn attempts to reorder column that is right now in position 'colIndex',
	// based on its FIRST or AFTER specs (if any)
	reorderColumn := func(colIndex int, first bool, after *sqlparser.ColName) error {
		var newCols []*sqlparser.ColumnDefinition // nil
		col := c.TableSpec.Columns[colIndex]
		switch {
		case first:
			newCols = append(newCols, col)
			newCols = append(newCols, c.TableSpec.Columns[0:colIndex]...)
			newCols = append(newCols, c.TableSpec.Columns[colIndex+1:]...)
		case after != nil:
			afterColFound := false
			// look for the AFTER column; it has to exist!
			for a, afterCol := range c.TableSpec.Columns {
				if afterCol.Name.String() == after.Name.String() {
					if colIndex < a {
						// moving column i to the right
						newCols = append(newCols, c.TableSpec.Columns[0:colIndex]...)
						newCols = append(newCols, c.TableSpec.Columns[colIndex+1:a+1]...)
						newCols = append(newCols, col)
						newCols = append(newCols, c.TableSpec.Columns[a+1:]...)
					} else {
						// moving column i to the left
						newCols = append(newCols, c.TableSpec.Columns[0:a+1]...)
						newCols = append(newCols, col)
						newCols = append(newCols, c.TableSpec.Columns[a+1:colIndex]...)
						newCols = append(newCols, c.TableSpec.Columns[colIndex+1:]...)
					}
					afterColFound = true
					break
				}
			}
			if !afterColFound {
				return errors.Wrap(ErrApplyColumnNotFound, after.Name.String())
			}
		default:
			// no change in position
		}

		if newCols != nil {
			c.TableSpec.Columns = newCols
		}
		return nil
	}

	columnExists := map[string]bool{}
	for _, col := range c.CreateTable.TableSpec.Columns {
		columnExists[col.Name.String()] = true
	}

	// apply a single AlterOption; only supported types are those generated by Diff()
	applyAlterOption := func(opt sqlparser.AlterOption) error {
		switch opt := opt.(type) {
		case *sqlparser.DropKey:
			// applies to either indexes or FK constraints
			// we expect the named key to be found
			found := false
			switch opt.Type {
			case sqlparser.NormalKeyType, sqlparser.PrimaryKeyType:
				for i, index := range c.TableSpec.Indexes {
					if index.Info.Name.String() == opt.Name.String() {
						found = true
						c.TableSpec.Indexes = append(c.TableSpec.Indexes[0:i], c.TableSpec.Indexes[i+1:]...)
						break
					}
				}
			case sqlparser.ForeignKeyType, sqlparser.CheckKeyType:
				for i, constraint := range c.TableSpec.Constraints {
					if constraint.Name.String() == opt.Name.String() {
						found = true
						c.TableSpec.Constraints = append(c.TableSpec.Constraints[0:i], c.TableSpec.Constraints[i+1:]...)
						break
					}
				}
			default:
				return errors.Wrap(ErrUnsupportedApplyOperation, sqlparser.CanonicalString(opt))
			}
			if !found {
				return errors.Wrap(ErrApplyKeyNotFound, opt.Name.String())
			}
		case *sqlparser.AddIndexDefinition:
			// validate no existing key by same name
			keyName := opt.IndexDefinition.Info.Name.String()
			for _, index := range c.TableSpec.Indexes {
				if index.Info.Name.String() == keyName {
					return errors.Wrap(ErrApplyDuplicateKey, keyName)
				}
			}
			for colName := range getKeyColumnNames(opt.IndexDefinition) {
				if !columnExists[colName] {
					return errors.Wrapf(ErrInvalidColumnInKey, "key: %v, column: %v", keyName, colName)
				}
			}
			c.TableSpec.Indexes = append(c.TableSpec.Indexes, opt.IndexDefinition)
		case *sqlparser.AddConstraintDefinition:
			// validate no existing constraint by same name
			for _, c := range c.TableSpec.Constraints {
				if c.Name.String() == opt.ConstraintDefinition.Name.String() {
					return errors.Wrap(ErrApplyDuplicateConstraint, opt.ConstraintDefinition.Name.String())
				}
			}
			c.TableSpec.Constraints = append(c.TableSpec.Constraints, opt.ConstraintDefinition)
		case *sqlparser.AlterCheck:
			// we expect the constraint to exist
			found := false
			constraintName := opt.Name.String()
			for _, constraint := range c.TableSpec.Constraints {
				checkDetails, ok := constraint.Details.(*sqlparser.CheckConstraintDefinition)
				if ok && constraint.Name.String() == constraintName {
					found = true
					checkDetails.Enforced = opt.Enforced
					break
				}
			}
			if !found {
				return errors.Wrap(ErrApplyConstraintNotFound, opt.Name.String())
			}
		case *sqlparser.DropColumn:
			// we expect the column to exist
			found := false
			colName := opt.Name.Name.String()
			for i, col := range c.TableSpec.Columns {
				if col.Name.String() == colName {
					found = true
					c.TableSpec.Columns = append(c.TableSpec.Columns[0:i], c.TableSpec.Columns[i+1:]...)
					break
				}
			}
			if !found {
				return errors.Wrap(ErrApplyColumnNotFound, opt.Name.Name.String())
			}
			delete(columnExists, colName)
		case *sqlparser.AddColumns:
			if len(opt.Columns) != 1 {
				// our Diff only ever generates a single column per AlterOption
				return errors.Wrap(ErrUnsupportedApplyOperation, sqlparser.CanonicalString(opt))
			}
			// validate no column by same name
			addedCol := opt.Columns[0]
			colName := addedCol.Name.String()
			for _, col := range c.TableSpec.Columns {
				if col.Name.String() == colName {
					return errors.Wrap(ErrApplyDuplicateColumn, addedCol.Name.String())
				}
			}
			c.TableSpec.Columns = append(c.TableSpec.Columns, addedCol)
			// see if we need to position it anywhere other than end of table
			if err := reorderColumn(len(c.TableSpec.Columns)-1, opt.First, opt.After); err != nil {
				return err
			}
			columnExists[colName] = true
		case *sqlparser.ModifyColumn:
			// we expect the column to exist
			found := false
			for i, col := range c.TableSpec.Columns {
				if col.Name.String() == opt.NewColDefinition.Name.String() {
					found = true
					// redefine. see if we need to position it anywhere other than end of table
					c.TableSpec.Columns[i] = opt.NewColDefinition
					if err := reorderColumn(i, opt.First, opt.After); err != nil {
						return err
					}
					break
				}
			}
			if !found {
				return errors.Wrap(ErrApplyColumnNotFound, opt.NewColDefinition.Name.String())
			}
		case sqlparser.TableOptions:
			// Apply table options. Options that have their DEFAULT value are actually remvoed.
			for _, option := range opt {
				func() {
					for i, existingOption := range c.TableSpec.Options {
						if option.Name == existingOption.Name {
							if isDefaultTableOptionValue(option) {
								// remove the option
								c.TableSpec.Options = append(c.TableSpec.Options[0:i], c.TableSpec.Options[i+1:]...)
							} else {
								c.TableSpec.Options[i] = option
							}
							// option found. No need for further iteration.
							return
						}
					}
					// option not found. We add it
					c.TableSpec.Options = append(c.TableSpec.Options, option)
				}()
			}
		default:
			return errors.Wrap(ErrUnsupportedApplyOperation, sqlparser.CanonicalString(opt))
		}
		return nil
	}
	for _, alterOption := range diff.alterTable.AlterOptions {
		if err := applyAlterOption(alterOption); err != nil {
			return err
		}
	}
	if err := c.postApplyNormalize(); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return err
	}
	return nil
}

// Apply attempts to apply given ALTER TABLE diff onto the table defined by this entity.
// This entity is unmodified. If successful, a new CREATE TABLE entity is returned.
func (c *CreateTableEntity) Apply(diff EntityDiff) (Entity, error) {
	dupCreateTable := &sqlparser.CreateTable{
		Temp:        c.Temp,
		Table:       c.Table,
		IfNotExists: c.IfNotExists,
		TableSpec:   nil,
		OptLike:     nil,
		Comments:    nil,
		FullyParsed: c.FullyParsed,
	}
	if c.TableSpec != nil {
		d := *c.TableSpec
		dupCreateTable.TableSpec = &d
	}
	if c.OptLike != nil {
		d := *c.OptLike
		dupCreateTable.OptLike = &d
	}
	if c.Comments != nil {
		d := *c.Comments
		dupCreateTable.Comments = &d
	}
	dup := &CreateTableEntity{CreateTable: *dupCreateTable}
	for diff != nil {
		alterDiff, ok := diff.(*AlterTableEntityDiff)
		if !ok {
			return nil, ErrEntityTypeMismatch
		}
		if !diff.IsEmpty() {
			if err := dup.apply(alterDiff); err != nil {
				return nil, err
			}
		}
		diff = diff.SubsequentDiff()
	}
	return dup, nil
}

// postApplyNormalize runs at the end of apply() and to reorganize/edit things that
// a MySQL will do implicitly:
// - edit or remove keys if referenced columns are dropped
func (c *CreateTableEntity) postApplyNormalize() error {
	// reduce or remove keys based on existing column list
	// (a column may have been removed)postApplyNormalize
	columnExists := map[string]bool{}
	for _, col := range c.CreateTable.TableSpec.Columns {
		columnExists[col.Name.String()] = true
	}
	nonEmptyIndexes := []*sqlparser.IndexDefinition{}

	keyHasNonExistentColumns := func(keyCol *sqlparser.IndexColumn) bool {
		if keyCol.Expression != nil {
			colNames := getNodeColumns(keyCol.Expression)
			for colName := range colNames {
				if !columnExists[colName] {
					// expression uses a non-existent column
					return true
				}
			}
		}
		if keyCol.Column.String() != "" {
			if !columnExists[keyCol.Column.String()] {
				return true
			}
		}
		return false
	}
	for _, key := range c.CreateTable.TableSpec.Indexes {
		existingKeyColumns := []*sqlparser.IndexColumn{}
		for _, keyCol := range key.Columns {
			if !keyHasNonExistentColumns(keyCol) {
				existingKeyColumns = append(existingKeyColumns, keyCol)
			}
		}
		if len(existingKeyColumns) > 0 {
			key.Columns = existingKeyColumns
			nonEmptyIndexes = append(nonEmptyIndexes, key)
		}
	}
	c.CreateTable.TableSpec.Indexes = nonEmptyIndexes

	return nil
}

func getNodeColumns(node sqlparser.SQLNode) (colNames map[string]bool) {
	colNames = map[string]bool{}
	if node != nil {
		_ = sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) {
			case *sqlparser.ColName:
				colNames[node.Name.String()] = true
			}
			return true, nil
		}, node)
	}
	return colNames
}

func getKeyColumnNames(key *sqlparser.IndexDefinition) (colNames map[string]bool) {
	colNames = map[string]bool{}
	for _, col := range key.Columns {
		if colName := col.Column.String(); colName != "" {
			colNames[colName] = true
		}
		for name := range getNodeColumns(col.Expression) {
			colNames[name] = true
		}
	}
	return colNames
}

// validate checks that the table structure is valid:
// - all columns referenced by keys exist
func (c *CreateTableEntity) validate() error {
	columnExists := map[string]bool{}
	for _, col := range c.CreateTable.TableSpec.Columns {
		colName := col.Name.String()
		if columnExists[colName] {
			return errors.Wrapf(ErrApplyDuplicateColumn, "column: %v", colName)
		}
		columnExists[colName] = true
	}
	// validate all columns referenced by indexes do in fact exist
	for _, key := range c.CreateTable.TableSpec.Indexes {
		for colName := range getKeyColumnNames(key) {
			if !columnExists[colName] {
				return errors.Wrapf(ErrInvalidColumnInKey, "key: %v, column: %v", key.Info.Name.String(), colName)
			}
		}
	}
	if partition := c.CreateTable.TableSpec.PartitionOption; partition != nil {
		// validate no two partitions have same name
		partitionExists := map[string]bool{}
		for _, p := range partition.Definitions {
			partitionName := p.Name.String()
			if partitionExists[partitionName] {
				return errors.Wrapf(ErrApplyDuplicatePartition, "partition: %v", partitionName)
			}
			partitionExists[partitionName] = true
		}
		// validate columns referenced by partitions do in fact exist
		// also, validate that all unique keys include partitioned columns
		partitionColNames := []string{}
		err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) {
			case *sqlparser.ColName:
				partitionColNames = append(partitionColNames, node.Name.String())
			}
			return true, nil
		}, partition.Expr)
		if err != nil {
			return err
		}

		for _, partitionColName := range partitionColNames {
			// Validate columns exists in table:
			if !columnExists[partitionColName] {
				return errors.Wrapf(ErrInvalidColumnInPartition, "column: %v", partitionColName)
			}

			// Validate all unique keys include this column:
			for _, key := range c.CreateTable.TableSpec.Indexes {
				if !key.Info.Unique {
					continue
				}
				colFound := false
				for _, col := range key.Columns {
					colName := col.Column.String()
					if colName == partitionColName {
						colFound = true
						break
					}
				}
				if !colFound {
					return errors.Wrapf(ErrMissingParitionColumnInUniqueKey, "column: %v not found in key: %v", partitionColName, key.Info.Name.String())
				}
			}
		}
	}
	return nil
}