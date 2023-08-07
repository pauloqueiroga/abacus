# Abacus

Abacus is a simple and rudimentary calculator that can perform basic operations on Jira and Azure DevOps APIs. Check out the `printUsage()` function within `abacus.go` for a list of available commands and arguments.

## Environment Variables

Abacus requires the following environment variables to be set:

- `AZDO_TOKEN`: The Azure DevOps token to use for authentication. See [Personal Access Tokens](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops).
- `JIRA_USER`: The Jira username to use for authentication.
- `JIRA_TOKEN`: The Jira token to use for authentication. See [API Tokens](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/).

## Jira operations

### `jql-pr <URL to Jira> <JQL query> [output CSV path]`

`jql-pr` is a command that takes a JQL query and returns a list of pull requests that match the query. This command makes use of at least one undocumented Jira API endpoint, so there's a chance it might not work sometime in the future. The command takes the following arguments:

- `URL to Jira`: Mandatory. The URL to the Jira instance to query.
- `JQL query`: Mandatory. The JQL query to execute.
- `output CSV path`: Optional. The path to the CSV file to write the results to. If not specified, the results will be written to a default output CSV path.
