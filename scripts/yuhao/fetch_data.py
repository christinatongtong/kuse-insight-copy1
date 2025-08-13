# Use GCP to load data and download relevant files

import argparse
import pandas as pd
from google.cloud import bigquery
from google.cloud import storage
import os
import concurrent.futures
import time
from datetime import datetime
import json

# import stripe  # Temporarily disabled
from typing import Dict, List, Optional
from dotenv import load_dotenv
import requests
from urllib.parse import urlencode

# Load environment variables from .env file
load_dotenv()

# Auth0 configuration
AUTH0_DOMAIN = os.getenv("AUTH0_DOMAIN")
AUTH0_CLIENT_ID = os.getenv("AUTH0_CLIENT_ID")
AUTH0_CLIENT_SECRET = os.getenv("AUTH0_CLIENT_SECRET")


def get_auth0_token():
    """Get Auth0 Management API access token."""
    url = f"https://{AUTH0_DOMAIN}/oauth/token"
    headers = {"content-type": "application/x-www-form-urlencoded"}
    data = {
        "grant_type": "client_credentials",
        "client_id": AUTH0_CLIENT_ID,
        "client_secret": AUTH0_CLIENT_SECRET,
        "audience": f"https://{AUTH0_DOMAIN}/api/v2/",
    }

    try:
        response = requests.post(url, headers=headers, data=urlencode(data))
        response.raise_for_status()
        return response.json()["access_token"]
    except Exception as e:
        print(f"Error getting Auth0 token: {e}")
        return None


def get_auth0_data(email: str) -> Dict:
    """Get user data from Auth0."""
    if not all([AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET]):
        print("Auth0 configuration missing")
        return {}

    token = get_auth0_token()
    if not token:
        return {}

    url = f"https://{AUTH0_DOMAIN}/api/v2/users-by-email"
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    params = {"email": email}

    try:
        response = requests.get(url, headers=headers, params=params)
        response.raise_for_status()
        users = response.json()

        if not users:
            return {}

        # Get the first user (most relevant)
        user = users[0]
        auth0_data = {
            "auth0.user_id": user.get("user_id", ""),
            "auth0.name": user.get("name", ""),
            "auth0.nickname": user.get("nickname", ""),
            "auth0.created_at": user.get("created_at", ""),
            "auth0.last_login": user.get("last_login", ""),
            "auth0.logins_count": user.get("logins_count", 0),
            "auth0.email_verified": user.get("email_verified", False),
            "auth0.user_metadata": json.dumps(user.get("user_metadata", {})),
            "auth0.app_metadata": json.dumps(user.get("app_metadata", {})),
            "auth0.identities": json.dumps(user.get("identities", [])),
        }
        return auth0_data
    except Exception as e:
        print(f"Error getting Auth0 data for {email}: {e}")
        return {}


def enrich_with_auth0_data(df):
    """Enrich DataFrame with Auth0 user data."""
    auth0_data = []
    for email in df["email"].unique():
        user_data = get_auth0_data(email)
        if user_data:
            auth0_data.append({"email": email, **user_data})

    if not auth0_data:
        return df

    auth0_df = pd.DataFrame(auth0_data)
    return df.merge(auth0_df, on="email", how="left")


def decode_unicode_str(s):
    """Decode unicode escaped string."""
    try:
        # If the string is a JSON string, parse it first
        if isinstance(s, str):
            try:
                s = json.loads(s)
            except json.JSONDecodeError:
                pass

        # If it's still a string with escaped unicode, decode it
        if isinstance(s, str):
            return s.encode().decode("unicode_escape")
        return s
    except Exception:
        return s


def process_dataframe(df):
    """Process dataframe to decode unicode strings in task_meta and result_meta."""
    if "task_meta" in df.columns:
        df["task_meta"] = df["task_meta"].apply(decode_unicode_str)
    if "result_meta" in df.columns:
        df["result_meta"] = df["result_meta"].apply(decode_unicode_str)
    return df


def read_email_list(csv_path):
    """Read emails from CSV file, skipping bad values."""
    try:
        df = pd.read_csv(csv_path)
        if "email" not in df.columns:
            print(
                f"Error: CSV file does not contain 'email' column. "
                f"Available columns: {df.columns.tolist()}"
            )
            return []

        # Remove NaN, empty strings, and invalid emails (basic validation)
        emails = df["email"].dropna().str.strip()
        emails = emails[emails != ""]
        # Basic email validation (contains @)
        emails = emails[emails.str.contains("@")]

        print(f"Found {len(emails)} valid emails in {csv_path}")
        return emails.tolist()
    except Exception as e:
        print(f"Error reading CSV file: {e}")
        return []


