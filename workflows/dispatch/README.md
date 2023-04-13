# artifact-events - vulnerabilities discovered by aa

If enabled, Artifact Analysis (aka Container Analysis) will create Pub/Sub events for each vulnerability found by automated AR scanning ([additional info](https://cloud.google.com/container-analysis/docs/pub-sub-notifications#container-analysis-pubsub-gcloud)). For the most part, you can think of a `note` as data about vulnerability (e.g. `CVE`), and `occurrence` as the data that connects that `note` to a particular artifact (i.e. `digest`). 

> Note: these events do not fire for notes and occurrence created via Artifact Analysis API.

Artifact Registry (AR) now also supports [remote repositories](https://cloud.google.com/artifact-registry/docs/repositories/remote-repo). This is basically a proxy to an upstream repo caches artifacts. You can create them for Docker Hub, Helm, dev language package like Maven or PiPI, or OS packages like Debian or RPM. 

The need part is that any artifact you pull via that kind of repo in AR can also benefit from the vulnerability scanning. Vulnerability events discovered that way will also be published to the topic just like AR does for "local" artifacts. This allows you to have aggregated notifications of artifact vulnerabilities whether they reside in AR or in a remote repo. 

GCP provides a number of ways in which you can process these events (Cloud Workflows, Cloud Build, Cloud Functions, etc.). In this walk-through we will overview:

* [manual](#manual) - quick way to get a few of the events output by AA
* [programmatic](#programmatic) - uses Cloud Function to dispatch AA events to Slack, Jira, REST API etc. 

## setup

If you haven't already done so, start by enabling the Artifact Analysis and Container Scanning API 

```shell
gcloud services enable containeranalysis.googleapis.com --project $PROJECT_ID
gcloud services enable containerscanning.googleapis.com --project $PROJECT_ID
```

When enabled, Artifact Analysis API will automatically creates the Pub/Sub occurrences topic. You can check if it exists:

```shell
gcloud pubsub topics list --project $PROJECT_ID
```

The results should include `container-analysis-occurrences-v1`. If these topics do not exist, you can create them yourself using the `gcloud pubsub topics create` command:

```shell
name: projects/$PROJECT/topics/container-analysis-occurrences-v1
```

### manual

The quickest way to capture messages output by Artifact Analyses (AA) is to create subscription on the `container-analysis-occurrences-v1` topic:

> The name doesn't matter. If you want, you can also create one for the notes topic.

```shell
gcloud pubsub subscriptions create vulns \
    --project $PROJECT_ID \
    --topic container-analysis-occurrences-v1
```

Now you will need to trigger the AA event:

> Simplest way to do that is to push an existing image using [crane](https://github.com/michaelsauter/crane).

```shell
crane cp $FROM_IMAGE $TO_IMAGE
```

Finally list the vulnerabilities that were discovered:

```shell
gcloud pubsub subscriptions pull vulns \
    --project $PROJECT_ID --auto-ack --limit 3 \
    --format="json(message.attributes, message.data.decode(\"base64\").decode(\"utf-8\"), message.messageId, message.publishTime)"
```

> If the above command returns an empty array (`[]`), give it a few seconds and rerun it. The length of the delay will depend on the size of your image and the number of vulnerabilities.

Each one of the vulnerabilities discovered by AA in your image will look something like this:

> Note, there will be other kinds of messages (e.g. DISCOVERY), you can ignore these.

```json
{
    "message": {
        "data": {
            "name": "projects/$PROJECT/occurrences/d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d",
            "kind": "VULNERABILITY", 
            "notificationTime": "2023-03-30T23:09:28.471565Z"
        },
        "messageId": "7309675999864387",
        "publishTime": "2023-03-30T23:09:28.592Z"
    }
}
```

Once you have the id of the occurrence, you can use the Container Analysis REST API to get the details:

```shell
curl -Ss -H "Content-Type: application/json; charset=utf-8" \
     -H "Authorization: Bearer $(gcloud auth application-default print-access-token)" \
     https://containeranalysis.googleapis.com/v1/projects/$PROJECT/occurrences/d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d
```

Alternatively, if you know the repo and image you can use `gcloud`:

```shell
gcloud artifacts docker images list $REGION-docker.pkg.dev/$PROJECT/$REPO/$IMAGE \
    --occurrence-filter "occurrenceId=\"d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d\"" \
    --format=json
```

Either way, the response will look something like this:

```json
{
    "name": "projects/$PROJECT/occurrences/d2342144-8a7e-4f3c-b3ba-87ebbe3ac72d",
    "resourceUri": "https://us-west1-docker.pkg.dev/$PROJECT/$REPO/$IMAGE@sha256:5ffd30269c7bde2e29453bb9b8d3618814b7034e37aef299e3c071acbb565911",
    "noteName": "projects/$PROJECT/notes/CVE-2019-7577",
    "kind": "VULNERABILITY",
    "createTime": "2023-03-30T23:09:28.443028Z",
    "updateTime": "2023-03-30T23:09:28.443028Z",
    "vulnerability": {
        "severity": "MEDIUM",
        "cvssScore": 6.8,
        "packageIssue": [
            {
                "affectedCpeUri": "cpe:/o:canonical:ubuntu_linux:18.04",
                "affectedPackage": "libsdl2",
                "affectedVersion": {
                    "name": "2.0.8+dfsg1",
                    "revision": "1ubuntu1.18.04.5~18.04.1",
                    "kind": "NORMAL",
                    "fullName": "2.0.8+dfsg1-1ubuntu1.18.04.5~18.04.1"
                },
                "fixedCpeUri": "cpe:/o:canonical:ubuntu_linux:18.04",
                "fixedPackage": "libsdl2",
                "fixedVersion": {
                    "kind": "MAXIMUM"
                },
                "packageType": "OS",
                "effectiveSeverity": "LOW"
            }
        ],
        "shortDescription": "CVE-2019-7577",
        "longDescription": "NIST vectors: AV:N/AC:M/Au:N/C:P/I:P/A:P",
        "relatedUrls": [
            {
                "url": "http://people.ubuntu.com/~ubuntu-security/cve/CVE-2019-7577",
                "label": "More Info"
            }
        ],
        "effectiveSeverity": "LOW",
        "cvssv3": {
            "baseScore": 8.8,
            "exploitabilityScore": 2.8,
            "impactScore": 5.9,
            "attackVector": "ATTACK_VECTOR_NETWORK",
            "attackComplexity": "ATTACK_COMPLEXITY_LOW",
            "privilegesRequired": "PRIVILEGES_REQUIRED_NONE",
            "userInteraction": "USER_INTERACTION_REQUIRED",
            "scope": "SCOPE_UNCHANGED",
            "confidentialityImpact": "IMPACT_HIGH",
            "integrityImpact": "IMPACT_HIGH",
            "availabilityImpact": "IMPACT_HIGH"
        }
    }
}
```


## programmatic

You can also instrument the entire process in Cloud Function. This example will dispatch each artifact vulnerability discovered by AA to a target sender (Slack channel, Jira API, custom API).

Start by creating a service account:

```shell
gcloud iam service-accounts create vuln-dispatcher
```

Next, grant that account the necessary roles (`roles/secretmanager.secretAccessor` to allow it to read secrets, and `roles/containeranalysis.occurrences.viewer` to allow it to access the AA API): 

```shell
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:vuln-dispatcher@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor" \
    --condition=None
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:vuln-dispatcher@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/containeranalysis.occurrences.viewer" \
    --condition=None 
```

Next, create a secret to store your secret:

```shell
gcloud secrets create vuln-dispatcher --replication-policy=automatic
echo -n "secret-config-value" | gcloud secrets versions add vuln-dispatcher --data-file=-
```

After that, capture the project number: 

```shell
export PROJECT_NUMBER="$(gcloud projects describe $PROJECT_ID --format='get(projectNumber)')"
```

Customize the included code to your needs. The [main entry point for the GCF function (fn.go)](./fn.go) has a variable which dictates what type of dispatcher will be used, set it to the one you need:

```go
var (
	// default occurrence sender
	sender OccurrenceSender = stdout.Sender

	// uncomment the one you prefer
	// sender OccurrenceSender = slack.Sender
	// sender OccurrenceSender = jira.Sender
	// sender OccurrenceSender = rest.Sender
)
```

Finally, deploy the function itself (in this example, with configuration for Jira).

> More details on the options available in `gcloud functions deploy` available [here](https://cloud.google.com/sdk/gcloud/reference/functions/deploy).

```shell
gcloud functions deploy vuln-dispatcher \
    --project=$PROJECT_ID \
    --region=$REGION \
    --runtime=go120 \
    --entry-point=Execute \
    --trigger-event=providers/cloud.pubsub/eventTypes/topic.publish \
    --trigger-resource=container-analysis-occurrences-v1 \
    --set-secrets="/secrets/dispatcher=projects/$PROJECT_NUMBER/secrets/vuln-dispatcher:latest" \
    --service-account="vuln-dispatcher@$PROJECT_ID.iam.gserviceaccount.com"
```

Now whenever Container Analyses finds new vulnerability, the image reference URI, and the CVE along with few other metadata bits will be published to your target system.

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.
