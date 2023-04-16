package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func SaveToCSV(domain string, records []string) error {
	filename := fmt.Sprintf("%s.csv", domain)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Record Type", "Value"})

	for _, record := range records {
		recordParts := strings.SplitN(record, " ", 2)
		if len(recordParts) == 2 {
			writer.Write(recordParts)
		}
	}

	return nil
}
