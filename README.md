# UserInsightGoRibin

Local Run
```shell
gcloud auth application-default login

```

1.  downlaod the the most recent users.csv from mixpanel to sources/mixpanel/users.csv：
2.  fetch user data from the past three months from bigquery, download it to sources/big_query/users.csv -


"""
SELECT DISTINCT user_id
FROM tasks
WHERE created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 3 MONTH) AND task_type = 'communication';
"""
```

To Run:

```shell
go run main.go
```

To Clound Run, refer to /jobs
