FROM debian:stretch-slim

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends ca-certificates sudo

COPY build/mackerel-container-agent /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/mackerel-container-agent"]
