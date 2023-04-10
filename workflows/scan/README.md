# artifact-events - scan new GCR or AR image

The Vulnerability scanning provided by [Container Scanning API](https://console.cloud.google.com/marketplace/product/google/containerscanning.googleapis.com) provides pretty good base image coverage (Alpine, CentOS, Debian, RedHat, Ubuntu etc.). The [coverage is expending](https://cloud.google.com/container-analysis/docs/os-overview) but at this time, this service only covers Go and Java languages in user space of the artifact (i.e. apps and dependencies included by user on top of the base image).

This is not a big deal as there are a number of free vulnerability scanners available in the open source community (e.g. [grype](https://github.com/anchore/grype), [snyk](https://github.com/snyk/cli), or [trivy](https://github.com/aquasecurity/trivy)), so you can use the events that both GCR and AR provide to wire up these scanners to process all new images published to your registry. 

In this example, we will use `grype` and `trivy` scanners and save the discovered vulnerabilities into GCS bucket for comparison. 

> For ease of reproducibility, this demo skips snyk as it requires API token. See [this repo](https://github.com/mchmarny/vimp/tree/main/cloud/gcp) for fully functional example of how to use Secret Manger to securely pass token to Cloud Build job.

## setup 

First, [fork this repo](https://github.com/mchmarny/artifact-events/fork), so you can make changes, clone it locally, and navigate into the scan workflow directory: 

```shell
git clone https://github.com/$GITHUB_USER/artifact-events.git
cd artifact-events/workflows/scan
```

### github connection

Next, make sure you already have GitHub connection:

> If not, [navigate to the Triggers page](https://console.cloud.google.com/cloud-build/triggers) in Cloud Build as this step will need to be created in UI

```shell
gcloud alpha builds connections list \
    --project=$PROJECT_ID \
    --region $REGION
```

### gcs bucket 

To store the results of scans, create a GCS bucket: 

```shell
gcloud storage buckets create gs://scanner-data-$PROJECT_ID
```

### pub/sub topic

Try to create the `gcr` topic:

> Note, that topic may already exists, so if you see `Resource already exists in the project` error just move on to the next step: 

```shell
gcloud pubsub topics create gcr --project=$PROJECT_ID
```

### gcb trigger

Next, create a pub/sub trigger using the provided [build configurations file](./scan.yaml). More detail about the parameters used below [here](https://cloud.google.com/build/docs/automate-builds-pubsub-events):

```shell
gcloud alpha builds triggers create pubsub \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=scan-new-registry-image \
    --topic=projects/$PROJECT_ID/topics/gcr \
    --build-config=workflows/scan/scan.yaml \
    --substitutions=_DIGEST='$(body.message.data.digest)',_ACTION='$(body.message.data.action)',_BUCKET=$BUCKET \
    --subscription-filter='_ACTION == "INSERT"' \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

Cloud Build will automatically extract the payload (base64 encoded in the message), so in our workflow we can references the raw value (`body.message.data`). Using GCB substitutions then we can create the key environment variables. `_ACTION` is only used for filtering the appropriate messages, while the `_DIGEST` variable will have the fully-qualified URI of the image,including digest.

## test

To test, push any new image to any registry in the same project: 

> Simplest way to do that is to copy an existing image using `crane` but you can also build on from scratch. Make sure to substitute your images.

```shell
crane cp \
    $REGION-docker.pkg.dev/$PROJECT_ID/$FROM_REPO/$IMAGE \
    $REGION-docker.pkg.dev/$PROJECT_ID/$TO_REPO/$IMAGE
```

AR in this case will publish an event onto the `gcr` topic with the fully-qualified URI of the image (including digest). 

```json
{
    "message": {
        "data": {
            "action": "INSERT", 
            "digest": "us-west1-docker.pkg.dev/$PROJECT/repo/image@sha256:54bc0fead59f304f1727280c3b520aeea7b9e6fd405b7a6ee1dddc8d78044516", 
            "tag": "us-west1-docker.pkg.dev/$PROJECT/repo/image:latest"
        },
        "messageId": "7309198396944430",
        "publishTime": "2023-03-30T21:56:52.254Z"
    }
}
```

At this point you should see the new reports being saved in your GCE bucket. The names of each report will be a combination of the image sha + the scanner type (e.g. `54bc0fead59f304f1727280c3b520aeea7b9e6fd405b7a6ee1dddc8d78044516-trivy.json`).

## on-demand scanning

You can also use the exact same workflow to scan images on-demand using GCB WebHooks. First, create a trigger: 

```shell
gcloud alpha builds triggers create manual \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=scan-manual-registry-image \
    --build-config=workflows/scan/scan.yaml \
    --substitutions=_DIGEST='$(body.image)' \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

Next, capture the trigger ID: 

```shell
export TRIGGER_ID=$(gcloud beta builds triggers describe \
    scan-manual-registry-image \
    --project=$PROJECT_ID \
    --region=$REGION \
    --format='value(id)')
```

You can run this trigger now manually, by invoking from `curl`. 

> This assumes that you have the necessary role to execute the build.

```shell
curl -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" \
     "https://cloudbuild.googleapis.com/v1/projects/$PROJECT_ID/locations/$REGION/triggers/$TRIGGER_ID:run"
```

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.
