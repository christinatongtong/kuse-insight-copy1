## Cloud Run Job for UserInsight

prepare `.env` file

```shell
OPENAI_API_KEY=
MIXPANEL_TOKEN=
PINECONE_API_KEY=
PINECONE_NAMESPACE=documents
PINECONE_INDEX_HOST=documents-dso1lmi.svc.gcp-us-central1-4a9f.pinecone.io
```

## Local Test

1. Prepare

- bigquery read permission of kuse_ai project: https://console.cloud.google.com/bigquery?inv=1&invt=Ab1lGQ&project=kuse-ai&supportedpurview=project&ws=!1m10!1m4!4m3!1skuse-ai!2sinsight!3suser_predict!1m4!4m3!1skuse-ai!2smixpanel_analytics!3smp_people_data_view

2. run main.py

```shell
python main.py
```

if google bigquery auth fail, then run

```shell
gcloud auth application-default login

```

## Deploy

1. Google Auth

```
make auth
```

2. build & push image

```shell
make ENV=prd VS=1.0
```

3. create or edit Cloud Run job

- https://console.cloud.google.com/run/jobs/details/us-east4/user-insight-job/executions?inv=1&invt=Ab1pRg&project=kuse-ai&supportedpurview=project
- modify the image to new image
- Execute