def query_user_data(emails, client):
    """Query prompts and results for the given emails."""
    if not emails:
        print("No emails to query")
        return pd.DataFrame()

    # Format emails for SQL IN clause
    email_list = "', '".join(emails)

    query = f"""
    SELECT * FROM EXTERNAL_QUERY("kuse-ai.us.kuse-ai-main", 
        \"\"\"
        WITH RankedTasks AS (
            SELECT 
                u.id,
                u.email,
                u.image_url,
                CONVERT(task_meta USING utf8) AS task_meta,
                CONVERT(result_meta USING utf8) AS result_meta,
                t.created_at,
                t.task_type,
                ROW_NUMBER() OVER (PARTITION BY u.email ORDER BY t.created_at DESC) as rn
            FROM tasks t 
            JOIN user u ON t.user_id = u.id 
            WHERE u.email IN ('{email_list}')
        )
        SELECT 
            id,
            email,
            image_url,
            task_meta,
            result_meta,
            created_at,
            task_type
        FROM RankedTasks
        WHERE rn <= 100
        \"\"\");
    """

    try:
        query_job = client.query(query)
        results_df = query_job.to_dataframe()
        print(f"Retrieved {len(results_df)} tasks for {len(emails)} users")
        return results_df
    except Exception as e:
        print(f"Error querying user data: {e}")
        return pd.DataFrame()


def download_file(row, user_dir, storage_client):
    """Download a single file. Used for parallel downloading."""
    try:
        filename = row["filename"]
        filepath = row["filepath"]  # e.g upload/u1231214.pdf
        file_size = row["file_size"]

        # e.g upload/u1231214.pdf -> gs://smart-design-assets/upload/u1231214.pdf
        bucket_name = "smart-design-assets"
        blob_path = filepath
        bucket = storage_client.bucket(bucket_name)
        blob = bucket.blob(blob_path)

        destination_path = os.path.join(user_dir, filename)
        blob.download_to_filename(destination_path)
        return True, f"Downloaded {filename} ({file_size} bytes)"
    except Exception as e:
        return False, f"Error downloading {row.get('filename', 'unknown')}: {e}"


def query_and_download_files(
    emails, client, storage_client, output_dir, max_workers=10
):
    """Query and download input files for the given emails in parallel."""
    if not emails:
        print("No emails to query for files")
        return

    start_time = time.time()
    total_files = 0
    successful_downloads = 0

    for email in emails:
        try:
            # Query files for this user
            query = f"""
            SELECT * FROM EXTERNAL_QUERY("kuse-ai.us.kuse-ai-main",
                \"\"\"
                SELECT filename, filepath, file_size 
                FROM files f 
                JOIN user i ON f.user_id = i.id 
                WHERE i.email = '{email}'
                ORDER BY f.created_at DESC
                LIMIT 20
                \"\"\");
            """

            query_job = client.query(query)
            files_df = query_job.to_dataframe()

            if files_df.empty:
                print(f"No files found for {email}")
                continue

            file_count = len(files_df)
            total_files += file_count
            print(f"Found {file_count} files for {email}")

            # Create directory for this user's files
            user_dir = os.path.join(output_dir, email.replace("@", "_at_"))
            os.makedirs(user_dir, exist_ok=True)

            # Download files in parallel
            with concurrent.futures.ThreadPoolExecutor(
                max_workers=max_workers
            ) as executor:
                # Create a list of future tasks
                future_to_file = {
                    executor.submit(download_file, row, user_dir, storage_client): row[
                        "filename"
                    ]
                    for _, row in files_df.iterrows()
                }

                # Process results as they complete
                for future in concurrent.futures.as_completed(future_to_file):
                    filename = future_to_file[future]
                    success, message = future.result()
                    if success:
                        successful_downloads += 1
                    print(message)

        except Exception as e:
            print(f"Error processing files for {email}: {e}")

    elapsed_time = time.time() - start_time
    print(
        f"File processing complete: {successful_downloads}/{total_files} "
        f"files downloaded in {elapsed_time:.2f} seconds"
    )


