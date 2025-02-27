FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# dependencies
COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

# build
COPY app ./
RUN go build -o ./bin/app cmd/main.go

FROM alpine AS runner
 
COPY --from=builder /usr/local/src/bin/app /
COPY app/.env .

CMD ["/app"]