package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

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
			log.Println("Skipping clone of", gitUrl)
		}
		err = getGitLog(localReposFolder, repository, pr[field["lastMergeCommit"]], pr[field["pullRequestId"]], pr[field["authorUsername"]], pr[field["closedDate"]], output)
		if err != nil {
			return err
		}
	}

	return nil
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
