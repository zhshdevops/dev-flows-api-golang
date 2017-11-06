FROM alpine
MAINTAINER huangxin@tenxcloud.com

## code volume path
VOLUME /app

#RUN mkdir -p /root/.ssh
COPY known_hosts /root/.ssh/

COPY build.sh /
COPY upload.js /
COPY download.js /

## install necessary pacakges
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories \
  && apk update && apk upgrade \
  && apk add --no-cache git openssh-client subversion nodejs nodejs-npm bash \
  && npm install --registry=https://registry.npm.taobao.org urllib \
  && chmod 755 /build.sh && chmod 755 /download.js && chmod 755 /upload.js

CMD ["/build.sh"]
