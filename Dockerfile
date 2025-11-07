FROM golang:1.25-alpine AS build

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

RUN apk add --no-cache ca-certificates upx

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -tags netgo -ldflags="-s -w" -o /out/auth ./cmd/authd

RUN upx -9 /out/auth

FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates curl

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build /out/auth /app/auth

EXPOSE 8080

ENV PORT=8080

USER app
ENTRYPOINT ["/app/auth"]
