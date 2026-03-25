FROM --platform=$BUILDPLATFORM golang:1.26.1-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH
ARG SERVICE_NAME=audit
ARG VERSION=dev

RUN apk add --no-cache git

WORKDIR /src

COPY go.work ./
COPY api/gen/go.mod api/gen/go.sum ./api/gen/
COPY app/audit/service/go.mod app/audit/service/go.sum ./app/audit/service/

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Name=${SERVICE_NAME}.service" \
    -o /src/bin/${SERVICE_NAME} ./app/${SERVICE_NAME}/service/cmd/server

FROM alpine:3.19

ARG SERVICE_NAME=audit

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /src/bin/${SERVICE_NAME} /app/${SERVICE_NAME}

VOLUME /app/configs

ENV TZ=Asia/Shanghai
ENV SERVICE_NAME=${SERVICE_NAME}

CMD ["/bin/sh", "-c", "/app/${SERVICE_NAME} -conf /app/configs/"]
