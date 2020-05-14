FROM ubuntu:20.04

USER root

# bind-tools is required by the init container to use dig.
RUN apt-get update -y && \
    apt-get install --no-install-recommends -y ca-certificates dnsutils && \
    rm -rf /var/lib/apt/lists/*
RUN useradd -ms /bin/bash giantswarm

USER giantswarm

ADD ./chart-operator /chart-operator

ENTRYPOINT ["/chart-operator"]
