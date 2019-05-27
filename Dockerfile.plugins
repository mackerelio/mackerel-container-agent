ARG REPOSITORY
ARG TAG

FROM $REPOSITORY:$TAG

ENV DEBIAN_FRONTEND noninteractive
ENV BUNDLE_AGENT_PLUGINS apache2|elasticsearch|fluentd|gostats|haproxy|jmx-jolokia|memcached|mysql|nginx|php-apc|php-fpm|php-opcache|plack|postgres|redis|sidekiq|snmp|squid|uwsgi-vassal
ENV BUNDLE_CHECK_PLUGINS cert-file|elasticsearch|file-age|file-size|http|jmx-jolokia|log|memcached|mysql|postgresql|redis|ssh|ssl-cert|tcp

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends curl gnupg2
RUN echo "deb [arch=amd64] http://apt.mackerel.io/v2/ mackerel contrib" > /etc/apt/sources.list.d/mackerel.list
RUN curl -LfsS https://mackerel.io/file/cert/GPG-KEY-mackerel-v2 | apt-key add -

RUN apt-get update -yq && \
    apt-get install -yq --no-install-recommends mackerel-agent-plugins mackerel-check-plugins && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists

RUN find /usr/bin/ -type l -regextype posix-egrep -name 'mackerel-plugin-*' -a ! -regex ".*mackerel-plugin-(${BUNDLE_AGENT_PLUGINS})" -delete
RUN find /usr/bin/ -type l -regextype posix-egrep -name 'check-*' -a ! -regex ".*check-(${BUNDLE_CHECK_PLUGINS})" -delete
