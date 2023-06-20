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
		if len(os.Args) < 5 {
			fmt.Println("At least four arguments are required.")
			printUsage()
			return
		}
		if len(os.Args) >= 7 {
			pullrequestsCsvPath = os.Args[6]
		}
		if len(os.Args) >= 6 {
			projectsCsvPath = os.Args[5]
		}
		getPullRequests(os.Args[2], os.Args[3], os.Args[4], projectsCsvPath, pullrequestsCsvPath)
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
	appendRecord(output, "id", "name")

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

func getPullRequests(baseUrl, minDate, maxDate, projectsCsvPath, pullrequestsCsvPath string) error {
	output, err := os.Create(pullrequestsCsvPath)
	if err != nil {
		return err
	}
	defer output.Close()
	appendRecord(output,
		"pullRequestId",
		"authorId",
		"creationDate",
		"closedDate",
		"sourceRefName",
		"targetRefName",
		"mergeStatus",
		"reviewersCount",
		"url",
	)

	url := fmt.Sprintf("%s/_apis/git/pullrequests?api-version=7.1-preview.1&searchCriteria.status=completed&searchCriteria.minTime=%s&searchCriteria.maxTime=%s", baseUrl, minDate, maxDate)
	req, err := newHttpRequest("get", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	prBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var mapped map[string]any
	err = json.Unmarshal(prBody, &mapped)
	if err != nil {
		return err
	}

	prList := mapped["value"].([]interface{})

	for _, p := range prList {
		pr := p.(map[string]any)
		author := pr["createdBy"].(map[string]any)
		reviewers := pr["reviewers"].([]interface{})
		appendRecord(output,
			fmt.Sprint(pr["pullRequestId"]),
			author["id"].(string),
			pr["creationDate"].(string),
			pr["closedDate"].(string),
			pr["sourceRefName"].(string),
			pr["targetRefName"].(string),
			pr["mergeStatus"].(string),
			fmt.Sprint(len(reviewers)),
			pr["url"].(string),
		)
	}

	return nil
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
	fmt.Println("  abacus pullrequests <URL to AzDO org> <minimum date> <maximum date> [projects CSV path] [output CSV path]")
	fmt.Println("               Retrieves Pull Requests for each of the project/user combination")
}
