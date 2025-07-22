package main

// This is likely an artifact which can be safely removed
// TODO: delete this file from repository

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

type Device struct {
	Device string `json:"device"`
	Port   string `json:"port"`
	IfType string `json:"ifType"`
	IfAlias string `json:"ifAlias"`
}

type JsonData struct {
	Current  int      `json:"current"`
	RowCount int      `json:"rowCount"`
	Rows     []Device `json:"rows"`
}

type DeviceInfo struct {
	Device  string `csv:"device"`
	Detail1 string `csv:"detail1"`
	Detail2 string `csv:"detail2"`
	Detail3 string `csv:"detail3"`
	Port    string `csv:"port"`
	IfType  string `csv:"ifType"`
	IfAlias string `csv:"ifAlias"`
}

func main() {
	// Load JSON data from file
	file, err := ioutil.ReadFile("switch_data_raw.json")
	if err != nil {
		log.Fatal(err)
	}

	var jsonData JsonData
	if err := json.Unmarshal(file, &jsonData); err != nil {
		log.Fatal(err)
	}

	// Regex pattern
	pattern := `.*list-large\\\\'>(.*?)<\\\/span>.*<br \\\/>(.*?) - (.*?) - (.*?)<br \\\/>.*`
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(err)
	}

	// Open a CSV file for writing
	csvFile, err := os.Create("results.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"device", "detail1", "detail2", "detail3", "port", "ifType", "ifAlias"}
	if err := writer.Write(headers); err != nil {
		log.Fatal(err)
	}

	// Process each row in the JSON data
	for _, row := range jsonData.Rows {
		// Apply regex to extract data
		matches := re.FindStringSubmatch(row.Device)
		if len(matches) == 5 { // ensure all capturing groups are present
			record := DeviceInfo{
				Device:  matches[1],
				Detail1: matches[2],
				Detail2: matches[3],
				Detail3: matches[4],
				Port:    row.Port,
				IfType:  row.IfType,
				IfAlias: row.IfAlias,
			}
			recordLine := []string{record.Device, record.Detail1, record.Detail2, record.Detail3, record.Port, record.IfType, record.IfAlias}
			if err := writer.Write(recordLine); err != nil {
				log.Fatal(err)
			}
		}
	}

	fmt.Println("Data has been written to results.csv")
}
