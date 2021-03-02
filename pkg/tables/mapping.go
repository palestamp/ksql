package tables

import (
	"context"
	"errors"
	"fmt"

	"github.com/kolide/osquery-go/plugin/table"
)

type DynamicTable struct {
	columns []table.ColumnDefinition
	rows    []map[string]string
}

func NewDynamicFromMap(m []map[string]interface{}) (*DynamicTable, error) {
	if len(m) == 0 {
		return nil, errors.New("empty data for dynamic table")
	}

	columns := make([]table.ColumnDefinition, 0)
	var rows []map[string]string

	for i, row := range m {

		mr := make(map[string]string)
		for k, v := range row {

			if i == 0 {
				columns = append(columns, table.TextColumn(k))
			}

			w, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("only text values are supported for dynamic table: %s %v", k, v)
			}

			mr[k] = w
		}

		rows = append(rows, mr)
	}

	return &DynamicTable{columns: columns, rows: rows}, nil
}

func (s *DynamicTable) Columns() []table.ColumnDefinition {
	return s.columns
}

func (d *DynamicTable) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	return d.rows, nil
}
