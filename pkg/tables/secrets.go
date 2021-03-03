package tables

import (
	"context"
	"strings"

	"github.com/kolide/osquery-go/plugin/table"
	log "github.com/sirupsen/logrus"

	"github.com/palestamp/ksql/pkg/kubeapi"
)

type Secrets struct {
	kc *kubeapi.KubeConfig
}

func NewSecrets(kc *kubeapi.KubeConfig) *Secrets {
	return &Secrets{kc: kc}
}
func (d *Secrets) Columns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("context"),
		table.TextColumn("namespace"),
		table.TextColumn("name"),
		table.TextColumn("key"),
		table.TextColumn("data"),
	}
}

func (d *Secrets) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	logger := log.WithField("generate", "secrets")
	logQueryContext(logger, queryContext)

	namespaces, err := listNamespaces(d.kc, queryContext)

	var rows []map[string]string
	for _, c := range namespaces {
		secrets, err := d.kc.ListSecrets(c.Context, c.Namespace)
		if err != nil {
			return nil, err
		}

		for _, s := range secrets {
			for k, d := range s.Data {
				rows = append(rows, map[string]string{
					"context":   c.Context,
					"namespace": c.Namespace,
					"name":      s.Name,
					"key":       k,
					"data":      strings.TrimSpace(string(d)),
				})
			}
		}
	}

	return rows, err
}
