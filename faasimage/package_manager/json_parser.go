package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type Package struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

func main() {
	// Read the file
	fileData, err := ioutil.ReadFile("/home/XXXX/FaaSSnapper/analysis/faasimage/data/input/gzip.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Parse JSON data
	var packagesData map[string][]string
	if err := json.Unmarshal(fileData, &packagesData); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Construct Package structs
	var packages []Package
	for packageName, files := range packagesData {
		pkg := Package{
			Name:  packageName,
			Files: files,
		}
		packages = append(packages, pkg)
	}

	// Print the constructed Package structs
	for _, pkg := range packages {
		fmt.Println("Package Name:", pkg.Name)
		fmt.Println("Package Files:", strings.Join(pkg.Files, ", "))
	}
}
