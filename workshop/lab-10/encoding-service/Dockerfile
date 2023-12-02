FROM golang:1.19

ADD go.mod /go/src/github.com/alevsk/server/go.mod
ADD go.sum /go/src/github.com/alevsk/server/go.sum
WORKDIR /go/src/github.com/alevsk/server/
# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/alevsk/server/
WORKDIR /go/src/github.com/alevsk/server/

ENV CGO_ENABLED=0

RUN go build -ldflags "-w -s" -a -o /bin/encode ./base64/
RUN go build -ldflags "-w -s" -a -o /bin/server ./main.go

CMD ["server"]
