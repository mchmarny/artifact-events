# artifact-events

Collection of workflows based on Google Container and Artifact Registry events:

* [Event on new GCR or AR image](./workflows/scan/README.md) - this example scans each new image for vulnerabilities and save resulting data into GCS bucket
* [Event on vulnerabilities found by Artifact Analyses](./workflows/dispatch/README.md) - this example dispatch vulnerability information to custom remediation systems (e.g. Jira), or chat service (e.g. Slack)

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

