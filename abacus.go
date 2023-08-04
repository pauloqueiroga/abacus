package main

import (
	"encoding/csv"
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
	gitlogCsvPath := "local-only/gitlog.csv"
	jiraTicketsCsvPath := "local-only/jira-tickets.csv"
	localReposFolder := "local-only/repos"

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
		if len(os.Args) >= 6 {
			pullrequestsCsvPath = os.Args[5]
		}
		err := gatherPullRequests(os.Args[2], os.Args[3], os.Args[4], pullrequestsCsvPath)
		if err != nil {
			log.Fatal(err)
		}
	case "gitlogs":
		if len(os.Args) < 3 {
			fmt.Println("At least two arguments are required.")
			printUsage()
			return
		}
		if len(os.Args) >= 4 {
			pullrequestsCsvPath = os.Args[3]
		}
		if len(os.Args) >= 5 {
			gitlogCsvPath = os.Args[4]
		}
		if len(os.Args) >= 6 {
			localReposFolder = os.Args[5]
		}
		err := calculateGitLogStats(os.Args[2], pullrequestsCsvPath, gitlogCsvPath, localReposFolder)
		if err != nil {
			log.Fatal(err)
		}
	case "jql-pr":
		if len(os.Args) < 4 {
			fmt.Println("At least three arguments are required.")
			printUsage()
			return
		}
		if len(os.Args) >= 5 {
			jiraTicketsCsvPath = os.Args[4]
		}
		err := getJiraTickets(os.Args[2], os.Args[3], jiraTicketsCsvPath)
		if err != nil {
			log.Fatal(err)
		}
	case "help":
		printUsage()
		return
	default:
		fmt.Println("Don't know how to process this command...", os.Args)
		printUsage()
		return
	}
}

func parseFields(r *csv.Reader) (map[string]int, error) {
	record, err := r.Read()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for i, v := range record {
		result[v] = i
	}

	return result, nil
}

func newHttpRequest(authUser, authPass string, method string, fullUrl string, reqBody io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, fullUrl, reqBody)
	if err != nil {
		return nil, err
	}

	r.SetBasicAuth(authUser, authPass)
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
	fmt.Println("  abacus pullrequests <URL to AzDO org> <minimum date> <maximum date> [output CSV path]")
	fmt.Println("            Retrieves Pull Requests' metadata")
	fmt.Println("  abacus gitlogs <Base Git URL> [input CSV path] [output CSV path] [local path for Git repositories]")
	fmt.Println("            Retrieves Git log statistics for the repositories and branches in the input CSV file")
	fmt.Println("  abacus jql-pr <URL to Jira> <JQL query> [output CSV path]")
	fmt.Println("            Retrieves Jira issues' metadata and linked Pull Requests")
	fmt.Println("  abacus help")
	fmt.Println("            Prints this message")
}
