package tables

import (
	"context"
	"strings"

	"github.com/kolide/osquery-go/plugin/table"
	"github.com/palestamp/ksql/pkg/kubeapi"
	log "github.com/sirupsen/logrus"
)

type Containers struct {
	kc *kubeapi.KubeConfig
}

func NewContainers(kc *kubeapi.KubeConfig) *Containers {
	return &Containers{kc: kc}
}

func (d *Containers) Columns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("context"),
		table.TextColumn("namespace"),
		table.TextColumn("deployment"),
		table.TextColumn("image"),
		table.TextColumn("tag"),
	}
}

func (d *Containers) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	logger := log.WithField("generate", "containers")
	logQueryContext(logger, queryContext)

	cs, err := listContainers(d.kc, queryContext)
	if err != nil {
		return nil, err
	}

	var rows []map[string]string
	for _, c := range cs {
		image, tag := splitTag(c.Container.Image)

		rows = append(rows, map[string]string{
			"context":    c.Context,
			"namespace":  c.Namespace,
			"deployment": c.Deployment,
			"image":      image,
			"tag":        tag,
		})
	}

	return rows, nil
}

func splitTag(s string) (string, string) {
	parts := strings.Split(s, ":")
	if len(parts) == 1 {
		return parts[0], "empty"
	}

	return parts[0], parts[1]
}
