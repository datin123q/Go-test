#!/bin/bash

set -euo pipefail

export AWS_DEFAULT_REGION=ap-northeast-1

awslocal sqs create-queue \
  --queue-name my-local-test-queue.fifo \
  --attributes FifoQueue=true || true
