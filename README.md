# artifact-events

Collection of Google Container and Artifact Registry workflows 

## New GCR or AR image 

First, make sure there are GitHub connections (connections need to be created in UI):

```shell
gcloud alpha builds connections list --region $REGION
```

Create trigger: 

```shell
gcloud alpha builds triggers create pubsub \
    --project=$PROJECT_ID \
    --region=$REGION \
    --name=scan-image \
    --topic=projects/$PROJECT_ID/topics/gcr \
    --build-config=scan-new-image.yaml \
    --substitutions=_DIGEST='$(body.message.data.digest)',_ACTION='$(body.message.data.action)',_SNYK_TOKEN=$SNYK_TOKEN,_BUCKET=$BUCKET \
    --subscription-filter='_ACTION == "INSERT"' \
    --repo=https://www.github.com/mchmarny/artifact-events \
    --repo-type=GITHUB \
    --branch=main
```

Next push a new image to any registry in that project: 

> Make sure to substitute your images.

```shell
crane cp \
    $REGION-docker.pkg.dev/$PROJECT_ID/$FROM_REPO/$IMAGE \
    $REGION-docker.pkg.dev/$PROJECT_ID/$TO_REPO/$IMAGE
```
