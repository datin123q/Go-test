#!/bin/bash

# 프로젝트 내 모든 go.mod 파일을 찾아 해당 디렉토리로 이동하여 go mod tidy를 실행하는 스크립트
# 현재 디렉토리에서 시작하여 모든 go.mod 파일을 찾습니다.
MOD_FILES=$(find . -type f -name 'go.mod')

# 각 go.mod 파일에 대해...
for MOD_FILE in $MOD_FILES; do
    # go.mod 파일이 있는 디렉토리로 이동
    MOD_DIR=$(dirname $MOD_FILE)

    echo "Running 'go mod tidy' in $MOD_DIR"
    (cd $MOD_DIR && go mod tidy)
    echo "---"
    echo
done

echo "All modules have been tidied."
