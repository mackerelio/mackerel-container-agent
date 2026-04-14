FROM golang:1.26-trixie AS builder

WORKDIR /go/src/app

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    --mount=type=bind,target=. \
    go build -o /usr/local/bin/mackerel-container-agent ./cmd/mackerel-container-agent/...

FROM debian:trixie-slim AS container-agent

ENV DEBIAN_FRONTEND=noninteractive
ENV GODEBUG=http2client=0

RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update -yq && \
    apt-get install -yq --no-install-recommends ca-certificates sudo

COPY --from=builder /usr/local/bin/mackerel-container-agent /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/mackerel-container-agent"]

FROM golang:1.26-trixie AS plugins-builder

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=plugins/go.sum,target=go.sum \
    --mount=type=bind,source=plugins/go.mod,target=go.mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=plugins/go.mod,target=go.mod \
    --mount=type=bind,source=plugins/go.sum,target=go.sum \
    go install \
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
