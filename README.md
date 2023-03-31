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
    --build-config=scan-new-image.yaml \
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

I hope this example gives you an idea how you can extend the image processing capabilities already available on GCP. 

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

