#!/bin/bash

# shellcheck disable=SC2046
script_home=$(cd $(echo $0 | xargs dirname) ; pwd ; cd - > /dev/null)

# shellcheck disable=SC2034
script_log_file="/var/log/$(echo $0 | xargs basename | sed -e 's/.sh/.log/g')"

USER_UID="ubuntu"
USER_GID="ubuntu"
DOCKER_COMPOSE_CMD=$(which docker-compose)

export deploy_src_path="/tmp/build"
export swagger_ui_home="/home/ubuntu/swagger-ui"
export save_path="${swagger_ui_home}/data/dist"

function print_logging() {
        printf " %s\n" "$*" | sudo tee -a ${script_log_file}
}

function print_run() {
        ret_val=''
        print_logging "${FUNCNAME[1]} | CMD | $*"
        eval "$*"
        export  ret_val=$?
}

function swagger_job() {
        sudo mv ${deploy_src_path}/*api.json  ${save_path}/
        sudo chown ${USER_UID}.${USER_GID} ${save_path}/*.json
        sudo chmod -x ${save_path}/*.json
        # Add 2023.09.08
        if test -f ${swagger_ui_home}/bin/env_create.sh ; then
                print_run "${swagger_ui_home}/bin/env_create.sh"
        fi
}

function main() {
        print_logging ">> $0 script environment's << "
        print_logging " ++ script_home               = ${script_home}"
        print_logging " ++ deploy_src_path           = ${deploy_src_path}"

        print_logging "Move to Swagger json file"
        swagger_job;
        print_logging "Swagger Json Deploy Finished!!"
}

main $*