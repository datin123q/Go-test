#!/bin/bash

# mysqldump 플래그 설명:
# -h: 데이터베이스 서버의 호스트명 또는 IP 주소
# -u: 데이터베이스 사용자 이름
# --set-gtid-purged=OFF: GTID 정리 비활성화
# --single-transaction: 단일 트랜잭션 사용으로 일관된 덤프 보장
# --routines: 저장 프로시저와 함수 포함
# --triggers: 트리거 포함
# --events: 이벤트 스케줄러 이벤트 포함
# --no-tablespaces: 테이블 스페이스 정보 제외

echo "mynumber dev 데이터베이스 덤프 생성 중"
SECONDS=0 # 타이머 시작

mysqldump --defaults-file=./esjp-dev-dump.cnf --defaults-group-suffix=mynumber \
  --set-gtid-purged=OFF \
  --single-transaction \
  --routines \
  --triggers \
  --events \
  --no-tablespaces \
  esjp_mynumberdb_dev > esjp_mynumberdb_dev_dump.sql &
pid=$! # 마지막으로 실행된 백그라운드 프로세스의 PID

# 진행 중 디스플레이
while kill -0 $pid 2>/dev/null; do echo -n "."; sleep 1; done

echo
echo "mynumber dev 데이터베이스 덤프 생성 완료. 소요 시간: $SECONDS 초."
echo "---"

echo "esjp dev 데이터베이스 덤프 생성 중"
SECONDS=0 # 타이머 재설정

mysqldump --defaults-file=./esjp-dev-dump.cnf --defaults-group-suffix=esjp \
  --set-gtid-purged=OFF \
  --single-transaction \
  --routines \
  --triggers \
  --events \
  --no-tablespaces \
  esjp_esdb_dev > esjp_esdb_dev_dump.sql &
pid=$!

# 진행 중 디스플레이
while kill -0 $pid 2>/dev/null; do echo -n "."; sleep 1; done

echo
echo "esjp dev 데이터베이스 덤프 생성 완료. 소요 시간: $SECONDS 초."


