package change

import (
	"fmt"
	"strings"

	"tablepro/internal/driver"
)

type Dialect struct {
	DatabaseType   driver.DatabaseType
	QuoteChar      string
	ParamStyle     string
	ParamPrefix    string
	SupportsSchema bool
}

func GetDialect(dbType driver.DatabaseType) *Dialect {
	switch dbType {
	case driver.DatabaseTypePostgreSQL:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      `"`,
			ParamStyle:     "dollar",
			ParamPrefix:    "$",
			SupportsSchema: true,
		}
	case driver.DatabaseTypeMySQL:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      "`",
			ParamStyle:     "question",
			ParamPrefix:    "?",
			SupportsSchema: false,
		}
	case driver.DatabaseTypeSQLite:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      `"`,
			ParamStyle:     "question",
			ParamPrefix:    "?",
			SupportsSchema: false,
		}
	case driver.DatabaseTypeDuckDB:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      `"`,
			ParamStyle:     "dollar",
			ParamPrefix:    "$",
			SupportsSchema: true,
		}
	case driver.DatabaseTypeMSSQL:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      "[",
			ParamStyle:     "at",
			ParamPrefix:    "@p",
			SupportsSchema: true,
		}
	case driver.DatabaseTypeClickHouse:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      `"`,
			ParamStyle:     "question",
			ParamPrefix:    "?",
			SupportsSchema: true,
		}
	default:
		return &Dialect{
			DatabaseType:   dbType,
			QuoteChar:      `"`,
			ParamStyle:     "question",
			ParamPrefix:    "?",
			SupportsSchema: false,
		}
	}
}

func (d *Dialect) QuoteIdentifier(name string) string {
	if d.QuoteChar == "[" {
		return "[" + name + "]"
	}
	return d.QuoteChar + name + d.QuoteChar
}

func (d *Dialect) ParamMarker(index int) string {
	switch d.ParamStyle {
	case "dollar":
		return fmt.Sprintf("$%d", index)
	case "at":
		return fmt.Sprintf("@p%d", index)
	default:
		return "?"
	}
}

func (d *Dialect) QualifiedTableName(schema, table string) string {
	if d.SupportsSchema && schema != "" {
		return d.QuoteIdentifier(schema) + "." + d.QuoteIdentifier(table)
	}
	return d.QuoteIdentifier(table)
}

func (d *Dialect) BuildWhereClause(primaryKeys []string, startIndex int) (string, int) {
	var parts []string
	idx := startIndex
	for _, pk := range primaryKeys {
		parts = append(parts, fmt.Sprintf("%s = %s", d.QuoteIdentifier(pk), d.ParamMarker(idx)))
		idx++
	}
	return strings.Join(parts, " AND "), idx
}