def flatten_task_data(df):
    """Flatten the task data by extracting specified fields from JSON."""
    # Create a copy of the dataframe
    flattened = df.copy()

    def extract_field(x, field):
        """Extract field from task_meta or result_meta."""
        if pd.isnull(x):
            return None
        if isinstance(x, dict):
            return x.get(field)
        if isinstance(x, str):
            try:
                return json.loads(x).get(field)
            except json.JSONDecodeError:
                return None
        return None

    # Extract fields from task_meta
    flattened["prompt"] = df["task_meta"].apply(lambda x: extract_field(x, "prompt"))
    flattened["file_ids"] = df["task_meta"].apply(
        lambda x: extract_field(x, "file_ids")
    )
    flattened["texts"] = df["task_meta"].apply(lambda x: extract_field(x, "texts"))
    flattened["urls"] = df["task_meta"].apply(lambda x: extract_field(x, "urls"))
    flattened["image_urls"] = df["task_meta"].apply(
        lambda x: extract_field(x, "image_urls")
    )
    flattened["video_ids"] = df["task_meta"].apply(
        lambda x: extract_field(x, "video_ids")
    )
    flattened["selected_text"] = df["task_meta"].apply(
        lambda x: extract_field(x, "selected_text")
    )
    flattened["mention_agents"] = df["task_meta"].apply(
        lambda x: extract_field(x, "mention_agents")
    )

    # Extract markdown from result_meta
    flattened["markdown"] = df["result_meta"].apply(
        lambda x: extract_field(x, "markdown")
    )

    # Select only the specified columns
    columns_to_keep = [
        "id",
        "email",
        "created_at",
        "task_type",
        "prompt",
        "file_ids",
        "texts",
        "urls",
        "image_urls",
        "video_ids",
        "selected_text",
        "mention_agents",
        "markdown",
    ]

    return flattened[columns_to_keep]


# def get_stripe_data(email: str) -> Dict:
#     """Query Stripe API for customer data and payment history."""
#     try:
#         # Search for customer by email
#         customers = stripe.Customer.list(email=email)

#         if not customers.data:
#             return {
#                 'customer_id': None,
#                 'name': None,
#                 'email': email,
#                 'city': None,
#                 'country': None,
#                 'created_at': None,
#                 'payment_methods': [],
#                 'payments': []
#             }

#         customer = customers.data[0]

#         # Get payment methods
#         payment_methods = stripe.PaymentMethod.list(
#             customer=customer.id,
#             type='card'
#         )

#         # Get payment history (charges)
#         charges = stripe.Charge.list(customer=customer.id)

#         # Format payment methods
#         formatted_payment_methods = []
#         for pm in payment_methods.data:
#             if pm.card:
#                 formatted_payment_methods.append({
#                     'type': pm.type,
#                     'brand': pm.card.brand,
#                     'last4': pm.card.last4,
#                     'exp_month': pm.card.exp_month,
#                     'exp_year': pm.card.exp_year
#                 })

#         # Format payments
#         formatted_payments = []
#         for charge in charges.data:
#             formatted_payments.append({
#                 'amount': charge.amount / 100.0,  # Convert cents to dollars
#                 'currency': charge.currency,
#                 'status': charge.status,
#                 'created_at': datetime.fromtimestamp(charge.created).isoformat(),
#                 'payment_method_details': charge.payment_method_details.type
#             })

#         return {
#             'customer_id': customer.id,
#             'name': customer.name,
#             'email': customer.email,
#             'city': customer.address.city if customer.address else None,
#             'country': customer.address.country if customer.address else None,
#             'created_at': datetime.fromtimestamp(customer.created).isoformat(),
#             'payment_methods': formatted_payment_methods,
#             'payments': formatted_payments
#         }
#     except stripe.error.StripeError as e:
#         print(f"Error querying Stripe for {email}: {str(e)}")
#         return None


# def enrich_with_stripe_data(df):
#     """Add Stripe customer and payment information to the dataframe."""
#     # Get unique emails
#     unique_emails = df['email'].unique()

#     # Create a new dataframe for user profiles
#     user_profiles = pd.DataFrame({'email': unique_emails})

#     # Query Stripe data for each email
#     stripe_data = {}
#     for email in unique_emails:
#         stripe_info = get_stripe_data(email)
#         if stripe_info:
#             stripe_data[email] = stripe_info

