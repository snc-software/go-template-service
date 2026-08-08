FROM golang:1.26-alpine AS build

ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/api ./cmd/api

FROM golang:1.26-alpine AS migrate

ARG GOOSE_VERSION=v3.27.0

RUN go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

WORKDIR /migrations
ENTRYPOINT ["goose"]

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/api /api

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/api"]
