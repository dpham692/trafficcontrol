#!/usr/bin/env bash
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# The following environment variables must be set (ordinarily by `docker run -e` arguments):
# TO_URI
# TO_USER
# TO_PASS
# CDN
# Check that env vars are set
envvars=(TESTTO_URI TESTTO_PORT TESTCACHES_URI TESTCACHES_PORT_START TM_URI)
for v in $envvars; do
  if [[ -z ${!v} ]]; then
    echo "$v is unset"
    exit 1
  fi
done

TO_API_VERSION=4.0
CFG_FILE=/tm/traffic-monitor-integration-test.cfg

start() {
  printf "DEBUG traffic_monitor_integration starting\n"

  exec /traffic_monitor_integration_test -test.v -cfg $CFG_FILE
}

init() {
  wait_for_endpoint "${TESTTO_URI}/api/${TO_API_VERSION}/servers"
  wait_for_endpoint "${TESTCACHES_URI}:${TESTCACHES_PORT_START}/_astats"
  wait_for_endpoint "${TM_URI}"
  TESTCACHES_ADDRESS=$(ping testcaches -4 -c 1 | head -n 1 | grep -Eo '[0-9]+.[0-9]+.[0-9]+.[0-9]+')
  TESTCACHES_GATEWAY=$(echo $TESTCACHES_ADDRESS | sed "s/\([0-9]\+.[0-9]\+.[0-9]\+.\)[0-9]/\11/")

  jq "(.. | .address?) |= \"$TESTCACHES_ADDRESS\" | (.. | .gateway?) |= \"$TESTCACHES_GATEWAY\"" \
    /tm/monitoring.json > /tm/monitoring.json.tmp && mv /tm/monitoring.json.tmp /tm/monitoring.json

  curl -Lvsk ${TESTTO_URI}/api/${TO_API_VERSION}/cdns/fake/snapshot -X POST -d "@/tm/snapshot.json"

  curl -Lvsk ${TESTTO_URI}/api/${TO_API_VERSION}/cdns/fake/configs/monitoring -X POST -d '@/tm/monitoring.json'

  curl -Lvsk ${TESTTO_URI}/api/${TO_API_VERSION}/servers -X POST -d '
[
  {
    "cachegroup": "foo",
    "cachegroupId": 0,
    "cdnId": 1,
    "cdnName": "fake",
    "deliveryServices": null,
		"fqdn": "trafficmonitor.traffic-monitor-integration.test",
    "guid": "foo",
    "hostName": "trafficmonitor",
    "httpsPort": null,
    "id": 1,
    "iloIpAddress": null,
    "iloIpGateway": null,
    "iloIpNetmask": null,
    "iloPassword": null,
    "iloUsername": null,
    "interfaceMtu": null,
    "interfaceName": "bond0",
    "ip6Address": null,
    "ip6Gateway": null,
    "interfaces": [
      {
        "ipAddresses": [
          {
            "address": "4.0.16.239.6",
            "gateway": "4.0.16.239.1",
            "serviceAddress": true
          },
          {
            "address": "fc01:9400:1000:8::6",
            "gateway": "fc01:9400:1000:8::1",
            "serviceAddress": true
          }
        ],
        "maxBandwidth": null,
        "monitor": true,
        "mtu": 1500,
        "name": "eth0"
      }
    ],
    "ipGateway": "4.0.0.0.1",
    "ipNetmask": "255.255.255.0",
    "lastUpdated": "2019",
    "mgmtIpAddress": null,
    "mgmtIpGateway": null,
    "mgmtIpNetmask": null,
    "offlineReason": "none",
    "physLocation": "",
    "physLocationId": 0,
    "profile": "Monitor0",
    "profileDesc": "nodesc",
    "profileId": 0,
    "rack": "",
    "revalPending": false,
    "routerHostName": "",
    "routerPortName": "",
    "status": "REPORTED",
    "statusId": 0,
    "tcpPort": 80,
    "type": "RASCAL",
    "typeId": 0,
    "updPending": false,
    "xmppId": "",
    "xmppPasswd": ""
  }
]
'

  cat >$CFG_FILE <<-EOF
{
  "trafficMonitor": {
    "url": "$TM_URI"
  },
  "default": {
    "session": {
      "timeoutInSecs": 30
    },
    "log": {
      "debug": "stdout",
      "event": "stdout",
      "info": "stdout",
      "error": "stdout",
      "warning": "stdout"
    }
  }
}
EOF

  echo "INITIALIZED=1" >>/etc/environment
}

function wait_for_endpoint() {
  try=0
  while [ $(curl -Lsk --write-out "%{http_code}" "$1" -o /dev/null) -ne 200 ] ; do
    echo "Waiting for $1 to return a 200 OK"
    try=$(expr $try + 1)
    if [[ $try -gt 5 ]]; then
      echo "Unable to get $1"
      exit 1
    fi
    sleep 5
  done
}
export -f wait_for_endpoint

source /etc/environment
if [ -z "$INITIALIZED" ]; then init; fi
start
