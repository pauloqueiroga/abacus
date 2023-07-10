package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

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
