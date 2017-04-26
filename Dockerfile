FROM golang:onbuild

MAINTAINER Caesar Kabalan <caesar.kabalan@gmail.com>

ADD . /go/src/github.com/celestialstats/clientdiscord

RUN go install github.com/celestialstats/clientdiscord

CMD [ "/go/bin/clientdiscord" ]
