#!/bin/bash

# 스크립트 이름: paramstore.sh
# 사용법: paramstore.sh get/set {env1} [env2] ...

# 사용법을 출력하는 함수
# 사용법을 출력하는 함수
usage() {
    echo ""
    echo "사용법: $0 get/set/clean env1 [env2] ..."
    echo "      env: local, dev, qa, stg, prod, 또는 all"
    echo "      예: $0 get dev"
    echo "          $0 get all"
    echo "          $0 set dev qa"
    echo "          $0 set qa stg local"
    echo "          $0 clean dev"
    exit 1
}

# 최소한의 파라미터 확인
if [ "$#" -lt 2 ]; then
    usage
fi

# 명령과 환경변수 추출
command=$1
shift   # 첫번째 인자(command) 제거
envs=$@ # 나머지 모든 인자들

# 가능한 명령
valid_commands="get set clean"
valid_envs="local dev qa stg prod all"

# 명령 검증
if [[ ! " $valid_commands " =~ " $command " ]]; then
    echo "명령 오류: 유효하지 않은 명령입니다. '${command}'"
    usage
fi

# 환경변수 검증
for env in $envs; do
    if [[ ! " $valid_envs " =~ " $env " ]]; then
        echo "환경변수 오류: 유효하지 않은 환경입니다. '${env}'"
        usage
    fi
done

# 'all'이 포함되어 있으면 all 만 있으면 된다.
if [[ " $envs " =~ " all " ]]; then
    envs="all"
fi

# 변수 선언
all_envs="local dev qa stg prod"

# 주어진 환경변수 하나에 대해 명령 실행
execute_command() {
    local environment=$1
    PREFIX="/esjp/backend/"
    parameter_name="${PREFIX}${environment}"
    file_path="./config/config.${environment}.yml"

    if [ "$command" == "set" ]; then
        # set 명령 - 파라미터 스토어에 값을 Upsert
        if [ ! -f "$file_path" ]; then
            echo "오류: 파일을 찾을 수 없습니다: $file_path"
            return 2
        fi
        value=$(<"$file_path")
        kms_key_id="alias/aws/ssm"
        update_status=$(aws ssm put-parameter --name "$parameter_name" --value "$value" --type "SecureString" --key-id "$kms_key_id" --tier "Advanced" --overwrite)

        if [ $? -ne 0 ]; then
            echo "파라미터 업데이트 오류: $update_status"
            return 3
        else
            echo "파라미터가 성공적으로 업데이트 되었습니다: $parameter_name"
        fi
    elif [ "$command" == "get" ]; then
        # get 명령
        value=$(aws ssm get-parameter --name "$parameter_name" --with-decryption --query "Parameter.Value" --output text 2>&1)
        if [ $? -ne 0 ]; then
            echo "파라미터 가져오기 오류: $value"
            return 2
        else
            printf "%b\n" "$value" > "$file_path"
            echo "YAML 파일이 생성되었습니다: $file_path"
        fi
    elif [ "$command" == "clean" ]; then
            # clean 명령 - 로컬 파일 삭제
            if [ -f "$file_path" ]; then
                rm "$file_path"
                if [ $? -ne 0 ]; then
                    echo "파일 삭제 오류: $file_path"
                    return 2
                else
                    echo "파일이 성공적으로 삭제되었습니다: $file_path"
                fi
            else
                echo "파일을 찾을 수 없습니다: $file_path"
            fi
        fi
}

# 모든 지정된 환경에 대해 명령 실행
if [ "$envs" == "all" ]; then
    # all 이면 모든 환경에 대해 실행
    for e in $all_envs; do
        execute_command $e
    done
else
    for env in $envs; do
        execute_command $env
    done
fi
