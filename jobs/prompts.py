from schemas import UserModel, UserProperty
from typing import List, Dict, Optional


USER_AVATAR_PROMPT = """
The picture given to you is an avatar, describe the content of this picture
"""

USER_INSIGHT_SYSTEM_PROMPT = """
You are an AI assistant specialized in user profiling. Based on the following user profile and task records, analyze and infer the user's attributes and return the result in a structured JSON format.

### Requirements

1. For each attribute, return your output in **JSON format**.
2. For each attribute, include a "candidates" array. Each candidate must contain:
   - "value": the predicted value
   - "confidence": a float between 0 and 1 (your estimated likelihood)
   - "evidence": a short explanation of what input led you to this conclusion
3. If no information is available for an attribute, return a candidate with:
   - ""value": null", ""confidence": 1.0", and appropriate "evidence"

The JSON should include the following top-level attributes:
- "primary_language": language the user input, Simplified Chinese and Traditional Chinese defined as two different language
- "gender": predicted gender
- "school": string
- "major": string
- "degree_level": string (e.g., Undergraduate, Master's, PhD)
- "industry": user's work domain (e.g., finance, tech, education), and the industry must in blow categories:
  - Technology & Software
  - Education
  - Healthcare
  - Finance & Business Services
  - Media & Design
  - Government & Non-Profit
  - Science & Research
  - Manufacturing & Hardware
  - Other
- "occupation": user's job title or role, and the occupation must in blow categories:
  - Data Analysis
  - Student
  - Teacher
  - Designer
  - Marketing
  - Healthcare
  - Tech Engineer
  - Other

Please analyze the user data below and return a single valid JSON object following the format above.
"""


def format_user_prompt(
    user_profile: UserModel,
    filenames: List[str],
    task_prompts: List[str],
    summaries: List[str],
    image_description: str,
    user_property: Optional[UserProperty],
) -> str:
    prompt = "## Input:\n"
    # user_profile
    prompt += ">User Base Profile:\n"
    prompt += f"- Email: {user_profile.email}\n"
    prompt += f"- GivenName: {user_profile.given_name}\n"
    prompt += f"- FamilyName: {user_profile.family_name}\n"
    prompt += f"- FullName: {user_profile.full_name}\n"
    prompt += f"- SettingOutputLanguage: {user_profile.output_language}\n"

    if user_property:
        prompt += f"- In Education or not: {user_property.is_education}\n"
        prompt += f"- Region: {user_property.get_user_region()}\n"
    if image_description:
        prompt += f"- Profile Image Description: {image_description}\n"

    # tasks
    if len(task_prompts) > 0:
        prompt += ">The prompt that the user had input, detect the primary_language by following prompts:\n"
        visited: Dict[str, bool] = {}
        for task_prompt in task_prompts:
            if visited.get(task_prompt, False):  # filter repeat prompt
                continue
            prompt += f"- {task_prompt}\n"
            visited[task_prompt] = True

    # filenames
    if len(filenames) > 0:
        prompt += ">The FileName that the user uploaded:\n"
        for filename in filenames:
            prompt += f"- {filename}\n"

    # summaries
    if len(summaries) > 0:
        prompt += ">The FileSummary that the user uploaded:\n"
        for summary in summaries:
            prompt += f"- {summary}\n"

    # format
    prompt += """
Please reason carefully and return only the final JSON result, with no explanation or formatting outside the JSON.
Now, output the persona JSON:
        """
    return prompt
