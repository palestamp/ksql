package kubeapi

import (
	"sort"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/palestamp/ksql/pkg/kubeconfig"
)

func NewKubeConfig(ignoredContexts []string) *KubeConfig {
	ignore := make(map[string]struct{})
	for _, k := range ignoredContexts {
		ignore[k] = struct{}{}
	}

	return &KubeConfig{ignoredContexts: ignore}
}

type KubeConfig struct {
	ignoredContexts map[string]struct{}
}

func (c *KubeConfig) ListContexts() ([]string, error) {
	kc := kubeconfig.New(kubeconfig.DefaultLoader)
	defer kc.Close()

	if err := kc.Parse(); err != nil {
		return nil, err
	}

	ctxs := kc.ContextNames()
	sort.Strings(ctxs)

	nc := make([]string, 0)
	for _, ctx := range ctxs {
		if _, ok := c.ignoredContexts[ctx]; ok {
			continue
		}

		nc = append(nc, ctx)
	}

	return nc, nil
}

var namespacesCache = make(map[string][]string)

func (c *KubeConfig) ListNamespaces(context string) ([]string, error) {
	logger := log.
		WithField("resource", "namespaces").
		WithField("context", context)

	if ns, ok := namespacesCache[context]; ok {
		logger.Info("Cache hit")
		return ns, nil
	}

	cs, err := c.getClientset(context)
	if err != nil {
		return nil, err
	}

	logger.Info("Requesting API")
	resp, err := cs.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0)
	for _, namespace := range resp.Items {
		namespaces = append(namespaces, namespace.GetName())
	}

	namespacesCache[context] = namespaces

	return namespaces, nil
}

type k struct {
	c string
	n string
}

var deploymentsCache = make(map[k][]appsv1.Deployment)

func (c *KubeConfig) ListDeployments(context, namespace string) ([]appsv1.Deployment, error) {
	logger := log.
		WithField("resource", "deployments").
		WithField("context", context).
		WithField("namespace", namespace)

	key := k{context, namespace}
	if ds, ok := deploymentsCache[key]; ok {
		logger.Info("Cache hit")
		return ds, nil
	}

	cs, err := c.getClientset(context)
	if err != nil {
		return nil, err
	}

	logger.Info("Requesting API")
	resp, err := cs.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deploymentsCache[key] = resp.Items

	return resp.Items, nil
}

var statefulsetCache = make(map[k][]appsv1.StatefulSet)

func (c *KubeConfig) ListStatefulSets(context, namespace string) ([]appsv1.StatefulSet, error) {
	logger := log.
		WithField("resource", "statefulsets").
		WithField("context", context).
		WithField("namespace", namespace)

	key := k{context, namespace}
	if ds, ok := statefulsetCache[key]; ok {
		logger.Info("Cache hit")
		return ds, nil
	}

	cs, err := c.getClientset(context)
	if err != nil {
		return nil, err
	}

	logger.Info("Requesting API")
	resp, err := cs.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	statefulsetCache[key] = resp.Items

	return resp.Items, nil
}

var secretsCache = make(map[k][]corev1.Secret)

func (c *KubeConfig) ListSecrets(k8sContext, namespace string) ([]corev1.Secret, error) {
	logger := log.
		WithField("resource", "secrets").
		WithField("context", k8sContext).
		WithField("namespace", namespace)

	key := k{k8sContext, namespace}
	if ds, ok := secretsCache[key]; ok {
		logger.Info("Cache hit")
		return ds, nil
	}

	cs, err := c.getClientset(k8sContext)
	if err != nil {
		return nil, err
	}

	logger.Info("Requesting API")
	secrets, err := cs.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	secretsCache[key] = secrets.Items

	return secrets.Items, nil
}

var clientsetCache = make(map[string]*kubernetes.Clientset)

func (kc *KubeConfig) getClientset(context string) (*kubernetes.Clientset, error) {
	if cs, ok := clientsetCache[context]; ok {
		return cs, nil
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context},
	).ClientConfig()

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	clientsetCache[context] = cs

	return cs, nil
}
