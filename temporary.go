package main

import (
	"fmt"
	"yamlchecker-cli/checker"
)

func main() {
	results := checker.CheckHTTP("Google", "https://httpbin.org/status/500")
	fmt.Println(results)
}
