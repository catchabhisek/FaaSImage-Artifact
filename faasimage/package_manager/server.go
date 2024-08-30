package main

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Package struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

var packageDir = "/home/user/FaaSSnapper/analysis/faasimage/data/packages"

func get_hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func get_files(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func addFileToTar(filePath string, tw *tar.Writer, addHash bool, is_symlink bool) error {

	relativePath, nil := filepath.Rel(packageDir, filePath)
	repoPath := filepath.Join("package_repo", relativePath)
	var filesize int64
	var hash string
	var tempFile *os.File

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return nil
	}

	if is_symlink {
		link, err := os.Readlink(filePath)
		if err != nil {
			fmt.Println("I am here baby")
			return err
		}

		header := &tar.Header{
			Name:     repoPath,
			Linkname: link,
			Typeflag: tar.TypeSymlink,
			ModTime:  fileInfo.ModTime(),
		}

		if err := tw.WriteHeader(header); err != nil {
			fmt.Println(err)
			return err
		}

		return nil

	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if addHash {
		hash, _ = get_hash(filePath)
		filesize = int64(len(hash))
		// Create temporary file for hash
		tempFile, err = os.CreateTemp("", "hash.txt")
		if err != nil {
			fmt.Println("Error creating temporary file:", err)
			return err
		}
		defer tempFile.Close() // Ensure temporary file is closed

		// Write hash to temporary file
		_, err = tempFile.WriteString(hash)
		if err != nil {
			fmt.Println("Error writing hash to temporary file:", err)
			return err
		}
	} else {
		filesize = stat.Size()
	}

	header := &tar.Header{
		Name:    repoPath,
		Mode:    int64(stat.Mode()),
		Size:    filesize,
		ModTime: fileInfo.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		fmt.Println(err)
		return err
	}

	if addHash {
		_, err = tempFile.Seek(0, 0) // Rewind temporary file to read from the beginning
		if err != nil {
			fmt.Println("Error seeking temporary file:", err)
			return err
		}
		_, err = io.Copy(tw, tempFile)
		if err != nil {
			fmt.Println("Error writing hash to archive:", err)
			return err
		}

	} else {
		if _, err = io.Copy(tw, file); err != nil {
			return err
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func handle_packages(w http.ResponseWriter, r *http.Request) {

	// Parse incoming JSON files
	var packages []Package

	if err := json.NewDecoder(r.Body).Decode(&packages); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create tar archive
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Print the constructed Package structs
	// for _, pkg := range packages {
	// 	fmt.Println("Package Name:", pkg.Name)
	// 	fmt.Println("Package Files:", strings.Join(pkg.Files, ", "))
	// }

	for _, pkg := range packages {
		// Get all files in the directory
		pkg_path := filepath.Join(packageDir, pkg.Name)
		pkg_files, err := get_files(pkg_path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, file := range pkg_files {
			if strings.Contains(file, "__pycache__") {
				fmt.Println(file)
				if err := addFileToTar(file, tw, false, true); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				continue
			}

			relPath, _ := filepath.Rel(pkg_path, file)

			fileInfo, err := os.Stat(file)
			if err != nil {
				fmt.Println((err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !fileInfo.IsDir() {
				if contains(pkg.Files, relPath) {
					if err := addFileToTar(file, tw, false, false); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				} else {
					if err := addFileToTar(file, tw, true, false); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}
		}
	}

	// Set response headers
	w.Header().Set("Content-Disposition", "attachment; filename=packages.tar")
	w.Header().Set("Content-Type", "application/octet-stream")

	// Write tar archive to response writer
	if _, err := buf.WriteTo(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/upload", handle_packages)
	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
