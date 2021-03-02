package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kolide/osquery-go"
	"github.com/kolide/osquery-go/plugin/table"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/palestamp/ksql/pkg/kubeapi"
	"github.com/palestamp/ksql/pkg/tables"
)

var (
	socket   = flag.String("socket", "", "Path to the extensions UNIX domain socket")
	timeout  = flag.Int("timeout", 3, "Seconds to wait for autoloaded extensions")
	interval = flag.Int("interval", 3, "Seconds delay between connectivity checks")
)

func setupLogger() {
	logrus.SetLevel(logrus.ErrorLevel)

	if envLvl, ok := os.LookupEnv("KSQL_LOG_LEVEL"); ok {
		lvl, err := logrus.ParseLevel(envLvl)
		if err != nil {
			return
		}

		logrus.SetLevel(lvl)
	}
}

func main() {
	flag.Parse()
	setupLogger()

	if *socket == "" {
		log.Fatalln("Missing required --socket argument")
	}

	serverTimeout := osquery.ServerTimeout(
		time.Second * time.Duration(*timeout),
	)

	serverPingInterval := osquery.ServerPingInterval(
		time.Second * time.Duration(*interval),
	)

	server, err := osquery.NewExtensionManagerServer(
		"k8s_extension",
		*socket,
		serverTimeout,
		serverPingInterval,
	)
	if err != nil {
		log.Fatalf("Error creating extension: %s\n", err)
	}

	c, err := Load("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %s\n", err)
	}

	for name, m := range c.Mappings {
		dm, err := tables.NewDynamicFromMap(m)
		if err != nil {
			log.Fatalf("Error loading mapping: %s - %s\n", name, err)
		}

		server.RegisterPlugin(NewPlugin(name, dm))
	}

	kc := &kubeapi.KubeConfig{}
	server.RegisterPlugin(
		NewPlugin("k8s_contexts", tables.NewContexts(kc)),
		NewPlugin("k8s_namespaces", tables.NewNamespaces(kc)),
		NewPlugin("k8s_containers", tables.NewContainers(kc)),
		NewPlugin("k8s_env_vars", tables.NewEnvVars(kc)),
		// NewPlugin("k8s_deployment_synonyms", tables.NewDeploymentSynonyms()),
	)

	log.Info("Starting server")
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func NewPlugin(name string, tbl tables.Table) *table.Plugin {
	return table.NewPlugin(name, tbl.Columns(), tbl.Generate)
}

type Config struct {
	Mappings map[string][]map[string]interface{} `json:"mappings"`
}

func Load(configFilepath string) (Config, error) {
	b, err := ioutil.ReadFile(configFilepath)
	if err != nil {
		return Config{}, fmt.Errorf("unable to load macro manifest file='%s': %v", configFilepath, err)
	}

	var c Config
	err = yaml.Unmarshal(b, &c)
	return c, err
}
