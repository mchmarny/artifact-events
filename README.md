# artifact-events

Collection of Google Container and Artifact Registry workflows 

## New GCR or AR images 

First, make sure there are GitHub connections (if not, this step will need to be created in UI):

```shell
gcloud alpha builds connections list --region $REGION
```

Next, create a pub/sub trigger using the [provided  build configurations file](scan-new-image.yaml). More detail about the parameters used below [here](https://cloud.google.com/build/docs/automate-builds-pubsub-events):

```shell
gcloud alpha builds triggers create pubsub \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=scan-registry-image \
    --topic=projects/$PROJECT_ID/topics/gcr \
    --build-config=scan.yaml \
    --substitutions=_DIGEST='$(body.message.data.digest)',_ACTION='$(body.message.data.action)',_SNYK_TOKEN=$SNYK_TOKEN,_BUCKET=$BUCKET \
    --subscription-filter='_ACTION == "INSERT"' \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

You will also have to create Pub/Sub topic where the resulting events will be propagated: 

```shell
gcloud pubsub topics create image-scans --project $PROJECT_ID
```

Now push any new image to any registry in the same project: 

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

Cloud Build will automatically extract the payload (base64 encoded in the message), so in our workflow we can references the raw value (`body.message.data`). Using GCB substitutions then we can create the key environment variables. `_ACTION` is only used for filtering the appropriate messages, while the `_DIGEST` variable will have the fully-qualified URI of the image,including digest.

Using the digest, the example [scan-new-image.yaml](scan-new-image.yaml) workflow then scans the image for vulnerabilities using three open source scanners: `grype`, `snyk`, and `trivy` and saves the resulting reports to GCS bucket.

## Report Processing

The above section showed how to scan new images using OSS vulnerability scanner and save them to GCS bucket. That step also publishes the data about the image and the resulting file onto a Pub/Sub topic. In this section we will cover the processing of these events.

These events will look somethings like this:

```json
{
    "message": {
        "attributes": {
            "file": "gs://artifact-events/14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55-grype.json",
            "format": "grype"
        },
        "data": "us-west1-docker.pkg.dev/cloudy-demos/events/test33@sha256:14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55",
        "messageId": "7329185432545049",
        "publishTime": "2023-04-02T12:22:08.464Z"
    }
}
```

If you want, you can look at these events by creating a pub/sub subscription to access these events:

```shell
gcloud pubsub subscriptions create image-scans-sub --project $PROJECT_ID --topic image-scans
```

And then printing the content of these events like this:

```shell
gcloud pubsub subscriptions pull image-scans-sub --project $PROJECT_ID --auto-ack --limit 3 \
    --format="json(message.attributes, message.data.decode(\"base64\").decode(\"utf-8\"), message.messageId, message.publishTime)"
```

GCB though will create its own subscription when we create trigger to process these events:

```shell
gcloud alpha builds triggers create pubsub \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=process-image \
    --topic=projects/$PROJECT_ID/topics/image-scans \
    --build-config=process.yaml \
    --substitutions=_DIGEST='$(body.message.data)',_REPORT='$(body.message.attributes.file)',_FORMAT='$(body.message.attributes.format)',_BUCKET=$BUCKET,_DATASET=$DATASET \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

To test the flow you can publish to the topic directly:

```shell
gcloud pubsub topics publish image-scans --message="us-west1-docker.pkg.dev/cloudy-demos/events/test38@sha256:14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55" --attribute="file=gs://artifact-events/14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55-snyk.json,format=snyk" --project=$PROJECT_ID
```

When completed, the data will be loaded into the BigQuery dataset set in the trigger (e.g. `dataset.table`).

## External Images 

In addition to the images from GCR and AR, you can also process other images as long as they are either public or accessible to the CGB service account. 

Start by creating the image queue topic: 

```shell
gcloud pubsub topics create image-queue --project $PROJECT_ID
```

Next create a trigger to process any new events on that queue with the same build config as above: 

```shell
gcloud alpha builds triggers create pubsub \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=scan-queue-image \
    --topic=projects/$PROJECT_ID/topics/image-queue \
    --build-config=scan.yaml \
    --substitutions=_DIGEST='$(body.message.data)',_SNYK_TOKEN=$SNYK_TOKEN,_BUCKET=$BUCKET \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

Now to process new image simply publish an image URI to that topic:

```shell
gcloud pubsub topics publish image-queue \
    --message=us-west1-docker.pkg.dev/cloudy-demos/events/test38@sha256:14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55 \
    --project=$PROJECT_ID
```

## On Demand

You can also process images on-demand using GCB WebHooks. First, create a trigger: 

```shell
gcloud alpha builds triggers create webhook \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=queue-image \
    --build-config=queue.yaml \
    --substitutions=_IMAGE='$(body.message.image)' \
    --secret=projects/799736955886/secrets/artifact-event/versions/2 \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

> make sure the secret is accessible to `service-$PROJECT_NUMBER@gcp-sa-cloudbuild.iam.gserviceaccount.com`

You can run this manually in the console or by invoking the webhook. 

```shell
curl -H "Content-Type: application/json" \
     -d '{"image":"redis"}' \
     "https://cloudbuild.googleapis.com/v1/projects/$PROJECT_ID/triggers/queue-image:webhook?key=$KEY"
```


## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

