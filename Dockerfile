FROM demoregistry.dataman-inc.com/library/alpine3.3-base:latest
MAINTAINER Will <zhguo.dataman-inc.com>

RUN mkdir -p /config

ADD config/haproxy_template.cfg /config/haproxy_template.cfg
ADD config/production.json /config/production.json

ADD . /gopath/src/github.com/QubitProducts/bamboo
ADD haproxy /usr/share/haproxy
ADD builder/supervisord.conf /etc/supervisord.conf
ADD builder/run.sh /run.sh
ADD builder/buildBamboo.sh /buildBamboo.sh
WORKDIR /

RUN sh /buildBamboo.sh

EXPOSE 5090

CMD sh /run.sh   
