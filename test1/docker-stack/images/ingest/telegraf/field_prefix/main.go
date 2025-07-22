package main

// This is documented in the readme.md file in the same directory which
// outlines the purpose and how it operates.

import (
	"fmt"
	"io"
	"os"

	"github.com/influxdata/telegraf/plugins/parsers/influx"
	influxSerializer "github.com/influxdata/telegraf/plugins/serializers/influx"
)

// prefixMetricFields modifies the metric fields by prefixing them with the uniqueField value.
func prefixMetricFields(uniqueField string, input io.Reader, output io.Writer, errorOutput io.Writer) {
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

		field, found := metric.GetField(uniqueField)
		if !found {
			fmt.Fprintf(errorOutput, "metric has no field %s: %v\n", uniqueField, metric)
			continue
		}

		// Use a new list to store field changes
		newFields := make(map[string]interface{})
		oldFields := []string{}

		for _, f := range metric.FieldList() {
			newKey := field.(string) + "|" + f.Key
			newFields[newKey] = f.Value
			oldFields = append(oldFields, f.Key)
		}

		// Add new fields after the loop
		for newKey, value := range newFields {
			metric.AddField(newKey, value)
		}

		for _, f := range oldFields {
			metric.RemoveField(f)
		}

		b, err := serializer.Serialize(metric)
		if err != nil {
			fmt.Fprintf(errorOutput, "ERR %v\n", err)
			continue
		}
		fmt.Fprint(output, string(b))
	}
}

// main function handles I/O and calls the core logic function
func main() {
	uniqueField := "dn"
	if len(os.Args) > 1 {
		uniqueField = os.Args[1]
	}
	prefixMetricFields(uniqueField, os.Stdin, os.Stdout, os.Stderr)
}
