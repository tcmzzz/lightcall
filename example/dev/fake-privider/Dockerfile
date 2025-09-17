FROM ghcr.io/tcmzzz/freeswitch:v1.10.9

RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
RUN apt-get -yq install rsync

RUN rm -rf /usr/local/freeswitch/scripts/*

COPY conf /tmp/conf
RUN rsync -av /tmp/conf/ /usr/local/freeswitch/conf/

RUN rm -f /entrypoint.sh
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
