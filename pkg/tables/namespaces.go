package tables

import (
	"context"

	"github.com/kolide/osquery-go/plugin/table"
	log "github.com/sirupsen/logrus"

	"github.com/palestamp/ksql/pkg/kubeapi"
)

type Namespaces struct {
	kc *kubeapi.KubeConfig
}

func NewNamespaces(kc *kubeapi.KubeConfig) *Namespaces {
	return &Namespaces{kc: kc}
}
func (d *Namespaces) Columns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("context"),
		table.TextColumn("name"),
	}
}

func (d *Namespaces) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	logger := log.WithField("generate", "namespaces")
	logQueryContext(logger, queryContext)

	contexts, err := d.kc.ListContexts()
	if err != nil {
		return nil, err
	}

	var rows []map[string]string
	for _, c := range filterConstraint(contexts, queryContext.Constraints["context"]) {

		namespaces, err := d.kc.ListNamespaces(c)
		if err != nil {
			return nil, err
		}

		for _, n := range namespaces {
			rows = append(rows, map[string]string{
				"context": c,
				"name":    n,
			})
		}
	}

	return rows, nil
}