#     # Add Stripe data as new columns with 'stripe.' prefix
#     user_profiles['stripe.customer_id'] = user_profiles['email'].map(
#         lambda x: stripe_data.get(x, {}).get('customer_id'))
#     user_profiles['stripe.name'] = user_profiles['email'].map(
#         lambda x: stripe_data.get(x, {}).get('name'))
#     user_profiles['stripe.city'] = user_profiles['email'].map(
#         lambda x: stripe_data.get(x, {}).get('city'))
#     user_profiles['stripe.country'] = user_profiles['email'].map(
#         lambda x: stripe_data.get(x, {}).get('country'))
#     user_profiles['stripe.created_at'] = user_profiles['email'].map(
#         lambda x: stripe_data.get(x, {}).get('created_at'))
#     user_profiles['stripe.payment_methods'] = user_profiles['email'].map(
#         lambda x: json.dumps(stripe_data.get(x, {}).get('payment_methods', [])))
#     user_profiles['stripe.payments'] = user_profiles['email'].map(
#         lambda x: json.dumps(stripe_data.get(x, {}).get('payments', [])))

#     return user_profiles


def main():
    parser = argparse.ArgumentParser(description="Process user data and download files")
    parser.add_argument(
        "--input-file", required=True, help="Path to CSV file containing emails"
    )
    parser.add_argument(
        "--output-dir", required=True, help="Directory to save output files"
    )
    parser.add_argument(
        "--parallel-workers", type=int, default=10, help="Number of parallel workers"
    )
    args = parser.parse_args()

    # Create timestamped output directory
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_dir = os.path.join(args.output_dir, timestamp)
    os.makedirs(output_dir, exist_ok=True)

    # Initialize clients
    bigquery_client = bigquery.Client()
    storage_client = storage.Client()

    # Read emails from input file
    input_df = pd.read_csv(args.input_file)
    emails = read_email_list(args.input_file)
    if not emails:
        return

    # Query user data
    df = query_user_data(emails, bigquery_client)
    if df.empty:
        return

    # save email, image_url to csv
    # deduplicate by email
    df = df.drop_duplicates(subset=["email"])
    # df[['email', 'image_url', ""]].to_csv(os.path.join(output_dir, 'user_profiles.csv'), index=False)
    # merge image url into input_df
    input_df = input_df.merge(df[["email", "image_url"]], on="email", how="left")
    input_df.to_csv(os.path.join(output_dir, "user_profiles.csv"), index=False)

    # # Process and decode unicode strings
    # df = process_dataframe(df)

    # # Flatten task data
    # df = flatten_task_data(df)

    # # Create user profiles with Auth0 data
    # user_profiles = pd.DataFrame({'email': df['email'].unique()})
    # auth0_data = []
    # for email in user_profiles['email']:
    #     user_data = get_auth0_data(email)
    #     if user_data:
    #         auth0_data.append({'email': email, **user_data})

    # if auth0_data:
    #     auth0_df = pd.DataFrame(auth0_data)
    #     user_profiles = user_profiles.merge(auth0_df, on='email', how='left')

    # # Merge with input file to preserve any additional columns
    # if 'email' in input_df.columns:
    #     user_profiles = pd.merge(
    #         input_df,
    #         user_profiles,
    #         on='email',
    #         how='left'
    #     )

    # # Save user profiles
    # user_profiles_file = os.path.join(output_dir, 'user_profiles.csv')
    # user_profiles.to_csv(user_profiles_file, index=False)
    # print(f"Saved user profiles to {user_profiles_file}")

    # # Save tasks data
    # tasks_file = os.path.join(output_dir, 'tasks.csv')
    # df.to_csv(tasks_file, index=False)
    # print(f"Saved tasks data to {tasks_file}")

    # # Save individual user task files
    # for email in df['email'].unique():
    #     user_tasks = df[df['email'] == email]
    #     filename = f"{email.replace('@', '_at_')}_tasks.csv"
    #     user_tasks_file = os.path.join(output_dir, filename)
    #     user_tasks.to_csv(user_tasks_file, index=False)
    #     print(f"Saved tasks for {email} to {user_tasks_file}")

    # # Download files
    # query_and_download_files(
    #     emails,
    #     bigquery_client,
    #     storage_client,
    #     output_dir,
    #     args.parallel_workers
    # )


if __name__ == "__main__":
    main()
