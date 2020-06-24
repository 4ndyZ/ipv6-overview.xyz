package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Website struct {
	Name    string
	Twitter string
}

type YAMLConfig struct {
	Websites []*Website
}

func main() {
	yamlContent, _ := ioutil.ReadFile("config.yml")
	yamlconfig := &YAMLConfig{}
	yaml.Unmarshal([]byte(yamlContent), yamlconfig)

	counter := 0

	for _, website := range yamlconfig.Websites {
		if len(website.Twitter) == 0 {
			continue
		}

		counter++

		fmt.Printf("%s -> https://twitter.com/%s\n", website.Name, strings.ReplaceAll(website.Twitter, "@", ""))

		if counter%5 == 0 {
			fmt.Println()
			fmt.Println()
			counter = 0
		}
	}
}
