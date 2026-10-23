FROM gsoci.azurecr.io/giantswarm/alpine:3.24.2

# bind-tools is required by the init container to use dig.
RUN apk add --no-cache ca-certificates bind-tools

ARG TARGETARCH
COPY chart-operator-linux-${TARGETARCH} /chart-operator

USER 1000:1000

ENTRYPOINT ["/chart-operator"]
