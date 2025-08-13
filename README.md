# UserInsightGo

if google auth fail, then run

```shell
gcloud auth application-default login

```

1.  mixpanel上下载最新的users.csv放到sources/mixpanel/users.csv：
    - https://mixpanel.com/project/3317175/view/3821831/app/users?mp_source=intro-complete#ixZ5pvp63CYU
2.  bigquery上获取最近三个月使用过的用户并下载到sources/big_query/users.csv -

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

3.  运行服务

```shell
go run main.go
```

## Clound Run

1. 准备好.env
2. `make auth`登录google镜像仓库
3. `make image`build镜像
4. `make push`推送镜像
