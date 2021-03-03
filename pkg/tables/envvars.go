package tables

import (
	"context"
	"fmt"
	"strings"

	"github.com/kolide/osquery-go/plugin/table"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/palestamp/ksql/pkg/kubeapi"
)

type EnvVars struct {
	kc *kubeapi.KubeConfig
}

func NewEnvVars(kc *kubeapi.KubeConfig) *EnvVars {
	return &EnvVars{kc: kc}
}
func (d *EnvVars) Columns() []table.ColumnDefinition {
	return []table.ColumnDefinition{
		table.TextColumn("context"),
		table.TextColumn("namespace"),
		table.TextColumn("deployment"),
		table.TextColumn("image"),
		table.TextColumn("tag"),
		table.TextColumn("env_key"),
		table.TextColumn("env_value"),
		table.TextColumn("env_is_secret"),
		table.TextColumn("secret_name"),
		table.TextColumn("secret_key"),
	}
}

func (d *EnvVars) Generate(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	logger := log.WithField("generate", "env-vars")
	logQueryContext(logger, queryContext)

	containers, err := listContainers(d.kc, queryContext)

	var rows []map[string]string
	for _, c := range containers {
		image, tag := splitTag(c.Container.Image)

		for _, e := range c.Container.Env {
			env := getEnvVar(e)

			rows = append(rows, map[string]string{
				"context":       c.Context,
				"namespace":     c.Namespace,
				"deployment":    c.Deployment,
				"image":         image,
				"tag":           tag,
				"env_key":       env.Name,
				"env_value":     strings.TrimSpace(env.Value),
				"env_is_secret": fmt.Sprintf("%t", env.IsSecret),
				"secret_name":   env.SecretName,
				"secret_key":    env.SecretKey,
			})
		}
	}

	return rows, err
}

func listNamespaces(kc *kubeapi.KubeConfig, qc table.QueryContext) ([]NamespaceWrap, error) {
	contexts, err := kc.ListContexts()
	if err != nil {
		return nil, err
	}

	var out []NamespaceWrap
	for _, c := range filterConstraint(contexts, qc.Constraints["context"]) {
		namespaces, err := kc.ListNamespaces(c)
		if err != nil {
			return nil, err
		}

		for _, n := range filterConstraint(namespaces, qc.Constraints["namespace"]) {
			out = append(out, NamespaceWrap{
				Context:   c,
				Namespace: n,
			})
		}
	}

	return out, nil
}

func listContainers(kc *kubeapi.KubeConfig, qc table.QueryContext) ([]ContainerWrap, error) {
	namespaces, err := listNamespaces(kc, qc)
	if err != nil {
		return nil, err
	}

	var out []ContainerWrap
	for _, n := range namespaces {
		deployments, err := kc.ListDeployments(n.Context, n.Namespace)
		if err != nil {
			return nil, err
		}

		for _, d := range deployments {
			if len(filterConstraint([]string{d.Name}, qc.Constraints["deployment"])) == 0 {
				continue
			}

			for _, cn := range d.Spec.Template.Spec.Containers {
				out = append(out, ContainerWrap{
					Context:    n.Context,
					Namespace:  n.Namespace,
					Deployment: d.Name,
					Container:  cn,
				})
			}
		}

		statefulSets, err := kc.ListStatefulSets(n.Context, n.Namespace)
		if err != nil {
			return nil, err
		}

		for _, s := range statefulSets {
			if len(filterConstraint([]string{s.Name}, qc.Constraints["deployment"])) == 0 {
				continue
			}

			for _, cn := range s.Spec.Template.Spec.Containers {
				out = append(out, ContainerWrap{
					Context:    n.Context,
					Namespace:  n.Namespace,
					Deployment: s.Name,
					Container:  cn,
				})
			}
		}
	}

	return out, nil
}

func getEnvVar(env corev1.EnvVar) EnvVar {
	e := EnvVar{Name: env.Name}

	if env.Value != "" {
		e.Value = env.Value
		return e
	}

	if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
		e.IsSecret = true
		e.SecretName = env.ValueFrom.SecretKeyRef.Name
		e.SecretKey = env.ValueFrom.SecretKeyRef.Key
	}

	return e
}

type NamespaceWrap struct {
	Context   string
	Namespace string
}

type ContainerWrap struct {
	Context    string
	Namespace  string
	Deployment string
	Container  corev1.Container
}

type EnvVar struct {
	Name       string
	Value      string
	IsSecret   bool
	SecretName string
	SecretKey  string
}
