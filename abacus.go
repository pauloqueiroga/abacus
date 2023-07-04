package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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
	case "help":
		printUsage()
		return
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

func gatherPullRequests(baseUrl, minDate, maxDate, pullrequestsCsvPath string) error {
	output, err := os.Create(pullrequestsCsvPath)
	if err != nil {
		return err
	}
	defer output.Close()

	appendRecord(output,
		"pullRequestId",
		"authorId",
		"authorDescriptor",
		"authorUsername",
		"creationDate",
		"closedDate",
		"repository",
		"project",
		"sourceRefName",
		"targetRefName",
		"mergeStatus",
		"reviewersCount",
		"url",
		"lastMergeCommit",
	)
	err = getPullRequests(os.Args[2], os.Args[3], os.Args[4], "refs/heads/main", output)
	if err != nil {
		return err
	}

	err = getPullRequests(os.Args[2], os.Args[3], os.Args[4], "refs/heads/master", output)
	if err != nil {
		return err
	}

	return nil
}

func getPullRequests(baseUrl, minDate, maxDate, targetRefName string, output *os.File) error {
	url := fmt.Sprintf("%s/_apis/git/pullrequests?api-version=7.1-preview.1&searchCriteria.status=completed&searchCriteria.queryTimeRangeType=closed&searchCriteria.minTime=%s&searchCriteria.maxTime=%s&searchCriteria.targetRefName=%s&$top=1500", baseUrl, minDate, maxDate, targetRefName)
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
		lastMerge := pr["lastMergeCommit"].(map[string]any)
		repository := pr["repository"].(map[string]any)
		project := repository["project"].(map[string]any)
		appendRecord(output,
			fmt.Sprint(pr["pullRequestId"]),
			author["id"].(string),
			author["descriptor"].(string),
			author["uniqueName"].(string),
			pr["creationDate"].(string),
			pr["closedDate"].(string),
			repository["name"].(string),
			project["name"].(string),
			pr["sourceRefName"].(string),
			pr["targetRefName"].(string),
			pr["mergeStatus"].(string),
			fmt.Sprint(len(reviewers)),
			pr["url"].(string),
			lastMerge["commitId"].(string),
		)
	}

	return nil
}

func calculateGitLogStats(baseUrl, pullrequestsCsvPath, gitlogCsvPath, localReposFolder string) error {
	input, err := os.Open(pullrequestsCsvPath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(gitlogCsvPath)
	if err != nil {
		return err
	}
	defer output.Close()

	appendRecord(output,
		"pullRequestId",
		"authorUsername",
		"closedDate",
		"repository",
		"lastMergeCommit",
		"linesAdded",
		"linesRemoved",
		"filePath",
	)

	csvReader := csv.NewReader(input)
	field, err := parseFields(csvReader)
	if err != nil {
		return err
	}

	for {
		pr, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		project := url.PathEscape(pr[field["project"]])
		repository := url.PathEscape(pr[field["repository"]])
		branch := strings.ReplaceAll(pr[field["targetRefName"]], "refs/heads/", "")
		gitUrl := fmt.Sprintf("%s/%s/_git/%s", baseUrl, project, repository)
		err = cloneIfNeeded(gitUrl, localReposFolder, branch)
		if err != nil {
			log.Println(err.Error())
		}
		err = getGitLog(localReposFolder, repository, pr[field["lastMergeCommit"]], pr[field["pullRequestId"]], pr[field["authorUsername"]], pr[field["closedDate"]], output)
		if err != nil {
			return err
		}
	}

	return nil
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

func cloneIfNeeded(gitUrl, localReposFolder, targetRefName string) error {
	commandLine := exec.Command("git", "clone", gitUrl, "--branch", targetRefName, "--single-branch")
	commandLine.Dir = localReposFolder
	err := commandLine.Run()
	if err != nil {
		return err
	}

	return nil
}

func getGitLog(localReposFolder, repository, lastMergeCommit, prId, author, closedDate string, output *os.File) error {
	commandLine := exec.Command("git", "log", lastMergeCommit, "--numstat", "-1", "--format=")
	commandLine.Dir = fmt.Sprintf("%s/%s", localReposFolder, repository)
	out, err := commandLine.Output()
	if err != nil {
		return err
	}

	for _, stat := range strings.Split(string(out), "\n") {
		if stat == "" {
			continue
		}
		record := make([]string, 0)
		record = append(record, prId, author, closedDate, repository, lastMergeCommit)
		record = append(record, strings.Split(stat, "\t")...)
		appendRecord(output, record...)
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
	fmt.Println("  abacus pullrequests <URL to AzDO org> <minimum date> <maximum date> [output CSV path]")
	fmt.Println("               Retrieves Pull Requests' metadata")
	fmt.Println("  abacus gitlogs <Base Git URL> [input CSV path] [output CSV path] [local path for Git repositories]")
	fmt.Println("            Retrieves Git log statistics for the repositories and branches in the input CSV file")
	fmt.Println("  abacus help")
	fmt.Println("            Prints this message")
}
