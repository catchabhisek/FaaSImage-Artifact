package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Package struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

func ParsePackages(bench string) ([]*Package, error) {
	// Read the file
	fileData, err := ioutil.ReadFile(fmt.Sprintf("/home/XXXX/FaaSSnapper/analysis/faasimage/data/input/%s.json", bench))
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, nil
	}

	// Parse JSON data
	var packagesData map[string][]string
	if err := json.Unmarshal(fileData, &packagesData); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, nil
	}

	extractDir := "./tmp"

	// Construct Package structs
	var packages []*Package
	for packageName, files := range packagesData {
		var missingFiles []string

		for _, file := range files {
			fileInfo, err := os.Stat(filepath.Join(extractDir, "package_repo", packageName, file))

			if os.IsNotExist(err) || (fileInfo.Size() == 32) {
				missingFiles = append(missingFiles, file)
			}
		}

		pkg := Package{
			Name:  packageName,
			Files: missingFiles,
		}
		packages = append(packages, &pkg)
	}

	return packages, nil
}

func main() {

	url := os.Args[1]
	parts := strings.Split(url, "/")
	benchmark := strings.TrimSuffix(parts[len(parts)-1], "_faas")

	startTime := time.Now()
	packages, err := ParsePackages(benchmark)

	if err != nil {
		fmt.Println("Error parsing packages:", err)
		return
	}

	// // Print the constructed Package structs
	// for _, pkg := range packages {
	// 	fmt.Println("Package Name:", pkg.Name)
	// 	fmt.Println("Package Files:", strings.Join(pkg.Files, ", "))
	// }

	// Convert packages info to JSON
	packagesJSON, err := json.Marshal(packages)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Send JSON packages to server
	resp, err := http.Post("http://localhost:8080/upload", "application/json", bytes.NewBuffer(packagesJSON))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Server returned non-200 status code:", resp.Status)
		return
	}

	// Save received tar archive
	tarData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Create a directory to extract the files
	extractDir := "./tmp"
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// Create a buffer from the tar archive data
	tarBuffer := bytes.NewBuffer(tarData)

	// Create a tar reader
	tr := tar.NewReader(tarBuffer)

	// Extract files from the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			fmt.Println("Error reading tar header:", err)
			return
		}

		// Construct path for extracted file
		extractPath := filepath.Join(extractDir, header.Name)

		// Create directories as necessary
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(extractPath, 0755); err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
			continue
		}

		// Get the directory path of the file
		dir := filepath.Dir(extractPath)

		// Create the directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		// Create directories as necessary
		if header.Typeflag == tar.TypeSymlink {
			fmt.Println(extractPath)
			err := os.Symlink(header.Linkname, extractPath)
			if err != nil {
				fmt.Println("Error creating symlink:", err)
				return
			}
			continue
		}

		// Create extracted file
		extractFile, err := os.Create(extractPath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer extractFile.Close()

		// Write file contents from tar archive
		if _, err := io.Copy(extractFile, tr); err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
	}

	fmt.Println("Tar archive extracted successfully.")
	elapsedTime := time.Since(startTime)
	fmt.Println("Execution time of someFunction:", elapsedTime)
}
