FROM golang:1.12 AS builder
WORKDIR /src

ADD go.mod go.sum main.go /src/
ADD db db

RUN go build -o main

# FROM gcr.io/distroless/base
FROM golang:1.12
COPY --from=builder /src/main /src/main
CMD /src/main --phishURL=https://go-phishing.herokuapp.com --port=:${PORT}
