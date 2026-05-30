package chdialect

import (
	"github.com/uptrace/bun/dialect/sqltype"
	"github.com/uptrace/bun/schema"
)

func (d *Dialect) fieldSQLType(field *schema.Field) string {
	switch field.DiscoveredSQLType {
	case sqltype.Boolean:
		return "Bool"
	case sqltype.SmallInt:
		return "Int16"
	case sqltype.Integer:
		return "Int32"
	case sqltype.BigInt:
		return "Int64"
	case sqltype.Real:
		return "Float32"
	case sqltype.DoublePrecision:
		return "Float64"
	case sqltype.VarChar, sqltype.Blob:
		return "String"
	case sqltype.Timestamp:
		return "DateTime"
	case sqltype.JSON, sqltype.JSONB:
		return d.jsonSQLType
	default:
		return field.DiscoveredSQLType
	}
}
