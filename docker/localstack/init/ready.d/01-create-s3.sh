#!/bin/bash

set -euo pipefail

export AWS_DEFAULT_REGION=ap-northeast-1

awslocal s3api create-bucket \
  --bucket mira-local-public-storage \
  --region ap-northeast-1 \
  --create-bucket-configuration LocationConstraint=ap-northeast-1 || true

awslocal s3api create-bucket \
  --bucket mira-local-private-storage \
  --region ap-northeast-1 \
  --create-bucket-configuration LocationConstraint=ap-northeast-1 || true

# ブラウザから署名付きURLへ直接PUTするため、両バケットにCORSが必要。
# private 側はKYC書類・マイナンバー等のアップロード先になる。
CORS_CONFIGURATION='{
    "CORSRules": [
      {
        "AllowedOrigins": ["http://localhost:3000", "http://localhost:3001", "http://localhost:23000", "http://localhost:23001"],
        "AllowedMethods": ["GET","HEAD","PUT","POST","DELETE"],
        "AllowedHeaders": ["*"],
        "ExposeHeaders": ["Content-Type","Content-Length","Content-Disposition"],
        "MaxAgeSeconds": 3000
      }
    ]
  }'

for bucket in mira-local-public-storage mira-local-private-storage; do
  awslocal s3api put-bucket-cors \
    --bucket "$bucket" \
    --cors-configuration "$CORS_CONFIGURATION"
done
