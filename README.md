# ELK-Kolide for Osquery

## Overview
A containerized setup for ELK and Kolide Fleet server, using docker-compose because doing this manually is a pain. The Kolide Fleet server is setup to ship logs to the logstash server via Filebeat.

## Setup
### Generate Certificates
Kolide Fleet server needs to be configured to use TLS certificates for communication with Osquery agents. These certificates should be generated and placed within the `kolide/certs` directory.
1. `openssl genrsa -out server.key 4096`
2. `openssl req -new -key server.key -out server.csr`
3. `openssl x509 -req -days 366 -in /tmp/server.csr -signkey server.key -out server.cert`
The `server.cert` certificate will automatically be appended to the Kolide containers's `/etc/ssl/certs/ca-certificates.crt` trusted certificate list during its startup.

### Configure Environment Variables
A number of environment variables need to be set prior to executing `setup.sh`.
1. `export ELK_VERSION=7.6.2`
2. `export MYSQL_PASS=mysqlpass`
3. `export REDIS_PASS=redispass`
4. `export JWT_KEY=jwtkey`
5. `export ELASTIC_PASS=elasticpass`

### Run Startup Script
`chmod +x setup.sh && ./setup.sh`

### Configure Kolide
Kolide needs to be configured after it's container has been launched. Access the Kolide server via `https://kolideserver:8080/` and follow the setup instructions.

### Add Osquery Query Packs
No query packs are installed by default on Kolide. To add query packs to Kolide you'll need to download the fleetctl binary from `https://github.com/kolide/fleet/releases` to your workstation.

Add the generated `server.cert` to your trusted certificate keystore otherwise fleetctl will produce TLS errors while trying to communicate with Kolide. Download the Osquery query pack to the same folder as fleetctl and then run the following:
1. `fleetctl config set --address https://kolideserver:8080`
2. `fleetctl login`
3. `fleetctl apply -f querypack.yaml`

Verify that you can see the installed query pack on the Kolide web interface.
