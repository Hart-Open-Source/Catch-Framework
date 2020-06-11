#!/bin/bash

ERROR=0
OSQUERYLOGPATH="/var/log/osquery"
OSQUERYPATH="/var/osquery"
ENROLLSECRET="enrollsecret"
ENROLLSECRETPATH="/var/osquery/enroll_secret"
OSXVERSION=$( sw_vers | grep ProductVersion | cut -d'.' -f2 )
TMPOSQUERY="/tmp/osquery.pkg"
OSQUERYSERVERPEM="/var/osquery/server.pem"
OSQUERYCONF="/var/osquery/osquery.conf"
OSQUERYPLIST="/var/osquery/com.facebook.osqueryd.plist"


if [ $OSXVERSION -eq 12 ]
then 
    curl -o $TMPOSQUERY https://pkg.osquery.io/darwin/osquery-3.3.0.pkg
else 
    curl -o $TMPOSQUERY https://pkg.osquery.io/darwin/osquery-4.0.2.pkg
fi

if [ ! -f $TMPOSQUERY ]
then
    exit 1
fi

installer -pkg $TMPOSQUERY -target /
mkdir $OSQUERYLOGPATH

if [ ! -d $OSQUERYPATH ]
then
    exit 1
fi

ln -s $OSQUERYPATH /usr/local/share/osquery
cp /var/osquery/osquery.example.conf $OSQUERYCONF
echo ENROLLSECRET > $ENROLLSECRETPATH
cp server.pem $OSQUERYSERVERPEM
cp com.facebook.osqueryd.plist $OSQUERYPLIST



PLIST=$(cat << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>KeepAlive</key>
  <true/>
  <key>Disabled</key>
  <false/>
  <key>Label</key>
  <string>com.facebook.osqueryd</string>
  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/osqueryd</string>
    <string>--enroll_secret_path=/var/osquery/enroll_secret</string>
    <string>--tls_server_certs=/var/osquery/server.pem</string>
    <string>--tls_hostname=kolide.local:8080</string>
    <string>--host_identifier=hostname</string>
    <string>--enroll_tls_endpoint=/api/v1/osquery/enroll</string>
    <string>--config_plugin=tls</string>
    <string>--config_tls_endpoint=/api/v1/osquery/config</string>
    <string>--config_tls_refresh=10</string>
    <string>--disable_distributed=false</string>
    <string>--distributed_plugin=tls</string>
    <string>--distributed_interval=3</string>
    <string>--distributed_tls_max_attempts=3</string>
    <string>--distributed_tls_read_endpoint=/api/v1/osquery/distributed/read</string>
    <string>--distributed_tls_write_endpoint=/api/v1/osquery/distributed/write</string>
    <string>--logger_plugin=tls</string>
    <string>--logger_tls_endpoint=/api/v1/osquery/log</string>
    <string>--logger_tls_period=10</string>
    <string>--allow_unsafe=true</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>ThrottleInterval</key>
  <integer>60</integer>
</dict>
</plist>
EOF
)
exit $ERROR

echo "Performing install" > /tmp/postinstall.log
curl -o /tmp/osquery-4.0.2.pkg https://pkg.osquery.io/darwin/osquery-4.0.2.pkg
installer -pkg /tmp/osquery-4.0.2.pkg -target /
ln -s /var/osquery /usr/local/share/osquery
mkdir /var/log/osquery
cp /var/osquery/osquery.example.conf /var/osquery/osquery.conf
echo "enrollsecret" > /var/osquery/enroll_secret
cp server.pem /var/osquery/server.pem
cp com.facebook.osqueryd.plist /var/osquery/com.facebook.osqueryd.plist
osqueryctl start
exit 0