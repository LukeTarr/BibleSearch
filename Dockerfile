# Build
FROM golang:1.25 AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/biblesearch .

# Run
FROM alpine:3.24 AS run-stage
WORKDIR /
COPY --from=build-stage /app/biblesearch /biblesearch
COPY ./assets /assets
COPY ./data /data
EXPOSE 8080
ENTRYPOINT ["/biblesearch"]
