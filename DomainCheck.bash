#!/bin/bash
# This is meant to be an interim script to (1) evaluate we have
# enough capabilities inside of the urlgetter pseudo-test and (2)
# prototype the next generation Web Connectivity.
#
# Very alpha. Very basic. Needs much more love.
#
# Inspired by great chats and great code. Main source of ispiration is
# currently https://github.com/Jigsaw-Code/net-analysis/blob/master/netanalysis/blocktest/measure.sh

if [ $# -ne 1 ]; then
  echo "usage: $0 domain" 1>&2
  exit 1
fi
domain=$1

function log() {
  echo "$@" 1>&2
}

function run() {
  log $@
  $@
}

function getfailure() {
  tail -n1 report.jsonl|jq -r .test_keys.failure
}

function output() {
  echo "$@"
}

function getipv4s() {
  echo $(tail -n1 report.jsonl|jq -r ".test_keys.queries|.[]|select(.hostname==\"$1\")|select(.query_type==\"A\")|.answers|.[].ipv4"|sort)
}

function getbody() {
  tail -n1 report.jsonl|jq -r ".test_keys.requests|.[0]|.respone.body"
}

log "* checking for DNS injection"
run ./miniooni -OResolverURL=udp://example.com:53 -i dnslookup://$domain urlgetter
if [ "$(getfailure)" = "null" ]; then
  output "MINIOONI_DNS_INJECTION=1"
fi

log "* checking for DNS bogons returned by the system resolver"
run ./miniooni -OResolverURL=system:/// -ORejectDNSBogons=true -i dnslookup://$domain urlgetter
if [ "$(getfailure)" != "null" ]; then
  output "MINIOONI_DNS_BOGONS=1"
fi

log "* checking for DNS consistency"
run ./miniooni -OResolverURL=system:/// -i dnslookup://$domain urlgetter
ipv4_system_rv=$(getfailure)
ipv4_system_list=$(getipv4s $domain)
output "MINIOONI_DNS_SYSTEM_IPV4=$ipv4_system_list"
run ./miniooni -OResolverURL=doh://google -i dnslookup://$domain urlgetter
ipv4_doh_rv=$(getfailure)
ipv4_doh_list=$(getipv4s $domain)
output "MINIOONI_DNS_DOH_IPV4=$ipv4_doh_list"
if [ "$ipv4_system_list" == "$ipv4_doh_list" ]; then
  output "MINIOONI_DNS_CONSISTENCY=1"
fi

log "* checking for SNI triggered censorship"
run ./miniooni -OTLSServerName=$domain -i tlshandshake://example.com:443 urlgetter
if [ "$(getfailure)" != "ssl_invalid_hostname" ]; then
  output "MINIOONI_SNI_BLOCKING=1"
fi

log "* checking for Host header triggered censorship"
run ./miniooni -OHTTPHost=$domain -ONoFollowRedirects=true -i http://example.com urlgetter
if [ "$(getfailure)" != "null" ]; then
  output "MINIOONI_HOST_BLOCKING=1"
fi

log "* checking for HTTP requests consistency"
run ./miniooni -i http://$domain urlgetter
body_vanilla=$(getbody)
run ./miniooni -OTunnel=psiphon -i http://$domain urlgetter
body_tunnel=$(getbody)
if [ "$body_vanilla" == "$body_tunnel" ]; then
  output "MINIOONI_BODY_CONSISTENCY=1"
fi