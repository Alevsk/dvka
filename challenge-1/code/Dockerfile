FROM golang:1.21.5-bullseye as golayer

RUN apt-get update && apt-get install -y ca-certificates

ADD go.mod /go/src/github.com/Alevsk/dvka/challenge-1/code/go.mod
ADD go.sum /go/src/github.com/Alevsk/dvka/challenge-1/code/go.sum

WORKDIR /go/src/github.com/Alevsk/dvka/challenge-1/code/

RUN go mod download

ADD . /go/src/github.com/Alevsk/dvka/challenge-1/code/

WORKDIR /go/src/github.com/Alevsk/dvka/challenge-1/code/

ENV CGO_ENABLED=0

RUN go build -ldflags "-w -s" -a -o lab1 ./cmd/app

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.3

LABEL org.opencontainers.image.authors="lenin@alevsk.com"

EXPOSE 8080

WORKDIR /app

COPY --from=golayer /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=golayer /go/src/github.com/Alevsk/dvka/challenge-1/code/cmd/app/static/ /app/static/
COPY --from=golayer /go/src/github.com/Alevsk/dvka/challenge-1/code/lab1 /app/

ENTRYPOINT ["/app/lab1"]
