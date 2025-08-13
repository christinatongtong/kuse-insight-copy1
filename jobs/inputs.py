from google.cloud import bigquery
from typing import List, Dict, Optional
from env import settings, logger
from schemas import UserModel, UserProperty
import json
from pinecone import Pinecone as PineconeClient


class BigQuery(object):
    def __init__(self):
        self._client = bigquery.Client()

    @property
    def project_id(self) -> str:
        return "kuse-ai"

    def kuse_ai_table(self) -> str:
        return f"{self.project_id}.us.kuse-ai-main"

    def mixpanel_table(self) -> str:
        return f"{self.project_id}.mixpanel_analytics.mp_people_data_view"

    def user_insight_table(self) -> str:
        return f"{self.project_id}.insight.user_predict"

    def _invoke(self, user_id: int, query: str, tag=""):
        try:
            job = self._client.query(query)
            return job.result()
        except Exception as err:
            logger.error(f"bigquery.{tag or 'invoke'} user_id: {user_id}, err: {err}")
        return []

    def load_user_ids(self) -> List[int]:
        """
        获取需要分析的user_id列表, 满足以下条件:
        - 最近一周使用过magic_cast
        - 总的使用次数 > 10
        """
        query = f"""
        SELECT * FROM EXTERNAL_QUERY(
          "{self.kuse_ai_table()}",
          \"""
        SELECT DISTINCT recent.user_id
        FROM (
          SELECT user_id
          FROM tasks
          WHERE created_at >= DATE_SUB(CURRENT_DATE, INTERVAL 1 WEEK)
            AND task_type = 'communication'
        ) AS recent
        JOIN (
          SELECT user_id
          FROM tasks
          WHERE task_type = 'communication'
          GROUP BY user_id
          HAVING COUNT(*) > {settings.min_task_count}
        ) AS all_time
        ON recent.user_id = all_time.user_id
          \"""
        )
        """

        results = self._invoke(user_id=0, query=query, tag="load_user_ids")

        user_ids = []
        for row in results:
            user_id = row.get("user_id", "")
            if not user_id:
                continue
            user_ids.append(user_id)
        return user_ids

    # 根据user_id和version作为联合索引去更新数据
    def upsert_user_predict(self, version: int, row_data: Dict[str, str | int]):
        """
        把预测分析的结果加上版本保存到bigquery里备份
        """
        row_data["version"] = version
        # 构建 SELECT 部分
        select_clause = ", ".join(
            [
                f"{repr(v)} AS {k}" if isinstance(v, str) else f"{v} AS {k}"
                for k, v in row_data.items()
            ]
        )

        # 构建 SET 子句（排除主键）
        set_clause = ", ".join(
            [f"{k} = S.{k}" for k in row_data if k not in ["user_id", "version"]]
        )

        # 构建 INSERT 子句
        columns = ", ".join(row_data.keys())
        values = ", ".join([f"S.{k}" for k in row_data.keys()])

        query = f"""
        MERGE `{self.user_insight_table()}` T
        USING (SELECT {select_clause}) S
        ON T.user_id = S.user_id AND T.version = S.version
        WHEN MATCHED THEN
          UPDATE SET {set_clause}
        WHEN NOT MATCHED THEN
          INSERT ({columns}) VALUES ({values})
        """

        self._invoke(user_id=0, query=query, tag="upsert_user_predict")

    def load_user_profile(self, user_id: int) -> Optional[UserModel]:
        """
        通过bigquery从kuse_ai项目的mysql数据库user表里查询用户信息
        """
        query = f"""
        SELECT * FROM EXTERNAL_QUERY(
          "{self.kuse_ai_table()}",
          \"""
            SELECT
              u.id,
              u.email,
              u.given_name,
              u.family_name,
              u.image_url,
              u.output_language,
              u.full_name
            FROM user AS u
            WHERE id = {user_id}
          \"""
        )
        """

        results = self._invoke(user_id=user_id, query=query, tag="load_user_profile")
        for row in results:
            model = UserModel(user_id=row[0])
            model.__dict__.update(
                {
                    "email": row[1],
                    "given_name": row[2],
                    "family_name": row[3],
                    "image_url": row[4],
                    "output_language": row[5],
                    "full_name": row[6],
                }
            )
            return model

        return None

    def load_user_prompts(self, user_id: int) -> List[str]:
        """
        通过bigquery从kuse_ai项目的mysql数据库tasks表里查询用户用过的prompt
        """
        query = f"""
        SELECT * FROM EXTERNAL_QUERY(
          "{self.kuse_ai_table()}",
          \"""
           SELECT
            u.id,
            CONVERT(task_meta USING utf8) AS task_meta
          FROM tasks AS u
          WHERE u.task_type = 'communication'
          AND u.user_id = {user_id}
          \"""
        )
        """

        results = self._invoke(user_id=user_id, query=query, tag="load_user_prompts")
        prompts: List[str] = list()
        for row in results:
            try:
                d = json.loads(row[1])

                prompt = d.get("prompt", "")
                if not prompt:
                    continue
                prompts.append(prompt)
            except Exception as err:
                logger.error(
                    f"load_user_prompts.err {err}, row: {row}, user_id: {user_id}"
                )
                continue

        return prompts

    def load_user_filenames(self, user_id: int) -> List[str]:
        """
        通过bigquery从kuse_ai项目的mysql数据库files表里查询用户上传过的文件名
        """
        query = f"""
        SELECT * FROM EXTERNAL_QUERY(
          "{self.kuse_ai_table()}",
          \"""
          SELECT
            u.id,
	    u.filename
          FROM files AS u
          WHERE u.id = {user_id}
          \"""
        )
        """

        results = self._invoke(user_id=user_id, query=query, tag="load_user_filenames")
        file_names: List[str] = list()
        for row in results:
            file_name = row[1]
            if not file_name:
                continue
            file_names.append(file_name)

        return file_names

    def load_user_from_mixpanel(self, user_id: int) -> Optional[UserProperty]:
        """
        通过bigquery从mixpanel获取用户的其他个人信息
        """
        query = f"""
        SELECT * FROM `{self.mixpanel_table()}` WHERE distinct_id = "{user_id}" LIMIT 1
        """

        results = self._invoke(
            user_id=user_id, query=query, tag="load_user_from_mixpanel"
        )
        try:
            for row in results:
                user = UserProperty(user_id=user_id)
                user.load_from_mixpanel(row[0])
                return user
        except Exception as err:
            logger.error(f"load_user_from_mixpanel user_id: {user_id}, err:{err}")
        return None


