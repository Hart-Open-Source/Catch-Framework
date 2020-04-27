#!/bin/bash

sed -i -e "s/\(ELASTIC_PASSWORD:\) elasticpass/\1 $ELASTIC_PASS/g" docker-compose.yml
sed -i -e "s/\(elasticsearch.password:\) elasticpass/\1 $ELASTIC_PASS/g" kibana/config/kibana.yml
sed -i -e "s/\(xpack.monitoring.elasticsearch.password:\) elasticpass/\1 $ELASTIC_PASS/g" logstash/config/logstash.yml
sed -i -e "s/\(password =>\) \"elasticpass\"/\1 \"$ELASTIC_PASS\"/g" logstash/pipeline/60-osquery-output.conf

docker-compose up -d
