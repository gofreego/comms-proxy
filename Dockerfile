FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o application .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /src/application .
COPY dev.yaml .
COPY api/docs /app/api/docs
RUN chmod +x application

EXPOSE 80 8081

CMD [ "/app/application" ]
