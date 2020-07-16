FROM debian:stretch-slim

ENV DEBIAN_FRONTEND noninteractive
ENV GODEBUG http2client=0

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends ca-certificates sudo && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists

COPY build/mackerel-container-agent /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/mackerel-container-agent"]
