#!/bin/bash

# 이미 실행 중인 컨테이너가 있으면 중지하고 제거
docker stop mysql8.0.32 > /dev/null 2>&1 && docker rm mysql8.0.32 > /dev/null 2>&1

echo -n "기존 컨테이너를 중지하고 삭제하는 중입니다..."

# 컨테이너 삭제가 완료될 때까지 기다리기
until [ "$(docker ps -aq -f name=mysql8.0.32)" == "" ]; do
  echo -n "."
  sleep 1
done

echo -e "\n컨테이너 삭제 완료. 새 컨테이너를 시작합니다."

# Docker에서 MySQL 8.0.32 컨테이너 시작
docker run --rm --name mysql8.0.32 -e MYSQL_ROOT_PASSWORD=password -d -p 3306:3306 mysql:8.0.32

echo -n "MySQL 서버가 준비될 때까지 기다리는 중..."

# MySQL 서버가 준비될 때까지 기다림
until docker exec mysql8.0.32 mysql -u root -ppassword -e "SELECT 1" > /dev/null 2>&1; do
  echo -n "."
  sleep 1
done

echo -e "\nMySQL 서버가 준비되었습니다. 데이터베이스 복원을 시작합니다."

# esjp_mynumberdb_dev 데이터베이스를 Docker 컨테이너에서 복원
docker exec -i mysql8.0.32 mysql -u root -ppassword -e "CREATE DATABASE IF NOT EXISTS mynumberdb;"
docker exec -i mysql8.0.32 sh -c 'exec mysql -u root -ppassword mynumberdb' < esjp_mynumberdb_dev_dump.sql&
pid=$!
while kill -0 $pid 2>/dev/null; do sleep 1; echo -n "."; done
echo

# esjp_esdb_dev 데이터베이스를 Docker 컨테이너에서 복원
docker exec -i mysql8.0.32 mysql -u root -ppassword -e "CREATE DATABASE IF NOT EXISTS energysharesdb;"
docker exec -i mysql8.0.32 sh -c 'exec mysql -u root -ppassword energysharesdb' < esjp_esdb_dev_dump.sql&
pid=$!
while kill -0 $pid 2>/dev/null; do sleep 1; echo -n "."; done
echo

echo "데이터베이스 복원이 성공적으로 완료되었습니다."
