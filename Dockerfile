FROM golang:1.25.12-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /trident-mcp ./cmd/trident-mcp

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /trident-mcp /usr/local/bin/trident-mcp
ENTRYPOINT ["trident-mcp"]
