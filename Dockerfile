FROM alpine:3.8

RUN apk add --no-cache ca-certificates

ADD ./chart-operator /chart-operator

ENTRYPOINT ["/chart-operator"]
