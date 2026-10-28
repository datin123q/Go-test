# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.1

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=unknown
ARG COMMIT=unknown
ARG DEPLOY_ENV=dev

RUN apk add --no-cache build-base git pkgconfig

WORKDIR /src

COPY . .

RUN make admin-api CMD=build DEPLOY_ENV=${DEPLOY_ENV}

FROM alpine:3.19 AS runner
ARG VERSION=unknown
ARG COMMIT=unknown

RUN apk add --no-cache ca-certificates tzdata libstdc++ curl \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app

ENV VERSION_INFO=${VERSION}-${COMMIT} \
    PORT=8080

COPY --from=builder --chown=app:app /src/apps/admin-api/bin /app
COPY --from=builder --chown=app:app /src/config /app/config

RUN ls -la /app

USER app

EXPOSE 8080

CMD ["/bin/sh", "-c", "/app/admin-api migrate && /app/admin-api run"]
