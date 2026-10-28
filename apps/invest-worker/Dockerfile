# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.1

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG DEPLOY_ENV=dev
ARG AWS_DEPLOY_REGION=ap-northeast-1

RUN apk add --no-cache build-base git pkgconfig

WORKDIR /src

COPY . .

WORKDIR /src/apps/invest-worker

RUN mkdir -p /src/apps/invest-worker/bin \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
       go build -ldflags "-s -w" -o /src/apps/invest-worker/bin/invest-worker .

RUN cp /src/config/config.base.yml /src/apps/invest-worker/bin/config.base.yml \
    && sed -i "s/ENV_VALUE/${DEPLOY_ENV}/g" /src/apps/invest-worker/bin/config.base.yml \
    && sed -i "s/REGION_VALUE/${AWS_DEPLOY_REGION}/g" /src/apps/invest-worker/bin/config.base.yml

FROM alpine:3.19 AS runner

RUN apk add --no-cache ca-certificates tzdata libstdc++ curl \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder --chown=app:app /src/apps/invest-worker/bin/invest-worker /app/invest-worker
COPY --from=builder --chown=app:app /src/apps/invest-worker/bin/config.base.yml /app/config.base.yml
COPY --from=builder --chown=app:app /src/config/code /app/config/code

USER app

CMD ["/app/invest-worker"]
