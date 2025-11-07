FROM golang:alpine AS builder

RUN mkdir /src
ADD . /src
WORKDIR /src
RUN go mod tidy -v
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/main.go


FROM alpine:latest AS production
RUN apk add --no-cache tzdata
# Set timezone ke Asia/Jakarta
ENV TZ=Asia/Jakarta

WORKDIR /app
COPY --from=builder /src/api .
CMD ["./api"]