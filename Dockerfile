# Fetch
FROM golang:latest AS fetch-stage
COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod tidy

# Generate Templ files
FROM ghcr.io/a-h/templ:latest AS generate-stage
COPY --chown=65532:65532 . /app
WORKDIR /app
RUN ["templ", "generate"]

# Build
FROM golang:latest AS build-stage
COPY --from=generate-stage /app /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/biblesearch

# Run
FROM alpine:latest AS run-stage
WORKDIR /
COPY --from=build-stage /app/biblesearch /biblesearch
COPY ./assets /assets
COPY ./data /data
COPY .env /.env
EXPOSE 8080
ENTRYPOINT ["/biblesearch"]
