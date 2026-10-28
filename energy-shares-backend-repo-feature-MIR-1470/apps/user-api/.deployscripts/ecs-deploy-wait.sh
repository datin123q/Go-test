#! /bin/bash

# Usage:
# ./ecs-wait.sh --cluster 'cluster-name' --service 'service-name' --task-definition 'task-definition'

set -e

function _timeout {
    gtimeout "$@"
}

# Read command line arguments
while [[ $# -gt 0 ]]
do
    key="$1"

    case $key in
        --cluster)
            ECS_CLUSTER="$2"
            shift
            ;;
        --service)
            ECS_SERVICE="$2"
            shift
            ;;
        --task-definition)
            ECS_TASK_DEFINITION="$2"
            shift
            ;;
        --region)
            REGION="$2"
            shift
            ;;
        --aws-access-key-id)
            AWS_ACCESS_KEY_ID="$2"
            shift
            ;;
        --aws-secret-access-key)
            AWS_SECRET_ACCESS_KEY="$2"
            shift
            ;;
        *)
            echo "ERROR: $1 is not a valid option"
            exit 2
        ;;
    esac
    shift
done

# Set AWS credentials
if [ "${AWS_ACCESS_KEY_ID}" != "" ] && [ "${AWS_SECRET_ACCESS_KEY}" != "" ]
then
    aws configure set aws_access_key_id "${AWS_ACCESS_KEY_ID}"
    aws configure set aws_secret_access_key "${AWS_SECRET_ACCESS_KEY}"
fi

# Get info about executing deployments in ECS
echo "aws ecs describe-services --cluster \"$ECS_CLUSTER\" --services \"$ECS_SERVICE\" --region \"$REGION\""
service_info=$(aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$REGION")
deployments=$(echo "$service_info" | jq -r '.services[0].deployments[].taskDefinition' | awk -F/ '{print $2}')

# Check if our task is currently deploying
in_progress=0
for deployment in ${deployments[@]}
do
    if [ "${deployment}" == "${ECS_TASK_DEFINITION}" ]
    then
        in_progress=1
    fi
done

# Keep checking every 5 seconds, for up to 3 mins for the task to start deploying in ECS
interval=5
timeout=180
while [[ $in_progress -lt 1 && $timeout -gt 0 ]]
do
    sleep $interval
    timeout=$[timeout - interval]

    service_info=$(aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$REGION")
    deployments=$(echo "$service_info" | jq -r '.services[0].deployments[].taskDefinition' | awk -F/ '{print $2}')

    for deployment in ${deployments[@]}
    do
        if [ "${deployment}" == "${ECS_TASK_DEFINITION}" ]
        then
            in_progress=1
        fi
    done
done

# If task has started deploying, test whether it deploys successfully
success=0
if [[ $in_progress -lt 1 ]]
then
    echo "ERROR"
    exit 2
else
    # Wait until the service is stable - either new task deployed successfully, or it fails and old task remains active
    aws ecs wait services-stable --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$REGION"

    # Check whether the new task is deployed to determine whether deployment was successful or failed
    service_info=$(aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$REGION")
    deployments=$(echo "$service_info" | jq -r '.services[0].deployments[].taskDefinition' | awk -F/ '{print $2}')

    for deployment in ${deployments[@]}
    do
        if [ "${deployment}" == "${ECS_TASK_DEFINITION}" ]
        then
            success=1
        fi
    done

    if [[ success -lt 1 ]]
    then
        echo "ERROR"
        exit 2
    else
        echo "SUCCESS" # don't change this message - used in Jenkinsfile logic!
    fi
fi
