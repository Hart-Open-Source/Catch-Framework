#!/bin/sh

sleep 10
cat /certs/server.cert >> /etc/ssl/certs/ca-certificates.crt 
update-ca-certificates
echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" > /etc/apk/repositories
apk update
apk add filebeat
filebeat -c /filebeat/filebeat.yml &
fleet prepare db
fleet serve