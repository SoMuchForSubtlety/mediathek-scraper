package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
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
	} `graphql:"searchPage(client: \"ard\", text: $searchTerm)"`
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
	workers := flag.Int("workers", 1, "the maximum number of parallel downloads")
	searchTerm := flag.String("search", "", "the term to search for")
	regex := flag.String("regex", "", "regular expression that needs to be matched by the title")
	dlLocation := flag.String("path", "", "the location to save the downloaded files")
	download := flag.Bool("download", false, "download the search results")
	flag.Parse()

	var r *regexp.Regexp

	if *searchTerm == "" {
		fmt.Println("please provide a search term")
		return
	} else if *workers < 1 {
		fmt.Println("there needs to be at least one worker")
		return
	} else if *dlLocation != "" {
		if _, err := os.Stat(*dlLocation); os.IsNotExist(err) {
			fmt.Println("specified path does not exist")
			return
		} else if (*dlLocation)[len(*dlLocation)-1] != '/' {
			*dlLocation += "/"
		}
	}

	if *regex != "" {
		// check for regex match
		var err error
		r, err = regexp.Compile(*regex)
		if err != nil {
			fmt.Println("invalid regex")
			return
		}
	}

	guard := make(chan bool, *workers)

	client := graphql.NewClient("https://api.ardmediathek.de/public-gateway", nil)

	var numberOfVods resultNumber

	variables := map[string]interface{}{
		"searchTerm": graphql.String(*searchTerm),
	}

	err := client.Query(context.Background(), &numberOfVods, variables)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Found %v results\n\n", int(numberOfVods.SearchPage.VodTotal))

	lines, err := readLines(path)
	if err != nil {
		log.Fatal(err)
	}
	var dlWait sync.WaitGroup
	for i := 0; i <= (int(numberOfVods.SearchPage.VodTotal) / 24); i++ {
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
				fmt.Printf("ERROR: %v\n", err)
				return
			}
		}
		for _, result := range page.SearchPage.VodResults {
			if *minDuration != -1 && int(result.Duration) < *minDuration {
				// check for min duration
				continue
			} else if *maxDuration != -1 && int(result.Duration) > *maxDuration {
				// check for max duration
				continue
			} else if r != nil {
				// check for regex match
				if !r.MatchString(string(result.MediumTitle)) {
					continue
				}
			}
			title := string(result.MediumTitle)
			r, _ := regexp.Compile("[^/]*$")
			link := "https://www.ardmediathek.de/ard/player/" + r.FindString(string(result.Links.Target.Href))
			guard <- true
			fmt.Printf("Foud: %v\n", title)
			fmt.Printf("%v\n\n", link)
			if *download && checkIfNew(title, lines) {
				fmt.Println("starting download")
				fmt.Println("")
				go func() {
					dlWait.Add(1)
					err := downloadVOD(link, *dlLocation+title)
					if err != nil {
						fmt.Println(err)
					} else {
						addEntry(title)
					}
					<-guard
					dlWait.Done()
				}()
			} else {
				<-guard
			}
		}
	}
	dlWait.Wait()
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
