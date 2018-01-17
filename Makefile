PREFIX = harbor.enncloud.cn/qinzhao-harbor
TAG = v1.4
db_url=10.39.0.251
vip_url=10.39.0.102

SOURCE_DIR=$(shell pwd)
PROJECT=$(shell basename $(SOURCE_DIR))
IMAGE=$(PREFIX)/cicd:$(TAG)

.PHONY: all build clean

all: build

build: clean
	docker run --rm -v $(SOURCE_DIR):/go/src/$(PROJECT) -w /go/src/$(PROJECT) golang:1.8.3 \
	/bin/sh -c "CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -gcflags \"-N -l\""

gorun:  
	CGO_ENABLED=0 bee run

image: build
	docker build -t $(IMAGE) -f Dockerfile .
	echo "docker build  $(IMAGE) success"
	docker push $(IMAGE)
	echo "docker push  $(IMAGE) success"

run:
	docker run  --name cicd \
	       -v /paas/tce/logs:/tmp \
               -v /etc/localtime:/etc/localtime  \
               -p 48090:8090 \
               -e DB_HOST=$(db_url) \
               -e DB_PORT=3306 \
               -e DB_NAME=tenxcloud_2_0 \
               -e DB_USER=tcepaas \
               -e DB_PASSWORD=xboU58vQbAbN \
               -e BIND_BUILD_NODE=true \
               -e DEVOPS_EXTERNAL_PROTOCOL=http \
               -e DEVOPS_HOST=$(vip_url):48090 \
               -e DEVOPS_EXTERNAL_HOST=$(vip_url):48090 \
               -e EXTERNAL_ES_URL=http://paasdev.enncloud.cn:9200 \
               -e USERPORTAL_URL=http://$(vip_url) \
               -e FLOW_DETAIL_URL=http://$(vip_url) \
               -e CICD_IMAGE_BUILDER_IMAGE=enncloud/image-builder:v2.2 \
               -e CICD_REPO_CLONE_IMAGE=qinzhao-harbor/clone-repo:v2.2 \
               -d ${IMAGE} \
          /run.sh -u tcepaas -p xboU58vQbAbN -H $(db_url) -P 3306
	        

clean:
	rm -f dev-flows-api-golang  >/dev/null

