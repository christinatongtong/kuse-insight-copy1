import csv
from typing import List
from openpyxl import load_workbook
from datetime import datetime, timedelta

user_results = {}
occupations = {}
industry = {}
industry_categories = {
    "Education": "Education",
    "Technology / IT": "Technology / IT",
    "Healthcare / Medical": "Healthcare / Medical",
    "Finance / Accounting": "Finance / Accounting",
    "Engineering / Construction": "Engineering / Construction",
    "Marketing / Media / Design": "Marketing / Media / Design",
    "Public Sector / Government": "Public Sector / Government",
    "Retail / Consumer Goods / E-commerce": "Retail / Consumer Goods / E-commerce",
    "Legal": "Legal",
    "Science / Sustainability / Research": "Science / Sustainability / Research",
    "Social / Nonprofit / Community": "Social / Nonprofit / Community",
    "Arts / Culture / Creative": "Arts / Culture / Creative",
    "Travel / Hospitality / Food & Beverage": "Travel / Hospitality / Food & Beverage",
}


def read_csv_file(filepath: str):
    """
    读取 CSV 文件并返回列表形式的数据（含表头）
    """
    with open(filepath, newline="", encoding="utf-8") as csvfile:
        reader = csv.reader(csvfile)
        data = [row for row in reader]
    return data


def load_user_results():
    data = read_csv_file("./results/results.csv")
    for i in range(len(data)):
        if i == 0:
            header = data[i]
            continue
        user_results[data[i][0]] = data[i]
    return header


def cluster():
    rows = []
    for k, v in occupations.items():
        rows.append(k)
    print(f"occupations: {rows}")

    rows = []
    for k, v in industry.items():
        rows.append(k)
    print(f"industry: {rows}")


def is_recent(dt_str: str) -> bool:
    if not dt_str:
        return False
    # 解析为 datetime 对象
    dt = datetime.strptime(dt_str, "%Y-%m-%dT%H:%M:%S")

    # 当前时间
    now = datetime.now()

    # 三个月前的时间（粗略以90天计算）
    three_months_ago = now - timedelta(days=14)

    # 判断是否在最近三个月内
    return three_months_ago <= dt <= now


def upload():
    header = load_user_results()
    # print("--", header)
    # print("xx", user_results["malecaroh#gmail.com"])
    rows = []
    guests = []
    for k, v in user_results.items():
        row = list()
        isGuest = False
        for column in range(len(v)):
            value = v[column]
            if header[column] == "is_guest_mode" and value == "true":
                isGuest = True
            if value == "<nil>" or value == "":
                value = "-"
            if header[column] == "is_student" or header[column] == "gender":
                value = value.lower()
            if header[column] == "occupation":
                occupations[value] = 1
            if header[column] == "industry":
                industry[value] = 1

            row.append(value)

        if isGuest:
            guests.append(row)
        else:
            rows.append(row)
    create_csv_file("./predict.csv", header, rows)
    create_csv_file("./predict_guest.csv", header, guests)


def create_csv_file(filepath, header, rows):
    with open(filepath, "w", newline="", encoding="utf-8") as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(header)
        writer.writerows(rows)

    print(f"CSV 文件已创建：{filepath}")


if __name__ == "__main__":
    upload()