class Pinecone(object):
    def __init__(self):
        self._client = PineconeClient(
            api_key=settings.pinecone_api_key, source_tag="Insight"
        )
        self._index = self._client.Index(host=settings.pinecone_index_host)

    def search_user_file_summary(self, user_id: int) -> List[str]:
        """
        通过pinecont获取用户上传过的文件的summary
        """
        summaries: List[str] = list()
        try:
            query = {
                "user_id": {"$eq": user_id},
                "is_summary": {
                    "$eq": 1,
                },
            }
            result = self._index.query(
                namespace=settings.pinecone_namespace,
                top_k=10,
                include_metadata=True,
                filter=query,
                vector=[0 for _ in range(1536)],
            )
            for doc in result.get("matches", []):
                summary = doc.get("metadata", {}).get("text", "")
                if not summary:
                    continue
                summaries.append(summary)
        except Exception as err:
            logger.error(
                f"pinecone.search_user_file_summary user_id: {user_id}, err: {err}"
            )

        return summaries


bq = BigQuery()
pc = Pinecone()

if __name__ == "__main__":
    # print(bq.load_user_ids())

    # bq.upsert_user_predict(
    #     0,
    #     {
    #         "user_id": 1,
    #         "version": 2,
    #         "occupation": "software developer",
    #         "industry": "AI",
    #         "school": "xxxxxx",
    #         "primary_language": "traditional chinese",
    #         "major": "compute science",
    #     },
    # )

    # user = bq.load_user_profile(user_id=64011)
    # if user is not None:
    #     print(user.__dict__)

    # print(bq.load_user_prompts(user_id=64011))
    # print(bq.load_user_filenames(user_id=285093))
    # print(pc.search_user_file_summary(user_id=55308))
    user_property = bq.load_user_from_mixpanel(user_id=55308)
    if user_property:
        print(f"xxxxxx {user_property.__dict__}")
    else:
        print("xx not found")
