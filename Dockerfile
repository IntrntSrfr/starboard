FROM golang:1.20-alpine AS build
WORKDIR /starboard
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o starboard ./cmd/starboard

FROM alpine:latest
WORKDIR /starboard
COPY --from=build /starboard/starboard .
ENTRYPOINT ["./starboard"]