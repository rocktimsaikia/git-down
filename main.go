package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

type UrlData struct {
	user_name   string
	repo_name   string
	branch_name string
	dirs        []string
}

// Get the required data from the given Github URL
func GetUrlData(url string) UrlData {
	chunks := strings.Split(url, "/")
	ignored_chunks := []string{"https:", "github.com", "tree", ""}
	var data []string
	for _, chunk := range chunks {
		if !slices.Contains(ignored_chunks, chunk) {
			data = append(data, chunk)
		}
	}
	return UrlData{
		user_name:   data[0],
		repo_name:   data[1],
		branch_name: data[2],
		dirs:        data[3:],
	}
}

// Construct the API URL from the given URL data
func StructApiUrl(url_data UrlData) string {
	dir_string := strings.Join(url_data.dirs, "/")
	download_url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		url_data.user_name,
		url_data.repo_name,
		dir_string,
		url_data.branch_name,
	)
	return download_url
}

// Download a file from Github
func DownloadFile(url string, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: %s", resp.Status)
	}

	// Create the destination file
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the response body to the destination file
	_, err = io.Copy(file, resp.Body)
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please provide a github url")
		return
	}

	github_url := args[0]

	if github_url == "" {
		fmt.Println("Please provide a github url")
		return
	}

	url_data := GetUrlData(args[0])
	api_url := StructApiUrl(url_data)

	resp, err := http.Get(api_url)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer resp.Body.Close()

	resp_data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	json_data := string(resp_data)
	var list_data []map[string]interface{}
	json.Unmarshal([]byte(json_data), &list_data)

	download_dir := &url_data.dirs[len(url_data.dirs)-1]
	os.Mkdir(*download_dir, os.ModePerm)

	for _, data := range list_data {
		file_name := data["name"]
		download_url := data["download_url"]
		download_file_name := fmt.Sprintf("%s/%s", *download_dir, file_name)
		err := DownloadFile(download_url.(string), download_file_name)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		fmt.Println("Downloaded ", file_name, " to ", download_file_name)
	}
}
