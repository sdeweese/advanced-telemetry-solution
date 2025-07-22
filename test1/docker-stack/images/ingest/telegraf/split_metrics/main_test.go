package main

import (
	"bytes"
	"sort"
	"strings"
	"testing"
)

func TestSplitMetrics(t *testing.T) {
	// Input data simulating Telegraf metrics
	input := "cpu,host=server01 prefix1|dn=\"prefix1\",prefix1|descr=\"\",prefix1|usage_idle=99,prefix1|usage_user=1,prefix2|dn=\"prefix2\",prefix2|descr=\"hi\",prefix2|usage_idle=99,prefix2|usage_user=2 1626357742000000000\n"

	// Expected output, with prefixed field names
	expectedOutput := "cpu,dn=prefix1,host=server01 descr=\"\",dn=\"prefix1\",usage_idle=99,usage_user=1 1626357742000000000\n" +
		"cpu,dn=prefix2,host=server01 descr=\"hi\",dn=\"prefix2\",usage_idle=99,usage_user=2 1626357742000000000"

	// Set up a reader to simulate stdin
	stdin := bytes.NewBufferString(input)

	// Set up buffers to capture stdout and stderr
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Call the refactored function with the simulated input/output
	splitMetrics("dn", stdin, &stdout, &stderr)

	// Check the output
	final := sortLinesAndFields(stdout.String())
	if final != expectedOutput {
		t.Errorf("Unexpected output.\nExpected: %q\nGot: %q", expectedOutput, final)
	}

	// Check that there is no error output
	if stderr.Len() > 0 {
		t.Errorf("Unexpected errors:\n%s", stderr.String())
	}
}

func sortLinesAndFields(input string) string {
	// Step 1: Split the input string by new lines
	lines := strings.Split(input, "\n")

	// Step 2: Iterate over each line to process it
	for i, line := range lines {
		// Split each line by spaces
		parts := strings.Split(line, " ")

		if len(parts) < 2 {
			continue // skip if there are not enough parts
		}

		// Step 3: Take the second value and split by commas
		tags := strings.Split(parts[0], ",")
		fields := strings.Split(parts[1], ",")

		// Step 4: Sort the fields
		sort.Strings(tags)
		sort.Strings(fields)

		// Step 5: Reconstruct the sorted line
		parts[0] = strings.Join(tags, ",")
		parts[1] = strings.Join(fields, ",")
		lines[i] = strings.Join(parts, " ")
	}

	// Step 6: Sort the lines based on their content
	sort.Strings(lines)

	// Join the lines back together with new lines
	return strings.TrimLeft(strings.Join(lines, "\n"), "\n")
}
