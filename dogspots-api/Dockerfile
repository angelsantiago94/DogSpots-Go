FROM golang:1.26.9-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/reports ./cmd/reports

FROM alpine:3.22

RUN apk --no-cache add ca-certificates sqlite \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /data/db /data/media \
    && chown -R app:app /data

COPY --from=builder --chown=app:app /out/api /usr/local/bin/dogspots-api
COPY --from=builder --chown=app:app /out/reports /usr/local/bin/dogspots-reports

USER app:app

EXPOSE 8080

ENTRYPOINT ["dogspots-api"]