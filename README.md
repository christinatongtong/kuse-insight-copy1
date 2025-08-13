# UserInsightGo

Local Run
```shell
gcloud auth application-default login

```

1.  downlaod the the most recent users.csv from mixpanel to sources/mixpanel/users.csv：
    - https://mixpanel.com/project/3317175/view/3821831/app/users?mp_source=intro-complete#ixZ5pvp63CYU
2.  fetch user data from the past three months from bigquery, download it to sources/big_query/users.csv -

    - https://console.cloud.google.com/bigquery?inv=1&invt=Abzu-g&project=kuse-ai&supportedpurview=project&ws=!1m5!1m4!1m3!1skuse-ai!2sbquxjob_20e1ffdd_1975e3fa3b7!3sUS

    ```shell
    SELECT \* FROM EXTERNAL_QUERY(
    "kuse-ai.us.kuse-ai-main",
    """
    SELECT DISTINCT user_id
    FROM tasks
    WHERE created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 3 MONTH) AND task_type = 'communication';
    """
    )
    ```

To Run:

```shell
go run main.go
```

## Clound Run

refer to /jobs
