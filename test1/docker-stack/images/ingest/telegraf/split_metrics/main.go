package main

// This is documented in the readme.md file in the same directory which
// outlines the purpose and how it operates.


import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	influxSerializer "github.com/influxdata/telegraf/plugins/serializers/influx"
)

// splitMetrics extracts the prefix of each field key and adds it as a tag
func splitMetrics(newTag string, input io.Reader, output io.Writer, errorOutput io.Writer) {
	parser := influx.NewStreamParser(input)
	serializer := influxSerializer.Serializer{}
	if err := serializer.Init(); err != nil {
		fmt.Fprintf(errorOutput, "serializer init failed: %v\n", err)
		os.Exit(1)
	}

	for {
		metric, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				return // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				fmt.Fprintf(errorOutput, "parse ERR %v\n", parseErr)
				os.Exit(1)
			}
			fmt.Fprintf(errorOutput, "ERR %v\n", err)
			os.Exit(1)
		}


		newMetrics := make(map[string]telegraf.Metric)
		oldFields := []string{}
		for _, f := range metric.FieldList() {
			oldFields = append(oldFields, f.Key)
		}

		for _, f := range metric.FieldList() {
			res := strings.Split(f.Key, "|")
			val := res[0]
			field := res[1]
			if _, ok := newMetrics[val]; !ok {
				newMetrics[val] = metric.Copy()
				newMetrics[val].AddTag(newTag, val)
				for _, f := range oldFields {
					newMetrics[val].RemoveField(f)
				}
			}
			newMetrics[val].AddField(field, f.Value)
		}


		for _, m := range newMetrics {
			b, err := serializer.Serialize(m)
			if err != nil {
				fmt.Fprintf(errorOutput, "ERR %v\n", err)
				continue
			}
			fmt.Fprint(output, string(b))
		}
	}
}

// main function handles I/O and calls the core logic function
func main() {
	newTag := "dn"
	if len(os.Args) > 1 {
		newTag = os.Args[1]
	}
	splitMetrics(newTag, os.Stdin, os.Stdout, os.Stderr)
}
