package chdialect

type SerialIDNaming int

const (
	SerialIDByTableColumn SerialIDNaming = iota // table_field
	SerialIDByUUID                              // uuid.NewString()
)

func WithSerialIDNaming(n SerialIDNaming) DialectOption {
	return func(d *Dialect) { d.serialIDNaming = n }
}
