FROM harbor.enncloud.cn/paas/paas-tce:2.2.0.pro.harbor
#FROM harbor.enncloud.cn/qinzhao-harbor/cicd-base-image:v1.0
ADD conf /usr/src/cicd/conf
ADD dev-flows-api-golang /usr/src/cicd/
ADD run.sh /
# /etc/localtime
##============DB 环境变量设置
ENV DB_HOST 10.39.0.102
ENV DB_PORT 3306
ENV DB_NAME tenxcloud_2_0
ENV DB_USER root
ENV DB_PASSWORD enN12345
ENV CICD_REPO_CLONE_IMAGE qinzhao-harbor/clone-repo:v2.2
ENV CICD_IMAGE_BUILDER_IMAGE enncloud/image-builder:v2.2
# github config 默认是paasdev
#ENV GITHUB_REDIRECT_URL https://paasdev.enncloud.cn/api/v2/devops/repos/github/auth-callback  
#ENV CLIENT_ID 43b9c69e79ae49f32919
#ENV CLIENT_SECRET bf1c5dfd9ae48081073d786e3797cb08a1f0b59e
# 指定到某个节点构建以及协议相关
ENV BIND_BUILD_NODE true #开启指定到某个节点构建CICD
ENV DEVOPS_PROTOCOL http
ENV DEVOPS_EXTERNAL_PROTOCOL http
ENV DEVOPS_HOST 10.39.0.102:48090
ENV DEVOPS_EXTERNAL_HOST 10.39.0.102:48090
# 服务暴露的端口
ENV SERVER_PORT 8090
ENV EXTERNAL_ES_URL http://paasdev.enncloud.cn:9200
ENV USERPORTAL_URL https://paasdev.enncloud.cn
WORKDIR /
EXPOSE 8090
CMD ["./run.sh"]
               
