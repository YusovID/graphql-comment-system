FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY go.mod go.sum ./
RUN go mod download

# build
COPY app ./app
COPY .env ./

RUN go build -o /usr/local/src/bin/app /usr/local/src/app/cmd/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY --from=builder /usr/local/src/.env /

CMD ["/app"]