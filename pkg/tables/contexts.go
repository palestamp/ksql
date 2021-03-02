package tables

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/kolide/osquery-go/plugin/table"
	"github.com/palestamp/ksql/pkg/kubeapi"
)

func NewContexts(kc *kubeapi.KubeConfig) *Contexts {
	return &Contexts{kc: kc}
}

type Contexts struct {
	kc *kubeapi.KubeConfig
}

func (d *Contexts) Columns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("name"),
	}
}

func (d *Contexts) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	logger := log.WithField("generate", "contexts")
	logQueryContext(logger, queryContext)

	contexts, err := d.kc.ListContexts()
	if err != nil {
		return nil, err
	}

	var rows []map[string]string
	for _, c := range filterConstraint(contexts, queryContext.Constraints["name"]) {
		rows = append(rows, map[string]string{
			"name": c,
		})
	}

	return rows, nil
}
