# Base image to build docker image using TenxCloud CICD service
FROM docker:1.12.1
MAINTAINER wanglei@tenxcloud.com

# For json parse
RUN apk add --update --no-cache curl jq

# Disable the base one
ENTRYPOINT []

Add build.sh /

CMD ["/build.sh"]
