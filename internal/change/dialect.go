package change

import (
	"strconv"

	"tablepro/internal/driver"
)

type Dialect struct {
	ParamMarker     func(index int) string
	QuoteIdentifier func(name string) string
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func GetDialect(dbType driver.DatabaseType) *Dialect {
	switch dbType {
	case driver.DatabaseTypePostgreSQL:
		return &Dialect{
			ParamMarker:     func(i int) string { return "$" + itoa(i) },
			QuoteIdentifier: func(n string) string { return `"` + n + `"` },
		}
	case driver.DatabaseTypeMySQL:
		return &Dialect{
			ParamMarker:     func(_ int) string { return "?" },
			QuoteIdentifier: func(n string) string { return "`" + n + "`" },
		}
	case driver.DatabaseTypeSQLite, driver.DatabaseTypeDuckDB:
		return &Dialect{
			ParamMarker:     func(_ int) string { return "?" },
			QuoteIdentifier: func(n string) string { return `"` + n + `"` },
		}
	case driver.DatabaseTypeMSSQL:
		return &Dialect{
			ParamMarker:     func(i int) string { return "@p" + itoa(i) },
			QuoteIdentifier: func(n string) string { return "[" + n + "]" },
		}
	case driver.DatabaseTypeClickHouse:
		return &Dialect{
			ParamMarker:     func(_ int) string { return "?" },
			QuoteIdentifier: func(n string) string { return `"` + n + `"` },
		}
	default:
		return &Dialect{
			ParamMarker:     func(_ int) string { return "?" },
			QuoteIdentifier: func(n string) string { return `"` + n + `"` },
		}
	}
}
