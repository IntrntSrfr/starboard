FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o starboard ./cmd/starboard

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=build /app/starboard .

ENTRYPOINT ["./starboard"]
