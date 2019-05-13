FROM alpine:3.9-giantswarm

RUN apk add --no-cache ca-certificates

ADD ./chart-operator /chart-operator

ENTRYPOINT ["/chart-operator"]
