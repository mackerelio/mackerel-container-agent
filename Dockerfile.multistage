FROM golang:1.17-stretch AS builder

WORKDIR /go/src/app

COPY go.sum go.mod ./
RUN go mod download

COPY . .
RUN make build

FROM debian:stretch-slim AS container-agent

ENV DEBIAN_FRONTEND noninteractive
ENV GODEBUG http2client=0

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends ca-certificates sudo && \
    rm -rf /var/lib/apt/lists

COPY --from=builder /go/src/app/build/mackerel-container-agent /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/mackerel-container-agent"]

FROM container-agent AS container-agent-with-plugins

ENV BUNDLE_AGENT_PLUGINS apache2|elasticsearch|fluentd|gostats|haproxy|jmx-jolokia|memcached|mysql|nginx|php-apc|php-fpm|php-opcache|plack|postgres|redis|sidekiq|snmp|squid|uwsgi-vassal
ENV BUNDLE_CHECK_PLUGINS cert-file|elasticsearch|file-age|file-size|http|jmx-jolokia|log|memcached|mysql|postgresql|redis|ssh|ssl-cert|tcp
ENV MKR_INSTALL_PLUGINS json

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends curl gnupg2
RUN echo "deb [arch=amd64,arm64] http://apt.mackerel.io/v2/ mackerel contrib" > /etc/apt/sources.list.d/mackerel.list
RUN curl -LfsS https://mackerel.io/file/cert/GPG-KEY-mackerel-v2 | apt-key add -

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends mackerel-agent-plugins mackerel-check-plugins mkr && \
    rm -rf /var/lib/apt/lists

RUN find /usr/bin/ -type l -regextype posix-egrep -name 'mackerel-plugin-*' -a ! -regex ".*mackerel-plugin-(${BUNDLE_AGENT_PLUGINS})" -delete
RUN find /usr/bin/ -type l -regextype posix-egrep -name 'check-*' -a ! -regex ".*check-(${BUNDLE_CHECK_PLUGINS})" -delete

RUN echo ${MKR_INSTALL_PLUGINS} | tr ' ' '\n' | xargs -I@ mkr plugin install mackerel-plugin-@
ENV PATH $PATH:/opt/mackerel-agent/plugins/bin

