package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shurcooL/graphql"
)

const path = "log.txt"

type query struct {
	SearchPage struct {
		Title         graphql.String
		VodPageNumber graphql.Int
		VodTotal      graphql.Int
		VodResults    []result
	} `graphql:"searchPage(client: \"ard\", text: $searchTerm, pageNumber: $number)"`
}
type resultNumber struct {
	SearchPage struct {
		VodTotal graphql.Int
	} `graphql:"searchPage(client: \"ard\", text: \"tatort\")"`
}

type result struct {
	Type        graphql.String
	MediumTitle graphql.String
	Duration    graphql.Int
	Links       struct {
		Target struct {
			Href graphql.String
		}
	}
}

func main() {
	minDuration := flag.Int("minduration", -1, "the minimum duration (in seconds) for a VOD to be considured")
	maxDuration := flag.Int("maxduration", -1, "the maximum duration (in seconds) for a VOD to be considured")
	searchTerm := flag.String("search", "", "the term to search for")
	regex := flag.String("regex", "", "regex the title needs to match")
	dlLocation := flag.String("path", "", "the location to save the downloaded files")
	download := flag.Bool("download", false, "download the VODs")
	flag.Parse()

	if *searchTerm == "" {
		fmt.Println("please provide a search term")
		return
	} else if *dlLocation != "" {
		if _, err := os.Stat(*dlLocation); os.IsNotExist(err) {
			fmt.Println("specified path does not exist")
			return
		} else if (*dlLocation)[len(*dlLocation)-1] != '/' {
			*dlLocation += "/"
		}
	}

	client := graphql.NewClient("https://api.ardmediathek.de/public-gateway", nil)

	var numberOfVods resultNumber
	err := client.Query(context.Background(), &numberOfVods, nil)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	lines, err := readLines(path)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i <= (int(numberOfVods.SearchPage.VodTotal) / 24); i++ {
		fmt.Printf("querying page %v\n", i)
		var page query

		variables := map[string]interface{}{
			"number":     graphql.Int(i),
			"searchTerm": graphql.String(*searchTerm),
		}

		err := client.Query(context.Background(), &page, variables)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			time.Sleep(time.Second * 1)
			err := client.Query(context.Background(), &page, variables)
			if err != nil {
				continue
			}
		}
		for _, result := range page.SearchPage.VodResults {
			if *minDuration != -1 && int(result.Duration) < *minDuration {
				// check for min duration
				continue
			} else if *maxDuration != -1 && int(result.Duration) > *maxDuration {
				// check for max duration
				continue
			} else if *regex != "" {
				// check for regex match
				r, err := regexp.Compile(*regex)
				if err != nil {
					fmt.Println("invalid regex")
					os.Exit(1)
				}
				if !r.MatchString(string(result.MediumTitle)) {
					continue
				}
			}
			titleRegex, _ := regexp.Compile("^[^(-]*")
			properTitle := titleRegex.FindString(string(result.MediumTitle))
			properTitle = strings.TrimSpace(properTitle)
			fmt.Println(result.MediumTitle)
			if *download && checkIfNew(properTitle, lines) {
				addEntry(properTitle)
				lines = append(lines, properTitle)
				r, _ := regexp.Compile("[^/]*$")
				fmt.Printf("downloading %v\n", result.MediumTitle)
				err := downloadVOD("https://www.ardmediathek.de/ard/player/"+r.FindString(string(result.Links.Target.Href)), *dlLocation+properTitle)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func addEntry(input string) error {
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(input + "\n"); err != nil {
		return err
	}
	return nil
}

func checkIfNew(input string, lines []string) bool {
	for _, line := range lines {
		if line == input {
			return false
		}
	}
	return true
}

func readLines(path string) ([]string, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
