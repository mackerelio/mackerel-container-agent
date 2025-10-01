FROM golang:1.24.4-bookworm AS builder

WORKDIR /go/src/app

COPY go.sum go.mod ./
RUN go mod download

COPY . .
RUN make build

FROM debian:bookworm-slim AS container-agent

ENV DEBIAN_FRONTEND=noninteractive
ENV GODEBUG=http2client=0

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends ca-certificates sudo && \
    rm -rf /var/lib/apt/lists

COPY --from=builder /go/src/app/build/mackerel-container-agent /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/mackerel-container-agent"]

FROM golang:1.24.4-bookworm AS plugins-builder

COPY plugins/go.sum plugins/go.mod ./
RUN go mod download

RUN go install \
    github.com/mackerelio/go-check-plugins \
    github.com/mackerelio/mackerel-agent-plugins \
    github.com/mackerelio/mackerel-plugin-json \
    github.com/mackerelio/mkr

FROM container-agent AS container-agent-with-plugins

# for compat. deb packages installed path.
COPY --from=plugins-builder /go/bin/go-check-plugins /usr/bin/mackerel-check
COPY --from=plugins-builder /go/bin/mackerel-agent-plugins /usr/bin/mackerel-plugin
COPY --from=plugins-builder /go/bin/mackerel-plugin-json /opt/mackerel-agent/plugins/bin/mackerel-plugin-json
COPY --from=plugins-builder /go/bin/mkr /usr/bin/mkr

ENV PATH=$PATH:/opt/mackerel-agent/plugins/bin

RUN /bin/bash -c 'cd /usr/bin; for i in apache2 elasticsearch fluentd gostats haproxy jmx-jolokia memcached mysql nginx php-apc php-fpm php-opcache plack postgres redis sidekiq snmp squid uwsgi-vassal;do ln -s ./mackerel-plugin mackerel-plugin-$i; done'
RUN /bin/bash -c 'cd /usr/bin; for i in cert-file elasticsearch file-age file-size http jmx-jolokia log memcached mysql postgresql redis ssh ssl-cert tcp;do ln -s ./mackerel-check check-$i; done'
