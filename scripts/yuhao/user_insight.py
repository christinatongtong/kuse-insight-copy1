import json
import os
from random import sample
from typing import Dict, List
from openai import AzureOpenAI
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
import pandas as pd
from pathlib import Path
from dotenv import load_dotenv

task = '1733835172.89993'
report_name = f"{task}-report.md"

def preprocess_user_data(data: pd.DataFrame) -> Dict[str, List[str]]:
    """
    Aggregate prompts by user_id in chronological order.
    
    Args:
        data: DataFrame with columns [user_id, prompt, timestamp]
    Returns:
        Dict mapping user_id to list of ordered prompts
    """
    # Sort by timestamp and group by user_id
    return (data
            .sort_values('created_at')
            .groupby('user_id')['prompt']
            .agg(list)
            .to_dict())

def analyze_single_user(args: tuple) -> Dict:
    """
    Analyze prompts for a single user.
    """
    user_id, prompts, client, deployment_name = args
    
    # Load config file
    with open('config.json', 'r') as f:
        config = json.load(f)
    
    # Combine prompts into a single context
    combined_prompt = "\n".join([f"Prompt {i+1}: {p}" for i, p in enumerate(prompts)])
    
    system_message = f"""
        You are a helpful assistant tasked with classifying user prompts. You specialize in accurately classifying rather than guessing. You prefer to return 'Unknown' or '[]' for lists unless you have full confidence.
    For each input, respond with the following:
    - Occupations: Based on the prompt, infer 2-3 the user's possible occupations from this list: {config['occupations']} or "Unknown" if unclear.
    - Task Types: Identify 2-3 main task types from this list: {config['task_types']} or "Unknown" if unclear.
    - Tags: Suggest 2-3 relevant tags from this list: {config['tags']} or "Unknown" if unclear.
    - Dissatisfied: If the user's prompt indicates dissatisfaction, return 'true'. Otherwise, return 'false'. Return "Unknown" if unclear.
    return the result in the following JSON format:
    {{
        "professions": ["list of professions or empty"],
        "task_types": ["list of task types or empty"],
        "tags": ["list of tags or empty"],
        "dissatisfied": "true/false/unknown"
    }}"""

    try:
        response = client.chat.completions.create(
            model=deployment_name,
            messages=[
                {"role": "system", "content": system_message},
                {"role": "user", "content": combined_prompt}
            ],
            temperature=0.3,
            response_format={"type": "json_object"}
        )
        
        analysis = json.loads(response.choices[0].message.content)
        analysis['user_id'] = user_id
        return analysis
        
    except Exception as e:
        print(f"Error analyzing user {user_id}: {str(e)}")
        return {
            "user_id": user_id,
            "profession": "unknown",
            "topics": [],
            "task_type": "unknown",
            "tags": [],
            "dissatisfied": "unknown"
        }

def analyze_users(data: pd.DataFrame, client: AzureOpenAI, deployment_name: str, max_workers: int = 200) -> List[Dict]:
    """
    Analyze all users concurrently.
    """
    # Preprocess data
    user_prompts = preprocess_user_data(data)
    
    # Prepare arguments for concurrent execution
    # # random sample 100 users
    # user_prompts = dict(sample(list(user_prompts.items()), 100))
    args_list = [(user_id, prompts, client, deployment_name) 
                 for user_id, prompts in user_prompts.items()]
    
    # Run analysis concurrently
    results = []
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        results = list(executor.map(analyze_single_user, args_list))
    
    return results

def save_results(results: List[Dict], output_dir: str = "outputs"):
    """
    Save analysis results to JSON file.
    """
    Path(output_dir).mkdir(exist_ok=True)
    output_path = Path(output_dir) / f"user_analysis_{task}.json"
    
    with open(output_path, 'w') as f:
        json.dump(results, f, indent=2)
    
    print(f"Results saved to {output_path}")

def main():
    # read the api key from .env
    load_dotenv()
    api_key = os.getenv("azure_api_key")
    api_version = os.getenv("azure_api_version")
    endpoint = os.getenv("azure_endpoint")
    deployment_name = os.getenv("azure_deployment_name")
    # Initialize Azure OpenAI client
    client = AzureOpenAI(
        api_key=api_key,
        api_version=api_version,
        azure_endpoint=endpoint,
    )
    
    # Try different options to read potentially problematic CSV
    try:
        df = pd.read_csv(f"{task}.csv", on_bad_lines='skip', encoding='utf-8')
    except Exception as e:
        print(f"First attempt failed: {e}")
        try:
            # Try with a different engine
            df = pd.read_csv(f"{task}.csv", engine='python', encoding='utf-8')
        except Exception as e:
            print(f"Second attempt failed: {e}")
            # Last resort: try with a larger field size limit
            import sys
            import csv
            maxInt = sys.maxsize
            while True:
                try:
                    csv.field_size_limit(maxInt)
                    break
                except OverflowError:
                    maxInt = int(maxInt/10)
            df = pd.read_csv(f"{task}.csv", encoding='utf-8')
    
    print(f"Loaded {task}.csv")
    print(f"Shape: {df.shape}")
    print(df.head())
    # # use the top 2000 rows for analysis to testing
    # df = df.head(2000)

    # Run analysis
    results = analyze_users(df, client, deployment_name)
    
    # Save results
    save_results(results)

if __name__ == "__main__":
    main()
