package chdialect

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/mod/semver"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/schema"
)

func init() {
	if Version() != bun.Version() {
		panic(fmt.Errorf("chdialect and Bun must have the same version: v%s != v%s",
			Version(), bun.Version()))
	}
}

type Dialect struct {
	schema.BaseDialect

	tables              *schema.Tables
	features            feature.Feature
	loc                 *time.Location
	dateTimeInputFormat DateTimeInputFormat // default Basic
	serialIDNaming      SerialIDNaming      // default table_field
}

var _ schema.Dialect = (*Dialect)(nil)

type DialectOption func(d *Dialect)

func New(opts ...DialectOption) *Dialect {
	d := new(Dialect)
	d.tables = schema.NewTables(d)
	d.features = feature.CTE |
		feature.TableTruncate |
		feature.TableNotExists |
		feature.SelectExists |
		feature.DefaultPlaceholder |
		feature.CompositeIn |
		feature.AlterColumnExists

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func WithoutFeature(other feature.Feature) DialectOption {
	return func(d *Dialect) {
		d.features = d.features.Remove(other)
	}
}

func (d *Dialect) Init(db *sql.DB) {
	var version string
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		log.Printf("can't discover ClickHouse version: %s", err)
		return
	}

	version = "v" + cleanupVersion(version)
	if version == "v" {
		return
	}

	// Lightweight UPDATE/DELETE (patch parts) for MergeTree-family tables.
	if semver.Compare(version, "v25.7") >= 0 {
		d.features |= feature.UpdateTableAlias | feature.DeleteTableAlias
	}

	// generateSerialID support
	if semver.Compare(version, "v25.1") >= 0 {
		d.features |= feature.AutoIncrement
	}

}

// cleanupVersion extracts a semver-compatible version from VERSION(), e.g. "26.3.9.8".
func cleanupVersion(s string) string {
	s = strings.TrimSpace(s)

	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' && b.Len() > 0:
			b.WriteRune(r)
		case b.Len() > 0:
			goto done
		}
	}
done:
	s = b.String()
	if s == "" {
		return ""
	}

	// golang.org/x/mod/semver supports major.minor.patch only.
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		s = strings.Join(parts[:3], ".")
	}
	return s
}

func (d *Dialect) Name() dialect.Name {
	return dialect.ClickHouse
}

func (d *Dialect) Features() feature.Feature {
	return d.features
}

func (d *Dialect) Tables() *schema.Tables {
	return d.tables
}

func (d *Dialect) OnTable(table *schema.Table) {
	panic("not implemented") // TODO: Implement
}

func (d *Dialect) IdentQuote() byte {
	return '`'
}

func (*Dialect) AppendString(b []byte, s string) []byte {
	b = append(b, '\'')
loop:
	for _, r := range s {
		switch r {
		case '\000':
			continue loop
		case '\'':
			b = append(b, "''"...)
			continue loop
		case '\\':
			b = append(b, '\\', '\\')
			continue loop
		}

		if r < utf8.RuneSelf {
			b = append(b, byte(r))
			continue
		}

		l := len(b)
		if cap(b)-l < utf8.UTFMax {
			b = append(b, make([]byte, utf8.UTFMax)...)
		}
		n := utf8.EncodeRune(b[l:l+utf8.UTFMax], r)
		b = b[:l+n]
	}
	b = append(b, '\'')
	return b
}

func (d *Dialect) AppendBytes(b []byte, bs []byte) []byte {
	if bs == nil {
		return dialect.AppendNull(b)
	}
	b = append(b, "unhex('"...)
	s := len(b)
	hex.Encode(b[s:], bs)
	b = append(b, '\'')
	b = append(b, ')')
	return b
}

func (d *Dialect) AppendJSON(b []byte, jsonb []byte) []byte {

	b = append(b, '\'')

	for _, c := range jsonb {
		switch c {
		case '\'':
			b = append(b, "''"...)
		case '\\':
			b = append(b, `\\`...)
		default:
			b = append(b, c)
		}
	}

	b = append(b, '\'')

	return b
}

// AppendSequence adds the appropriate instruction for the driver to create a sequence
// from which (autoincremented) values for the column will be generated.
func (d *Dialect) AppendSequence(b []byte, t *schema.Table, f *schema.Field) []byte {
	var name string
	switch d.serialIDNaming {
	case SerialIDByUUID:
		name = uuid.NewString()
	default:
		name = fmt.Sprintf("%s_%s", t.Name, f.Name)
	}
	b = append(b, " DEFAULT generateSerialID('"...)
	b = append(b, name...)
	b = append(b, "')"...)
	return b
}

// DefaultVarcharLen should be returned for dialects in which specifying VARCHAR length
// is mandatory in queries that modify the schema (CREATE TABLE / ADD COLUMN, etc).
// Dialects that do not have such requirement may return 0, which should be interpreted so by the caller.
func (d *Dialect) DefaultVarcharLen() int {
	return 255
}

// DefaultSchema should returns the name of the default database schema.
func (d *Dialect) DefaultSchema() string {
	return "default"
}
