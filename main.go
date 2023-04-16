package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

func saveToCSV(domain string, records []string) error {
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

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		color.Cyan("Enter domain names separated by commas, or type 'exit' to quit (default: erikbjerke.com):")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "exit" {
			break
		}

		domains := strings.Split(input, ",")
		if len(input) == 0 {
			domains = []string{"erikbjerke.com"}
		}

		saveOptions := make(map[string]bool)
		yesToAll := false
		none := false
		for _, domain := range domains {
			d := strings.TrimSpace(domain)
			if !yesToAll && !none {
				color.Cyan("Do you want to save the results for %s to a CSV file? (y(es)/n(o)/a(ll)/x(none))", d)
				saveOption, _ := reader.ReadString('\n')
				saveOption = strings.TrimSpace(strings.ToLower(saveOption))
				if saveOption == "a" {
					yesToAll = true
				} else if saveOption == "x" {
					none = true
				}
				saveOptions[d] = saveOption == "y" || saveOption == "a"
			} else if yesToAll {
				saveOptions[d] = true
			} else {
				saveOptions[d] = false
			}
		}

		var wg sync.WaitGroup
		ch := make(chan string)

		for _, domain := range domains {
			wg.Add(1)
			go func(d string) {
				defer wg.Done()
				records, err := FetchAllRecords(d)
				if err != nil {
					color.Red("Error fetching records for %s: %v", d, err)
					return
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Record Type", "Value"})
				table.SetAutoWrapText(false)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetRowLine(true)
				table.SetHeaderColor(
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
				)
				table.SetColumnColor(
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiYellowColor},
					tablewriter.Colors{tablewriter.FgHiWhiteColor},
				)
				table.SetRowSeparator("-")
				table.SetCenterSeparator("|")

				for _, record := range records {
					recordParts := strings.SplitN(record, " ", 2)
					if len(recordParts) == 2 {
						table.Append([]string{recordParts[0], recordParts[1]})
					}
				}

				ch <- d
				color.Cyan("DNS records for %s:", d)
				table.Render()

				if saveOptions[d] {
					err = saveToCSV(d, records)
					if err != nil {
						color.Red("Error saving to CSV: %v", err)
					} else {
						color.Green("Results for %s saved to %s.csv", d, d)
					}
				}
			}(strings.TrimSpace(domain))
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for range ch {
			// We use the channel to wait for all goroutines to complete.
		}
	}
}
