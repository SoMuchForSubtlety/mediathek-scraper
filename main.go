package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/graphql"
)

type query struct {
	SearchPage struct {
		Title         graphql.String
		VodPageNumber graphql.Int
		VodTotal      graphql.Int
		VodResults    []result
	} `graphql:"searchPage(client: \"ard\", text: \"tatort\", pageNumber: $number)"`
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
	client := graphql.NewClient("https://api.ardmediathek.de/public-gateway", nil)

	var numberOfVods resultNumber
	err := client.Query(context.Background(), &numberOfVods, nil)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	number := int(numberOfVods.SearchPage.VodTotal)
	fmt.Println(number / 24)
	fmt.Println(number)

	for i := 0; i <= (number / 24); i++ {
		fmt.Printf("querying page %v\n", i)
		var page query

		variables := map[string]interface{}{
			"number": graphql.Int(i),
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
		if len(page.SearchPage.VodResults) < 1 {
			fmt.Println("no content")
			break
		}
		for _, result := range page.SearchPage.VodResults {
			if len(result.MediumTitle) > 6 && result.MediumTitle[:7] == "Tatort:" {
				fmt.Println(result.MediumTitle)
				fmt.Println(result.Links.Target.Href)
			}
		}
	}
}
