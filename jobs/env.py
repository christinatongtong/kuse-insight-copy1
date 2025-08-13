from dotenv import load_dotenv
import os


class Settings(object):
    def __init__(self):
        self.openai_api_key = os.getenv("OPENAI_API_KEY")
        self.mixpanel_token = os.getenv("MIXPANEL_TOKEN")
        self.pinecone_api_key = os.getenv("PINECONE_API_KEY")
        self.pinecone_namespace = os.getenv("PINECONE_NAMESPACE")
        self.pinecone_index_host = os.getenv("PINECONE_INDEX_HOST")
        self.is_test: bool = os.getenv("ENVIRONMENT") != "cloud"  # type: ignore
        self.predict_confidence_threshold: float = 0.6
        self.min_task_count: int = 10

    @property
    def version(self) -> int:
        from datetime import datetime, time

        midnight = datetime.combine(datetime.now().date(), time.min)

        return int(midnight.timestamp())


def version(self) -> int:
    from datetime import datetime, time, timedelta

    now = datetime.now()
    # 获取本周一的日期（周一为0，周日为6）
    monday = now - timedelta(days=now.weekday())
    # 拼接凌晨时间
    monday_midnight = datetime.combine(monday.date(), time.min)

    return int(monday_midnight.timestamp())


load_dotenv()
settings = Settings()


class Logger(object):
    def __init__(self):
        self.version = settings.version
        if self.is_test:
            return
        import google.cloud.logging

        client = google.cloud.logging.Client()
        client.setup_logging()
        self._logger = client.logger("user-insight-job")

    def _print(self, msg: str) -> None:
        if not self.is_test and "DEBUG" not in msg:
            self._logger.log_text(msg)
        else:
            print(msg)

    def info(self, msg: str):
        self._print(f"[INFO][{self.version}] " + msg)

    def error(self, msg: str):
        self._print(f"[ERROR][{self.version}] " + msg)

    def warn(self, msg: str):
        self._print(f"[WARN][{self.version}] " + msg)

    def debug(self, msg: str):
        if not self.is_test:
            return
        print(f"[DEBUG][{self.version}] " + msg)

    @property
    def is_test(self) -> bool:
        return settings.is_test


logger = Logger()

if __name__ == "__main__":
    print(settings.version)
