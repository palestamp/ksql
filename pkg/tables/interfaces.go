package tables

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kolide/osquery-go/plugin/table"
	"github.com/sirupsen/logrus"
)

type Table interface {
	Columns() []table.ColumnDefinition
	Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error)
}

func logQueryContext(logger *logrus.Entry, qc table.QueryContext) {
	var clauses []string
	for name, column := range qc.Constraints {
		for _, constraint := range column.Constraints {
			clauses = append(clauses, fmt.Sprintf("%s(%s) %s '%s'", name, column.Affinity, stringifyConstraint(constraint.Operator), constraint.Expression))
		}
	}

	logger.Infof("query: %s", strings.Join(clauses, " and "))
}

func stringifyConstraint(op table.Operator) string {
	switch op {
	case table.OperatorEquals:
		return "="
	case table.Operator(68):
		return "!="
	case table.OperatorGreaterThan:
		return ">"
	case table.OperatorLessThanOrEquals:
		return "<="
	case table.OperatorLessThan:
		return "<"
	case table.OperatorGreaterThanOrEquals:
		return ">="
	case table.OperatorMatch:
		return "match"
	case table.OperatorLike:
		return "like"
	case table.OperatorGlob:
		return "glob"
	case table.OperatorRegexp:
		return "regexp"
	case table.OperatorUnique:
		return "uniq"
	}

	return strconv.Itoa(int(op))
}

func filterConstraint(in []string, constraints table.ConstraintList) []string {

	if len(constraints.Constraints) != 1 {
		return in
	}

	c := constraints.Constraints[0]

	var out []string

	switch c.Operator {
	case table.OperatorEquals:

		for _, el := range in {
			if el == c.Expression {
				out = append(out, el)
			}
		}

	case table.Operator(68):

		for _, el := range in {
			if el != c.Expression {
				out = append(out, el)
			}
		}

	default:
		return in
	}

	return out
}
