#!/bin/bash

print_usage() {
    echo "usage:"
    echo "$0 -u <username> -p <password> -H <db hostname> -P <db port>"
}

while getopts "u:p:H:P:h" opt; do
    case "${opt}" in
        u) db_username=${OPTARG};;
        p) db_password=${OPTARG};;
        H) db_hostname=${OPTARG};;
        P) db_port=${OPTARG};;
        h)
            print_usage
            exit 0
            ;;
    esac
done

Config_Devops_Server() {
    # if [ -n "$DEPLOYMENT_MODE" ]; then
    #     sed -i 's/[[:blank:]]*deployment_mode.*/deployment_mode=standard/' /api-server/conf/app.conf
    # fi
    #RunMode = dev HTTPPort = 8000EnableHTTPS
    #deployment_mode = enterprise
    sed -i "s/[[:blank:]]*deployment_mode[[:blank:]]*=.*/deployment_mode=enterprise/"  /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*EnableHTTP[[:blank:]]*=.*/EnableHTTP=true/"             /usr/src/cicd/conf/app.conf
     sed -i "s/[[:blank:]]*EnableHTTPS[[:blank:]]*=.*/EnableHTTPS=false/"         /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*HTTPSPort.*/HTTPPort=8090/"                          /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*HTTPPort[[:blank:]]*=.*/HTTPPort=8090/"             /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*RunMode.*/RunMode=pro/"                    /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*db_port.*/db_port=${db_port}/"             /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*db_host.*/db_host=${db_hostname}/"         /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*db_user.*/db_user=${db_username}/"         /usr/src/cicd/conf/app.conf
    sed -i "s/[[:blank:]]*db_password.*/db_password=${db_password}/" /usr/src/cicd/conf/app.conf
}




if [ ! -d "/var/lib/mysql/" ]; then
    mkdir -p /var/lib/mysql/
fi

if [ -z "${db_username}" ]; then
    echo "use default db username"
    db_username="tcepaas"
fi

if [ -z "${db_password}" ]; then
    echo "use default db password"
    db_password="xboU58vQbAbN"
fi

local_db_hostname="10.39.0.251"
if [ -z "${db_hostname}" ]; then
    echo "use default db hostname"
    db_hostname="${local_db_hostname}"
fi

if [ -z "${db_port}" ]; then
    echo "use default db port"
    db_port="3306"
fi

echo db_username ${db_username}
echo db_password ${db_password}
echo db_hostname ${db_hostname}
echo db_port ${db_port}

Config_Devops_Server

supervisord_template=$(cat <<EOF
[supervisord]
# run in foreground
nodaemon = true
pidfile = /tmp/supervisord.pid
logfile = /tmp/supervisord.log

[inet_http_server]
port = 0.0.0.0:60000


[program:dev-flows-api-golang]
command=/usr/src/cicd/dev-flows-api-golang

startretries=99999
exitcodes=0
redirect_stderr=true
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5

stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5


EOF
        )

echo "${supervisord_template}" > supervisord.conf


supervisord -c supervisord.conf