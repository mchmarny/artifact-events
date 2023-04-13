# artifact-events

Collection of workflows based on Google Container and Artifact Registry events:

* [Event on new GCR or AR image](./workflows/scan/README.md) - subscribe to the Pub/Sub events published by the registry, scans each new image for vulnerabilities and save resulting data into GCS bucket
* [Event on vulnerabilities found by Artifact Analyses](./workflows/dispatch/README.md) - subscribe to the Pub/Sub events published by Artifact Analyses service and dispatch the discovered vulnerability information to an internal remediation systems like (e.g. Jira), chat service (e.g. Slack), or custom REST API endpoint.

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

