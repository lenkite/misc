package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var sdkLinkRE = regexp.MustCompile("id=\"latestNeo.*?\" href=\"(sdk/.*?.zip)\"")
var toolsPageUrl = "https://tools.hana.ondemand.com"

func main() {
	toolsPageBody := getBody(toolsPageUrl)
	matches := sdkLinkRE.FindAllStringSubmatch(toolsPageBody, -1)
	if matches == nil {
		log.Fatal("Could not find links in Tools Body Page!")
	}
	for _, m := range matches {
		if len(m) < 2 {
			log.Fatalf("Unexpected Match: %v", m)
		}
		fmt.Println(toolsPageUrl + "/" + m[1])
	}
}

// Gets the Body
func getBody(url string) (body string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body = string(b)
	return
}
