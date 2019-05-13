FROM quay.io/giantswarm/alpine:3.9-giantswarm

USER root

RUN apk add --no-cache ca-certificates

USER giantswarm

ADD ./chart-operator /chart-operator

ENTRYPOINT ["/chart-operator"]
