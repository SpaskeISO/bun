package chdialect

import (
	"fmt"
	"time"
)

// DateTimeInputFormat selects how [Dialect.AppendTime] formats time values in SQL.
type DateTimeInputFormat int

const (
	// DateTimeInputBasic uses YYYY-MM-DD HH:MM:SS (ClickHouse default parsing).
	DateTimeInputBasic DateTimeInputFormat = iota

	// DateTimeInputBestEffort uses RFC3339Nano in generated SQL, for example
	// 2018-06-08T01:02:03.123456789Z.
	//
	// ClickHouse accepts that format only when best_effort parsing is enabled on the
	// connection (for example date_time_input_format and cast_string_to_date_time_mode
	// set to "best_effort" in clickhouse-go Options.Settings).
	//
	// Pass [WithDateTimeInputFormat](DateTimeInputBestEffort) to [New].
	// See https://clickhouse.com/docs/operations/settings/formats#date_time_input_format
	DateTimeInputBestEffort
)

const (
	// BasicTimeFormat matches ClickHouse date_time_input_format=basic.
	BasicTimeFormat = "2006-01-02 15:04:05"
	// BestEffortTimeFormat is RFC3339Nano for [DateTimeInputBestEffort].
	BestEffortTimeFormat = time.RFC3339Nano
)

// WithDateTimeInputFormat configures [Dialect.AppendTime] output.
// See [DateTimeInputBestEffort] for required ClickHouse session settings.
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
		b = t.AppendFormat(b, BestEffortTimeFormat)
	default:
		b = t.AppendFormat(b, BasicTimeFormat)
	}
	b = append(b, '\'')
	return b
}
