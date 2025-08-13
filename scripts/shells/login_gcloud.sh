#!/bin/bash

# install gcloud
# brew install --cask google-cloud-sdk # install gcloud

# 初始化 gcloud
# gcloud init

# # 或者直接登录
# gcloud auth login

# # 设置应用默认凭据（这是关键步骤）
gcloud auth application-default login

# # 设置默认项目
# gcloud config set project kuse-ai

# # 查看当前配置
# gcloud config list

# # 查看当前认证账户
# gcloud auth list

# # 测试访问
# gcloud projects list
