FROM --platform=$BUILDPLATFORM golang:1.26.1-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH
ARG SERVICE_NAME=master
ARG VERSION=dev

ENV GOWORK=off

RUN apk add --no-cache git

WORKDIR /src

# Copy go.mod/go.sum for each module first (better layer caching)
COPY api/gen/go.mod api/gen/go.sum ./api/gen/
COPY app/master/service/go.mod app/master/service/go.sum ./app/master/service/
COPY app/worker/service/go.mod app/worker/service/go.sum ./app/worker/service/

# Download dependencies for the target service
RUN cd /src/app/${SERVICE_NAME}/service && go mod download

COPY . .

RUN cd /src/app/${SERVICE_NAME}/service && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Name=${SERVICE_NAME}.service" \
    -o /src/bin/${SERVICE_NAME} ./cmd/server

FROM alpine:3.19

ARG SERVICE_NAME=master

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /src/bin/${SERVICE_NAME} /app/${SERVICE_NAME}

VOLUME /app/configs

ENV TZ=Asia/Shanghai
ENV SERVICE_NAME=${SERVICE_NAME}

CMD ["/bin/sh", "-c", "/app/${SERVICE_NAME} -conf /app/configs/"]
