# ClickHouse trivial example

Smoke test for `chdialect` against a local ClickHouse server.

## Run

```shell
go run .
```

## DateTimeInputBestEffort

If you use `chdialect.WithDateTimeInputFormat(chdialect.DateTimeInputBestEffort)`, bun writes
`time.Time` values as RFC3339Nano strings in SQL. ClickHouse only accepts that format when
**best_effort** parsing is enabled on the connection, for example:

```go
Settings: clickhouse.Settings(map[string]any{
    "date_time_input_format":        "best_effort",
    "cast_string_to_date_time_mode": "best_effort",
})
```

Set `date_time_input_format` and `cast_string_to_date_time_mode` to `best_effort`
(required for `INSERT ... VALUES` where string literals are cast to `DateTime` / `DateTime64`).

The default `DateTimeInputBasic` uses `YYYY-MM-DD HH:MM:SS` and works with ClickHouse defaults.

See [ClickHouse format settings](https://clickhouse.com/docs/operations/settings/formats#date_time_input_format).
