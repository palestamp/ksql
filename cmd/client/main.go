package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/kolide/osquery-go"
	"github.com/olekukonko/tablewriter"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Queries map[string]Query `json:"queries"`
}

type Query struct {
	Columns []string               `json:"columns"`
	Args    map[string]interface{} `json:"args"`
	Query   string                 `json:"query"`
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

var (
	socket  = flag.String("socket", "", "osquery socket")
	config  = flag.String("config", "config.yaml", "path to config")
	query   = flag.String("query", "", "query to execute")
	defines = flag.String("define", "", "definitions in format: arg1=value;list_arg1=val1,val2")
)

func parseDefines(s string) map[string]interface{} {
	out := make(map[string]interface{})

	if s == "" {
		return out
	}

	pairs := strings.Split(s, ";")
	for _, p := range pairs {

		tokens := strings.Split(p, "=")
		fmt.Println(tokens)
		if len(tokens) != 2 {
			panic("invalid")
		}

		if strings.Contains(tokens[1], ",") {
			out[tokens[0]] = strings.Split(tokens[1], ",")
		} else {
			out[tokens[0]] = tokens[1]
		}
	}

	return out
}

func populateArgs(into, from map[string]interface{}) map[string]interface{} {
	if into == nil {
		into = make(map[string]interface{})
	}

	for k, v := range from {
		into[k] = v
	}

	return into
}

func main() {
	flag.Parse()

	c, err := Load(*config)
	if err != nil {
		panic(err)
	}

	q, ok := c.Queries[*query]
	if !ok {
		panic("unknown query")
	}

	tmpl, err := template.New("q").Funcs(sprig.FuncMap()).Parse(q.Query)
	if err != nil {
		panic(err)
	}

	args := populateArgs(q.Args, parseDefines(*defines))

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, args)
	if err != nil {
		panic(err)
	}

	if *socket == "" {
		log.Fatalf("Usage: %s SOCKET_PATH", os.Args[0])
	}

	client, err := osquery.NewClient(*socket, 100*time.Second)
	if err != nil {
		log.Fatalf("Error creating Thrift client: %v", err)
	}
	defer client.Close()

	resp, err := client.Query(buf.String())
	if err != nil {
		log.Fatalf("Error communicating with osqueryd: %v", err)
	}

	if resp.Status.Code != 0 {
		log.Fatalf("osqueryd returned error: %s", resp.Status.Message)
	}

	var rows [][]string
	for _, v := range resp.Response {

		var row []string
		for _, h := range q.Columns {
			row = append(row, v[h])
		}

		rows = append(rows, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader(q.Columns)
	table.AppendBulk(rows)
	table.Render()
}
