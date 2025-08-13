from typing import Dict, List, Optional
from env import settings


class UserProperty(object):
    def __init__(self, user_id: int):
        self.user_id: int = user_id
        self.city: str = ""
        self.country_code: str = ""
        self.is_education: bool = False
        self.name: str = ""
        self.region: str = ""

    def load_from_mixpanel(self, d: dict):
        self.city = d.get("$city", "")
        self.country_code = d.get("$country_code", "")
        self.is_education = d.get("$is_student", False)
        self.name = d.get("$name", "")
        self.region = d.get("$region", "")

    def get_user_region(self) -> str:
        if self.city not in ["", "undefined"]:
            return self.city

        if self.region not in ["", "undefined"]:
            return self.region

        if self.country_code not in ["", "undefined"]:
            return self.country_code

        return "unknown"


class Candidate(object):
    def __init__(self):
        self.value: Optional[str] = ""
        self.evidence: str = ""
        self.confidence: float = 0.0

    def load_from_data(self, d: dict):
        self.__dict__.update(d)


class Candidates(object):
    def __init__(self):
        self.candidates: List[Candidate] = []

    def load_from_data(self, d):
        for v in d.get("candidates", []):
            candidate = Candidate()
            candidate.load_from_data(v)
            self.candidates.append(candidate)

    def pick(self) -> str:
        max_prob = 0.0
        max_result = "unknown"
        for candidate in self.candidates:
            if candidate.confidence < settings.predict_confidence_threshold:  # edge
                continue
            if candidate.confidence < max_prob:
                continue
            max_prob = candidate.confidence
            max_result = candidate.value if candidate.value else "unknown"
        return max_result.lower()


class UserPredict(object):
    def __init__(self, user_id: int):
        self.user_id = user_id
        self.occupation = Candidates()
        self.industry = Candidates()
        self.school = Candidates()
        self.primary_language = Candidates()
        self.major = Candidates()
        self.degree_level = Candidates()
        self.gender = Candidates()

    def load_from_data(self, d: dict):
        self.occupation.load_from_data(d.get("occupation", {}))
        self.industry.load_from_data(d.get("industry", {}))
        self.school.load_from_data(d.get("school", {}))
        self.primary_language.load_from_data(d.get("primary_language", {}))
        self.major.load_from_data(d.get("major", {}))
        self.degree_level.load_from_data(d.get("degree_level", {}))
        self.gender.load_from_data(d.get("gender", {}))

    def row_data(self) -> dict:
        """
        获取分析结果的dict形式
        """
        return {
            "user_id": self.user_id,
            "occupation": self.occupation.pick(),
            "industry": self.industry.pick(),
            "school": self.school.pick(),
            "primary_language": self.primary_language.pick(),
            "major": self.major.pick(),
            "degree_level": self.degree_level.pick(),
            "gender": self.gender.pick(),
        }

    def properties(self) -> dict:
        """
        获取需要上传到mixpanel上的数据
        """
        properties = dict()
        for key, value in self.row_data().items():
            if key == "user_id":
                continue
            value = value.lower()
            if value == "zh-cn":
                value = "simplified chinese"
            elif value == "zh-tw":
                value = "traditional chinese"

            properties[f"predict_{key}"] = value

        return properties


class UserModel(object):
    def __init__(self, user_id: int):
        self.user_id: int = user_id
        self.email: str = ""
        self.given_name: str = ""
        self.family_name: str = ""
        self.image_url: str = ""
        self.output_language: str = ""
        self.full_name: str = ""


if __name__ == "__main__":
    # ret = UserPredict(user_id=123)
    # ret.load_from_dict(
    #     {
    #         "predict_occupation": "software developer",
    #         "predict_industry": "AI",
    #         "predict_school": "nju",
    #         "predict_primary_language": "traditional chinese",
    #         "predict_major": "",
    #     }
    # )
    # print(ret.school)
    # print(ret.properties())
    result = {
        "primary_language": {
            "candidates": [
                {
                    "value": "Simplified Chinese",
                    "confidence": 0.6,
                    "evidence": "User used both English and Chinese, primarily issued commands in Chinese (e.g., '帮我把结果整理成一个Chart', '帮我生成一张图片', '你是谁'), and explicitly requested replies in Traditional Chinese for some specific tasks (e.g., 'reply use tranditional chinese language'), showing greater comfort with Simplified Chinese overall.",
                },
                {
                    "value": "Traditional Chinese",
                    "confidence": 0.3,
                    "evidence": "User explicitly requested replies in Traditional Chinese for specific questions but defaulted to Simplified Chinese otherwise.",
                },
                {
                    "value": "English",
                    "confidence": 0.1,
                    "evidence": "Some prompts in English (e.g., 'Hello', 'Who are you?'), but most functional requests and detailed instructions are in Chinese.",
                },
            ]
        },
        "gender": {
            "candidates": [
                {
                    "value": None,
                    "confidence": 1.0,
                    "evidence": "No gender-indicative data in the name, email, or user prompts.",
                }
            ]
        },
        "school": {
            "candidates": [
                {
                    "value": None,
                    "confidence": 1.0,
                    "evidence": "No data available regarding user's educational institution.",
                }
            ]
        },
        "major": {
            "candidates": [
                {
                    "value": None,
                    "confidence": 1.0,
                    "evidence": "No data available regarding user's major or field of study.",
                }
            ]
        },
        "degree_level": {
            "candidates": [
                {
                    "value": None,
                    "confidence": 1.0,
                    "evidence": "No data available regarding user's degree level.",
                }
            ]
        },
        "industry": {
            "candidates": [
                {
                    "value": "Technology & Software",
                    "confidence": 0.5,
                    "evidence": "User asks about SEO, image generation, data summary, and visualization—skills commonly aligned with technology/software domain.",
                },
                {
                    "value": "Other",
                    "confidence": 0.2,
                    "evidence": "Prompts cover a range of topics, though mostly technical in nature.",
                },
                {
                    "value": "Education",
                    "confidence": 0.2,
                    "evidence": "Asks for summaries and explanations that could align with research or educational use.",
                },
                {
                    "value": "Finance & Business Services",
                    "confidence": 0.1,
                    "evidence": "Asked about GDP data analysis, which relates to finance, but overall context leans more technical.",
                },
            ]
        },
        "occupation": {
            "candidates": [
                {
                    "value": "Data Analysis",
                    "confidence": 0.5,
                    "evidence": "User asks to summarize files, format results into charts/excels, and analyze multi-year GDP data, aligning with data analysis.",
                },
                {
                    "value": "Other",
                    "confidence": 0.2,
                    "evidence": "Some image/design tasks and generic prompts not strictly related to data analysis.",
                },
                {
                    "value": "Tech Engineer",
                    "confidence": 0.2,
                    "evidence": "Technical interest shown in prompting about image generation and SEO.",
                },
                {
                    "value": "Student",
                    "confidence": 0.1,
                    "evidence": "Requests for simple explanations and knowledge summaries could be those of a student, but insufficient evidence to raise confidence further.",
                },
            ]
        },
    }
    user_result = PredictResult()
    user_result.load_from_data(result)
    print(user_result.occupation.candidates[0].value)
    print(user_result.school.candidates[0].value)
    print(user_result.primary_language.pick())

    ret = UserPredict(user_id=123)
    ret.load_from_result(user_result)
    print(ret.__dict__)

    genders = Candidates()
    a = Candidate()
    a.value = "female"
    a.confidence = 0.7
    b = Candidate()
    b.value = None
    b.confidence = 0.3
    genders.candidates.append(a)
    genders.candidates.append(b)
    print(f"xxxxxx {genders.pick()}")
