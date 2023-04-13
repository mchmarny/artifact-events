# artifact-events

Collection of workflows based on Google Container and Artifact Registry events:

* [Scan new GCR or AR image](./workflows/scan/README.md) for vulnerabilities and save resulting data in BigQuery
* [Dispatch vulnerabilities found by Artifact Analyses](./workflows/dispatch/README.md) to custom remediation systems (e.g. Jira), or chat service (e.g. Slack)

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

