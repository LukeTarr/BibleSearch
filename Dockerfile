FROM golang:1.21.5-bookworm

WORKDIR /app

COPY . .

ARG CHROMA_URL
ENV CHROMA_URL=$CHROMA_URL

ARG OPENAI_API_KEY
ENV OPENAI_API_KEY=$OPENAI_API_KEY

RUN CGO_ENABLED=0 GOOS=linux go build -o /biblesearch
CMD sleep 3 && /biblesearch