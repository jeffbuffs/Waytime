package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
)

func main() {
	domains := make(chan string)
	var wg sync.WaitGroup

	// Start 10 workers to fetch robots.txt files
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range domains {
				url := "http://web.archive.org/cdx/search/cdx?url=" + domain + "/robots.txt&output=json"
				wayresp, err := http.Get(url)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				defer wayresp.Body.Close()

				var records [][]string
				err = json.NewDecoder(wayresp.Body).Decode(&records)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}

				if len(records) >= 1 {
					for _, record := range records[1:] {
						timestamp := record[1]
						waybackURL := fmt.Sprintf("http://web.archive.org/web/%s/%s", timestamp+"if_", record[2])
						resp, err := http.Get(waybackURL)
						if err != nil {
							fmt.Println("Error While Getting Data from urls", err)
							continue
						}
						body, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							fmt.Println("Error While Reading response", err)
							continue
						}
						sb := string(body)
						pathpattern := regexp.MustCompile(`/[\w\-./?=]*`)
						pathma := pathpattern.FindAllStringSubmatch(sb, -1)
						if resp.StatusCode >= 200 && resp.StatusCode <= 200 {
							for _, path := range pathma {
								fmt.Println(path[0])
							}
						}
					}
				}
			}
		}()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		domains <- scanner.Text()
	}
	close(domains)
	wg.Wait()
}
