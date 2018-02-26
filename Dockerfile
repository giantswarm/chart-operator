FROM alpine:3.7

RUN apk add --no-cache ca-certificates

ADD ./chart-operator /chart-operator

ENTRYPOINT ["/chart-operator"]
