FROM gsoci.azurecr.io/giantswarm/alpine:3.20.3-giantswarm

USER root

# bind-tools is required by the init container to use dig.
RUN apk add --no-cache ca-certificates bind-tools

USER giantswarm

ADD ./chart-operator /chart-operator

ENTRYPOINT ["/chart-operator"]
