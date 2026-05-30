package chdialect

import (
	"fmt"
	"time"
)

type DateTimeInputFormat int

const (
	DateTimeInputBasic DateTimeInputFormat = iota
	DateTimeInputBestEffort
)

const (
	BasicTimeFormat = "2006-01-02 15:04:05.999999"
)

func WithDateTimeInputFormat(f DateTimeInputFormat) DialectOption {
	return func(d *Dialect) { d.dateTimeInputFormat = f }
}

func WithTimeLocation(loc string) DialectOption {
	return func(d *Dialect) {
		location, err := time.LoadLocation(loc)
		if err != nil {
			panic(fmt.Errorf("chdialect can't load provided location %s: %s", loc, err))
		}
		d.loc = location
	}
}

func (d *Dialect) AppendTime(b []byte, tm time.Time) []byte {
	t := tm.UTC()
	if d.loc != nil {
		t = tm.In(d.loc)
	}
	b = append(b, '\'')
	switch d.dateTimeInputFormat {
	case DateTimeInputBestEffort:
		b = t.AppendFormat(b, time.RFC3339Nano) // ISO, UTC
	default:
		b = t.AppendFormat(b, BasicTimeFormat)
	}
	b = append(b, '\'')
	return b
}
