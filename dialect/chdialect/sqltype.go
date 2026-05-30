package chdialect

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/uptrace/bun/dialect/sqltype"
	"github.com/uptrace/bun/internal/tagparser"
	"github.com/uptrace/bun/schema"
)

var (
	timeType           = reflect.TypeFor[time.Time]()
	jsonRawMessageType = reflect.TypeFor[json.RawMessage]()
	nullStringType     = reflect.TypeFor[sql.NullString]()
)

const (
	chTypeInt8    = "Int8"
	chTypeInt16   = "Int16"
	chTypeInt32   = "Int32"
	chTypeInt64   = "Int64"
	chTypeInt128  = "Int128"
	chTypeInt256  = "Int256"
	chTypeUInt8   = "UInt8"
	chTypeUInt16  = "UInt16"
	chTypeUInt32  = "UInt32"
	chTypeUInt64  = "UInt64"
	chTypeUInt128 = "UInt128"
	chTypeUInt256 = "UInt256"

	chTypeFloat32 = "Float32"
	chTypeFloat64 = "Float64"

	chTypeDate       = "Date"
	chTypeDate32     = "Date32"
	chTypeTime       = "Time"
	chTypeTime64     = "Time64"
	chTypeDateTime   = "DateTime"
	chTypeDateTime64 = "DateTime64"

	chTypeEnum           = "Enum"
	chTypeUUID           = "UUID"
	chTypeIPv4           = "IPv4"
	chTypeIPv6           = "IPv6"
	chTypeArray          = "Array"
	chTypeBool           = "Bool"
	chTypeTuple          = "Tuple"
	chTypeMap            = "Map"
	chTypeVariant        = "Variant"
	chTypeLowCardinality = "LowCardinality"
	chTypeNullable       = "Nullable"
	chTypeNested         = "Nested"
	chTypeDynamic        = "Dynamic"
	chTypeJSON           = "JSON"

	chTypeDecimal    = "Decimal"
	chTypeDecimal32  = "Decimal32"
	chTypeDecimal64  = "Decimal64"
	chTypeDecimal128 = "Decimal128"
	chTypeDecimal256 = "Decimal256"

	chTypeString      = "String"
	chTypeFixedString = "FixedString"
)

func (d *Dialect) fieldSQLType(field *schema.Field) string {
	if field.UserSQLType != "" {
		return field.UserSQLType
	}

	typ := field.IndirectType

	if field.Tag.HasOption("array") {
		switch field.IndirectType.Kind() {
		case reflect.Array, reflect.Slice:
			elem := typ.Elem()
			// []byte is BLOB in Bun → store as String, not Array(UInt8).
			if elem.Kind() == reflect.Uint8 && typ.Kind() == reflect.Slice {
				return chTypeString
			}
			if elem.Kind() == reflect.Struct && !isCHWellKnownStruct(elem) {
				return chTypeArray + "(" + d.chTupleType(elem) + ")"
			}
			return chTypeArray + "(" + d.chSQLType(elem) + ")"
		}
	}

	if field.Tag.HasOption("map") {
		if typ.Kind() == reflect.Map {
			keyCH := d.chSQLType(typ.Key())
			valCH := d.chSQLType(typ.Elem())
			if keyCH != "" && valCH != "" {
				return chTypeMap + "(" + keyCH + ", " + valCH + ")"
			}
		}
	}

	if field.Tag.HasOption("tuple") {
		if typ.Kind() == reflect.Struct && !isCHWellKnownStruct(typ) {
			return d.chTupleType(typ)
		}
	}
	return d.chSQLType(typ)

}

func (d *Dialect) chSQLType(typ reflect.Type) string {
	if typ.Kind() == reflect.Pointer {
		inner := d.chSQLType(typ.Elem())
		if inner == "" {
			return ""
		}
		return chTypeNullable + "(" + inner + ")"
	}

	switch typ {
	case timeType:
		return chTypeDateTime
	case jsonRawMessageType:
		return d.jsonSQLType
	case nullStringType:
		return chTypeString
	}

	switch typ.Kind() {
	case reflect.Bool:
		return chTypeBool
	case reflect.Int8:
		return chTypeInt8
	case reflect.Int16:
		return chTypeInt16
	case reflect.Int32:
		return chTypeInt32
	case reflect.Int64:
		return chTypeInt64
	case reflect.Int:
		return chTypeInt64
	case reflect.Uint8:
		return chTypeUInt8
	case reflect.Uint16:
		return chTypeUInt16
	case reflect.Uint32:
		return chTypeUInt32
	case reflect.Uint64:
		return chTypeUInt64
	case reflect.Uint, reflect.Uintptr:
		return chTypeUInt64
	case reflect.Float32:
		return chTypeFloat32
	case reflect.Float64:
		return chTypeFloat64
	case reflect.String:
		return chTypeString
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return chTypeString
		}
		// Untagged []T: keep as String/JSON column unless user uses ,array or ,type:...
		return chTypeString
	case reflect.Map, reflect.Struct:
		// Untagged map/struct: JSON/String column unless ,map / ,tuple / ,type:...
		return d.jsonSQLType
	default:
		return chScalarFromDiscover(schema.DiscoverSQLType(typ))
	}

}

func (d *Dialect) chTupleType(typ reflect.Type) string {
	var parts []string
	for i := range typ.NumField() {
		selectedField := typ.Field(i)
		if !selectedField.IsExported() {
			continue
		}
		name := selectedField.Name
		if tag := selectedField.Tag.Get("bun"); tag != "" {
			if parsedName, ok := parseBunName(tag); ok && parsedName != "" {
				name = parsedName
			}
		}
		inner := d.chSQLType(selectedField.Type)
		if inner == "" {
			continue
		}
		parts = append(parts, name+" "+inner)
	}
	if len(parts) == 0 {
		return chTypeString
	}
	return chTypeTuple + "(" + strings.Join(parts, ",") + ")"
}

func chScalarFromDiscover(s string) string {
	switch s {
	case sqltype.Boolean:
		return chTypeBool
	case sqltype.SmallInt:
		return chTypeInt16
	case sqltype.Integer:
		return chTypeInt32
	case sqltype.BigInt:
		return chTypeInt64
	case sqltype.Real:
		return chTypeFloat32
	case sqltype.DoublePrecision:
		return chTypeFloat64
	case sqltype.VarChar, sqltype.Blob:
		return chTypeString
	case sqltype.Timestamp:
		return chTypeDateTime
	case sqltype.JSON, sqltype.JSONB:
		return chTypeJSON
	default:
		return s
	}
}

func isCHWellKnownStruct(typ reflect.Type) bool {
	return typ == timeType || typ == nullStringType
}

func parseBunName(tag string) (string, bool) {
	parsedTag := tagparser.Parse(tag)
	return parsedTag.Name, parsedTag.IsZero()
}
