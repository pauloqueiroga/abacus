package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
)

func getJiraTickets(baseUrl, jql, jiraTicketsCsvPath string) error {
	escapedJql := url.PathEscape(jql)
	url := fmt.Sprintf("%s/rest/api/3/search?maxResults=500&fields=id&jql=%s", baseUrl, escapedJql)
	mapped, err := jiraGetJson(url)
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
	mapped, err := jiraGetJson(url)
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

func setJiraField(baseUrl, fieldName, jiraSetCsvPath, jiraSetOutputCsvPath string) error {
	jiraFields, err := getJiraFields(baseUrl)
	if err != nil {
		return err
	}
	fieldId := jiraFields[fieldName]

	input, err := os.Open(jiraSetCsvPath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(jiraSetOutputCsvPath)
	if err != nil {
		return err
	}
	defer output.Close()

	appendRecord(output,
		"issueId",
		"issueKey",
		"fieldName",
		"previousValue",
		"newValue",
	)

	csvReader := csv.NewReader(input)
	field, err := parseFields(csvReader)
	if err != nil {
		return err
	}

	for {
		issue, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		issueKey := issue[field["issueKey"]]
		valueToSet := issue[field[fieldName]]
		previousValues, err := getJiraIssue(baseUrl, issueKey)
		if err != nil {
			return err
		}

		previousFields := previousValues["fields"].(map[string]interface{})
		originalValue := fmt.Sprintf("%v", previousFields[fieldId])

		if originalValue != valueToSet {
			err := updateJiraIssue(baseUrl, issueKey, fieldId, valueToSet)
			if err != nil {
				return err
			}
		}

		appendRecord(output,
			previousValues["id"].(string),
			issueKey,
			fieldName,
			originalValue,
			valueToSet,
		)
	}

	return nil
}

func getJiraIssue(baseUrl, issueKey string) (map[string]any, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", baseUrl, issueKey)
	return jiraGetJson(url)
}

func getJiraFields(baseUrl string) (map[string]string, error) {
	url := fmt.Sprintf("%s/rest/api/3/field", baseUrl)
	mapped, err := jiraGetJsonArray(url)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]string)

	for _, field := range mapped {
		fieldMap := field.(map[string]any)
		fields[fieldMap["name"].(string)] = fieldMap["id"].(string)
	}

	return fields, nil
}

func updateJiraIssue(baseUrl, issueKey, fieldKey, valueToSet string) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s", baseUrl, issueKey)
	body := fmt.Sprintf(`{"fields": {"%s": %s}}`, fieldKey, valueToSet)
	commandLine := exec.Command("curl",
		"--request", "PUT", url,
		"--header", "Content-Type: application/json",
		"--data", body,
		"--basic", "-u", os.Getenv("JIRA_USER")+":"+os.Getenv("JIRA_TOKEN"),
	)
	out, err := commandLine.Output()
	if err != nil {
		return err
	}

	log.Println(string(out))
	return nil
}

func jiraGetJson(url string) (map[string]any, error) {
	out, err := jiraGetBytes(url)
	if err != nil {
		return nil, err
	}

	var mapped map[string]any
	err = json.Unmarshal(out, &mapped)
	if err != nil {
		return nil, err
	}

	return mapped, nil
}

func jiraGetJsonArray(url string) ([]any, error) {
	out, err := jiraGetBytes(url)
	if err != nil {
		return nil, err
	}

	var mapped []any
	err = json.Unmarshal(out, &mapped)
	if err != nil {
		return nil, err
	}

	return mapped, nil
}

func jiraGetBytes(url string) ([]byte, error) {
	commandLine := exec.Command("curl", "--get", url, "--basic", "-u", os.Getenv("JIRA_USER")+":"+os.Getenv("JIRA_TOKEN"))
	return commandLine.Output()
}
