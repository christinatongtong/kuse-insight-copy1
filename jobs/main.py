from env import settings, logger
from inputs import bq, pc
from typing import List, Optional, Dict
from mixpanel import Mixpanel, Consumer
from schemas import UserPredict
from prompts import USER_INSIGHT_SYSTEM_PROMPT, format_user_prompt, USER_AVATAR_PROMPT
from openai import OpenAI
import json
import time


class UserInsight(object):
    def __init__(self):
        self.is_test: bool = settings.is_test
        self._mixpanel: Mixpanel = Mixpanel(
            token=settings.mixpanel_token, consumer=Consumer(retry_limit=2)
        )
        self._llm = OpenAI(api_key=settings.openai_api_key)
        self.version = settings.version
        logger.info(f"init finished, is_test: {self.is_test}, version: {self.version}")

    def run(self, user_ids: List[int]):
        if not user_ids or len(user_ids) == 0:
            user_ids = bq.load_user_ids()
        count = len(user_ids)
        logger.info(f"start predict job... count: {count}")
        for index in range(count):
            start = int(time.time())
            user_predict = self.predcit(user_id=user_ids[index])
            if not user_predict:
                logger.info(
                    f"user_insight.skip [{index + 1}/{count}] {user_ids[index]}"
                )
                continue
            self.update_predict(user_predict=user_predict)
            logger.info(
                f"user_insight.predict [{index + 1}/{count}] {user_predict.row_data()}, cost: {int(time.time()) - start}"
            )

    def predcit(self, user_id: int) -> Optional[UserPredict]:
        user_profile = bq.load_user_profile(user_id=user_id)
        if not user_profile:
            logger.warn(f"predict.load_user_profile.not_found user_id: {user_id}")
            return None

        task_prompts = bq.load_user_prompts(user_id=user_id)
        if len(task_prompts) <= settings.min_task_count:
            return None
        filenames = bq.load_user_filenames(user_id=user_id)
        summaries = pc.search_user_file_summary(user_id=user_id)

        user_property = bq.load_user_from_mixpanel(user_id=user_id)
        image_description = self.describe_image(user_profile.image_url)

        prompt = format_user_prompt(
            user_profile=user_profile,
            filenames=filenames,
            task_prompts=task_prompts,
            summaries=summaries,
            image_description=image_description,
            user_property=user_property,
        )
        reply = self._call_llm(prompt=prompt)
        if not reply:
            return None
        result = {}
        try:
            result = json.loads(reply)
        except Exception as err:
            logger.error(
                f"predict.unmarshal user_id:{user_id}, reply: {reply}, err: {err}"
            )
            return None

        user_predict = UserPredict(user_id=user_id)
        user_predict.load_from_data(result)
        return user_predict

    def _call_llm(self, prompt: str) -> str:
        try:
            response = self._llm.responses.create(
                model="gpt-4.1",
                instructions=USER_INSIGHT_SYSTEM_PROMPT,
                input=prompt,
            )
            return response.output_text
        except Exception as err:
            logger.error(f"call_llm err: {err}")
        return ""

    def describe_image(self, image_url: str) -> str:
        """
        given image url, describe the image
        """
        if not image_url:
            return ""

        try:
            response = self._llm.chat.completions.create(
                model="gpt-4.1",
                messages=[
                    {
                        "role": "user",
                        "content": [
                            {"type": "image_url", "image_url": {"url": image_url}},
                            {"type": "text", "text": USER_AVATAR_PROMPT},
                        ],
                    }
                ],
            )
            return response.choices[0].message.content
        except Exception as e:
            logger.error(f"describe_image url: {image_url} err: {str(e)}")
            return ""

    def update_predict(self, user_predict: UserPredict):
        """
        同步最新的分析结果到外部
        """
        if self.is_test:
            return
        self._update_to_mixpanel(user_predict=user_predict)
        self._update_to_bigquery(user_predict=user_predict)

    def _update_to_bigquery(self, user_predict: UserPredict):
        bq.upsert_user_predict(version=self.version, row_data=user_predict.row_data())

    def _update_to_mixpanel(self, user_predict: UserPredict):
        self._mixpanel.people_set(user_predict.user_id, user_predict.properties())


if __name__ == "__main__":
    insight = UserInsight()
    insight.run(user_ids=[])
