# artifact-events

Collection of Google Container and Artifact Registry workflows 

## New GCR or AR image 

First, make sure there are GitHub connections (if not, this step will need to be created in UI):

```shell
gcloud alpha builds connections list --region $REGION
```

Next, create a pub/sub trigger using the [provided  build configurations file](scan-new-image.yaml). More detail about the parameters used below [here](https://cloud.google.com/build/docs/automate-builds-pubsub-events):

```shell
gcloud alpha builds triggers create pubsub \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=scan-image \
    --topic=projects/$PROJECT_ID/topics/gcr \
    --build-config=scan.yaml \
    --substitutions=_DIGEST='$(body.message.data.digest)',_ACTION='$(body.message.data.action)',_SNYK_TOKEN=$SNYK_TOKEN,_BUCKET=$BUCKET \
    --subscription-filter='_ACTION == "INSERT"' \
    --repo=https://www.github.com/$GITHUB_USER/artifact-events \
    --repo-type=GITHUB \
    --branch=main
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

## Processing 

The above section showed how to scan new images using OSS vulnerability scanner and save them to GCS bucket. In this section we will cover the processing of these files. 

Start by creating Pub/Sub topic where the above event will be propagated: 

```shell
gcloud pubsub topics create image-scans --project $PROJECT_ID
```

Now when the next image is published you can create a subscription to access these events:

```shell
gcloud pubsub subscriptions create image-scans-sub --project $PROJECT_ID --topic image-scans
```

To print the content of the events published by scanning step:

```shell
gcloud pubsub subscriptions pull image-scans-sub --project $PROJECT_ID --auto-ack --limit 3 \
    --format="json(message.attributes, message.data.decode(\"base64\").decode(\"utf-8\"), message.messageId, message.publishTime)"
```

The results should look somethings like this:

```json
{
    "message": {
        "attributes": {
            "digest": "us-west1-docker.pkg.dev/cloudy-demos/events/test32@sha256:14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55",
            "format": "snyk"
        },
        "data": "gs://artifact-events/14dd03939d2d840d7375f394b45d340d95fba8e25070612ac2883eacd7f93a55-snyk.json",
        "messageId": "7322343576783078",
        "publishTime": "2023-04-01T12:05:22.125Z"
    }
}
```

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

