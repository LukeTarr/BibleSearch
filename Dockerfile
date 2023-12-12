FROM golang:1.21.5-bookworm

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /biblesearch
CMD sleep 3 && /biblesearch