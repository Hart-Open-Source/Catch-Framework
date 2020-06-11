# ELK-Kolide for Osquery

## Overview
A containerized setup for ELK, Kolide Fleet server and Catch, for automated HITRUST security audits against servers/containers and MacOS workstations.

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
6. `export JIRA_URL=https://jira.local`
7. `export JIRA_USER=jirauser`
8. `export JIRA_PASSWORD=jirapass`

### Run Startup Script
`chmod +x setup.sh && ./setup.sh`

### Configure Kolide
Kolide needs to be configured after it's container has been launched. Access the Kolide server via `https://kolideserver:8080/` and follow the setup instructions.

### Add Osquery Query Packs
No query packs are installed by default on Kolide. To add query packs to Kolide you'll need to download the fleetctl binary from `https://github.com/kolide/fleet/releases` to your workstation.

Add the generated `server.cert` to your trusted certificate keystore otherwise fleetctl will produce TLS errors while trying to communicate with Kolide. Copy the Osquery query pack to the same folder as fleetctl and then run the following:
1. `fleetctl config set --address https://kolideserver:8080`
2. `fleetctl login`
3. `mkdir ~/querypacks`
4. `cp elk-kolide-osquery/catch/osquery_packs/servers/hitrust-ubuntu-containers-pack.conf ~/querypacks`
5. `fleetctl convert -f ~/querypacks/hitrust-ubuntu-containers-pack.conf > ~/querypacks/hitrust-ubuntu-containers-pack.yaml`
6. `fleetctl apply -f ~/querypacks/hitrust-ubuntu-containers-pack.yaml`

Verify that you can see the installed query pack on the Kolide web interface under the packs section. Then select the uploaded pack in Kolide and choose "edit pack". Edit the target hosts you'd like the query pack to be applied to and save. This will push the query pack down to the selected target hosts' osquery agents and configure them to be used.

## Running a Security Audit
Currently, Catch is configured to do HITRUST security audits for servers/containers and MacOS workstations. However, the configuration files have been designed in a versatile way that new query packs can be created for just about any audting standard that can be measured on hosts via osquery. Catch will load all server configurations in the `catch/osquery_packs/servers/` path and workstations configurations in the `catch/osquery_packs/workstations/` path.

### Servers/Containers Audit
`http://catch.local:9090/audit?filter=servers`

### MacOS Workstations Audit
`http://catch.local:9090/audit?filter=workstations`

### Promethius Metrics
A Prometheius metrics endpoint has been included at `http://catch.local:9090/metrics`

### Generating Jira Tickets
An audit can automatically generate Jira ticket for each failed HITRUST control reference for each host, by appending `&jira=1` to the audit URL. Ensure that the Jira environmental variables have been set for authentication to the Jira server to use this functionality. For example:
`http://catch.local:9090/audit?filter=servers&jira=1`
