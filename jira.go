package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
)

func getJiraTickets(baseUrl, jql, jiraTicketsCsvPath string) error {
	escapedJql := url.PathEscape(jql)
	url := fmt.Sprintf("%s/rest/api/3/search?maxResults=50&fields=id&jql=%s", baseUrl, escapedJql)

	commandLine := exec.Command("curl", "--get", url, "--basic", "-u", os.Getenv("JIRA_USER")+":"+os.Getenv("JIRA_TOKEN"))
	out, err := commandLine.Output()
	if err != nil {
		return err
	}

	output, err := os.Create(jiraTicketsCsvPath)
	if err != nil {
		return err
	}
	defer output.Close()

	appendRecord(output,
		"issueId",
		"issueKey",
		"prId",
		"prUrl",
		"prStatus",
	)

	var mapped map[string]any
	err = json.Unmarshal(out, &mapped)
	if err != nil {
		return err
	}

	issueList := mapped["issues"].([]interface{})

	for _, record := range issueList {
		issue := record.(map[string]any)
		id := issue["id"].(string)
		key := issue["key"].(string)
		err := getJiraPullRequestInfo(baseUrl, id, key, output)
		if err != nil {
			log.Println("No pull request found for", id, key)
			continue
		}
	}

	return nil
}

func getJiraPullRequestInfo(baseUrl, id, key string, output *os.File) error {
	url := fmt.Sprintf("%s/rest/dev-status/1.0/issue/detail?issueId=%s&applicationType=GitForJiraCloud&dataType=pullrequest", baseUrl, id)
	commandLine := exec.Command("curl", "--get", url, "--basic", "-u", os.Getenv("JIRA_USER")+":"+os.Getenv("JIRA_TOKEN"))
	out, err := commandLine.Output()
	if err != nil {
		return err
	}

	var mapped map[string]any
	err = json.Unmarshal(out, &mapped)
	if err != nil {
		return err
	}

	detailList := mapped["detail"].([]interface{})
	for _, record := range detailList {
		detail := record.(map[string]any)
		prList := detail["pullRequests"].([]interface{})
		for _, prRecord := range prList {
			pr := prRecord.(map[string]any)
			prId := pr["id"].(string)
			prUrl := pr["url"].(string)
			prStatus := pr["status"].(string)

			appendRecord(output,
				id,
				key,
				prId,
				prUrl,
				prStatus,
			)
		}
	}

	return nil
}
