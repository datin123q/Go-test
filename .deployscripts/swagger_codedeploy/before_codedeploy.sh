#!/bin/bash

# shellcheck disable=SC2046
script_home=$(cd $(echo $0 | xargs dirname) ; pwd ; cd - > /dev/null)

deploy_type=$(echo ${script_home} | awk -F '/' '{print $(NF)}')

echo "Code Deploy Type : ${deploy_type}"
echo "Before script :  $0"
echo "Docker version : $(docker --version)"