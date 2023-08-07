# Abacus

Abacus is a simple and rudimentary calculator that can perform basic operations on Jira and Azure DevOps APIs, and Git. Check out the `printUsage()` function within `abacus.go` for a list of available commands and arguments.

## Environment Variables

Abacus requires the following environment variables to be set:

- `AZDO_TOKEN`: The Azure DevOps token to use for authentication. See [Personal Access Tokens](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops).
- `JIRA_USER`: The Jira username to use for authentication.
- `JIRA_TOKEN`: The Jira token to use for authentication. See [API Tokens](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/).

## Azure DevOps operations

### `projects <URL to AzDO org> [output CSV path]`

`projects` is a command that takes a URL to an Azure DevOps organization and returns a list of projects in that organization. The command takes the following arguments:

- `URL to AzDO org`: Mandatory. The URL to the Azure DevOps organization to query.
- `output CSV path`: Optional. The path to the CSV file to write the results to. If not specified, the results will be written to a default output CSV path.

### `pullrequests <URL to AzDO org> <minimum date> <maximum date> [output CSV path]`

`pullrequests` is a command that takes a URL to an Azure DevOps organization and returns a list of completed pull requests with `main` or `master` as the target branch. The command takes the following arguments:

- `URL to AzDO org`: Mandatory. The URL to the Azure DevOps organization to query.
- `minimum date`: Mandatory. The minimum date to query for pull requests. The date must be in the format `YYYY-MM-DD`.
- `maximum date`: Mandatory. The maximum date to query for pull requests. The date must be in the format `YYYY-MM-DD`.
- `output CSV path`: Optional. The path to the CSV file to write the results to. If not specified, the results will be written to a default output CSV path.

## Git operations

### `gitlogs <Base Git URL> [input CSV path] [output CSV path] [local path for Git repositories]`

`gitlogs` makes a lot of assumptions about your intentions and the structure of the repositories URLs, which may be specific to how Git repositories hosted in Azure DevOps behave. It also will attempt to clone the first target branch for each repository it finds in the input file, but will quickly give up if the local folder already exists. The command takes the following arguments:

- `Base Git URL`: Mandatory. The base URL to where Git repositories are found in your Azure DevOps organization. This might not be too obvious to figure out, you might have to do some digging around to find it.
- `input CSV path`: Optional. The path to the CSV file with pull request and repository data. This command was designed around the CSV output by the `pullrequests` command, but it allows for some flexibility, see the code for more details. If not specified, the command will look for the CSV file in a default input CSV path.
- `output CSV path`: Optional. The path to the CSV file to write the results to. If not specified, the results will be written to a default output CSV path.
- `local path for Git repositories`: Optional. The path to the directory where Git repositories will be cloned to. If not specified, the command will use a default local path for Git repositories.

## Jira operations

### `jql-pr <URL to Jira> <JQL query> [output CSV path]`

`jql-pr` is a command that takes a JQL query and returns a list of pull requests that match the query. This command makes use of at least one undocumented Jira API endpoint, so there's a chance it might not work sometime in the future. The command takes the following arguments:

- `URL to Jira`: Mandatory. The URL to the Jira instance to query.
- `JQL query`: Mandatory. The JQL query to execute.
- `output CSV path`: Optional. The path to the CSV file to write the results to. If not specified, the results will be written to a default output CSV path.
