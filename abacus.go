package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("At least two arguments are required.")
		printUsage()
		return
	}

	projectsCsvPath := "local-only/projects.csv"
	usersCsvPath := "local-only/users.csv"
	pullrequestsCsvPath := "local-only/pullrequests.csv"

	switch os.Args[1] {
	case "projects":
		if len(os.Args) >= 4 {
			projectsCsvPath = os.Args[3]
		}
		err := getProjects(os.Args[2], projectsCsvPath)
		if err != nil {
			log.Fatal(err)
		}

	case "pullrequests":
		if len(os.Args) >= 6 {
			pullrequestsCsvPath = os.Args[5]
		}
		if len(os.Args) >= 5 {
			usersCsvPath = os.Args[4]
		}
		if len(os.Args) >= 4 {
			projectsCsvPath = os.Args[3]
		}
		getPullRequests(os.Args[2], projectsCsvPath, usersCsvPath, pullrequestsCsvPath)
	default:
		fmt.Println("Don't know how to process this command...", os.Args)
		printUsage()
		return
	}
}

func getProjects(baseUrl, projectsCsvPath string) error {
	output, err := os.Create(projectsCsvPath)
	if err != nil {
		return err
	}
	defer output.Close()

	req, err := newHttpRequest("get", baseUrl+"/_apis/projects?api-version=7.0", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	projBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var mapped map[string]any
	err = json.Unmarshal(projBody, &mapped)
	if err != nil {
		return err
	}

	projList := mapped["value"].([]interface{})

	for _, p := range projList {
		proj := p.(map[string]any)
		appendRecord(output, proj["id"].(string), proj["name"].(string))
	}

	return nil
}

func getPullRequests(baseUrl, projectsCsvPath, usersCsvPath, pullrequestsCsvPath string) {
	panic("unimplemented")
}

func newHttpRequest(method string, url string, reqBody io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	r.SetBasicAuth("anything", os.Getenv("AZDO_TOKEN"))
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-type", "application/json")
	return r, nil
}

func appendRecord(file *os.File, fields ...string) {
	escapedJson, err := json.Marshal(fields)
	if err != nil {
		log.Fatal(err)
	}

	escapedCsv := strings.Trim(string(escapedJson), "[]")
	_, err = file.WriteString(fmt.Sprintf("%s\n", escapedCsv))
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("USAGE:")
	fmt.Println("  abacus projects <URL to AzDO org> [output CSV path]")
	fmt.Println("            Retrieves a list of projects from the given Azure DevOps Organization URL")
	fmt.Println("  abacus pullrequests <URL to AzDO org> [projects CSV path] [users CSV path] [output CSV path]")
	fmt.Println("               Retrieves Pull Requests for each of the project/user combination")
}
