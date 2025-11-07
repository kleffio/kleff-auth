FROM golang:1.25-alpine AS build

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

RUN apk add --no-cache ca-certificates upx

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -tags netgo -ldflags="-s -w" -o /out/auth ./cmd/authd

RUN upx -9 /out/auth


FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/auth /app/auth

EXPOSE 8080

ENV PORT=8080

USER nonroot:nonroot
ENTRYPOINT ["/app/auth"]
