#!/bin/bash
# This is meant to be an interim script to (1) evaluate we have
# enough capabilities inside of the urlgetter pseudo-test and (2)
# prototype the next generation Web Connectivity.
#
# Very alpha. Very basic. Needs much more love.
#
# Inspired by great chats and great code. Main source of ispiration is
# currently https://github.com/Jigsaw-Code/net-analysis/blob/master/netanalysis/blocktest/measure.sh

function fatal() {
  echo "FATAL: $@" 1>&2
  exit 1
}

if [ $# -ne 1 ]; then
  fatal "usage: $0 domain"
fi
domain=$1

function require() {
  if ! [ -x "$(command -v $1)" ]; then
    fatal "$1 not installed, please run: $2"
  fi
}

require jq "sudo apt install jq"
require uuidgen "sudo apt install uuid-runtime"

uuid=$(uuidgen)

function log() {
  echo "$@" 1>&2
}

function run() {
  log $@
  $@
}

function urlgetter() {
  #run ./miniooni -A session=$uuid $@ urlgetter
  run ./miniooni -n --no-bouncer -A session=$uuid $@ urlgetter
}

function getfailure() {
  tail -n1 report.jsonl|jq -r .test_keys.failure
}

function output() {
  echo "$@"
}

function getipv4first() {
  tail -n1 report.jsonl|jq -r ".test_keys.queries|.[]|select(.hostname==\"$1\")|select(.query_type==\"A\")|.answers|.[0].ipv4"
}

function getipv4list() {
  tail -n1 report.jsonl|jq -r ".test_keys.queries|.[]|select(.hostname==\"$1\")|select(.query_type==\"A\")|.answers|.[].ipv4"|sort
}

function getbody() {
  # Implementation note: requests stored in LIFO order
  tail -n1 report.jsonl|jq -r ".test_keys.requests|.[0]|.response.body"
}

testdomain=${MINIOONI_TEST_DOMAIN:-example.org}
dohurl=${MINIOONI_DOH_URL:-doh://google}

log "* getting IP address of $testdomain"
urlgetter -OResolverURL=$dohurl -i dnslookup://$testdomain
if [ "$(getfailure)" != "null" ]; then
  fatal "cannot determine IP address of $testdomain"
fi
testip=$(getipv4first $testdomain)
if [ -z "$testip" ]; then
  fatal "no available IPv4 for $testdomain"
fi
output "MINIOONI_DOH_URL=$dohurl"
output "MINIOONI_TEST_DOMAIN=$testdomain"
output "MINIOONI_TEST_IP=$testip"

log "* checking for sni-triggered blocking"
urlgetter -OTLSServerName=$domain -i tlshandshake://$testip:443
if [ "$(getfailure)" != "ssl_invalid_hostname" ]; then
  output "MINIOONI_SNI_BLOCKING=1"
fi

log "* checking for host-header-triggered censorship"
urlgetter -OHTTPHost=$domain -ONoFollowRedirects=true -i http://$testip
if [ "$(getfailure)" != "null" ]; then
  output "MINIOONI_HOST_BLOCKING=1"
fi

log "* checking for DNS injection"
urlgetter -OResolverURL=udp://$testip:53 -i dnslookup://$domain
if [ "$(getfailure)" = "null" ]; then
  output "MINIOONI_DNS_INJECTION=1"
fi

log "* resolving $domain using the system resolver"
# Implementation note: this saves the IPs _and_ records the error.
urlgetter -OResolverURL=system:/// -ORejectDNSBogons=true -i dnslookup://$domain
if [ "$(getfailure)" = "dns_bogon_error" ]; then
  output "MINIOONI_DNS_BOGONS=1"
fi
ipv4_system_list=$(mktemp $TMPDIR/miniooni-domain-check.XXXXXX)
log $ipv4_system_list
getipv4list $domain > $ipv4_system_list
log $(cat $ipv4_system_list)

log "* resolving $domain using a DoH resolver"
urlgetter -OResolverURL=$dohurl -i dnslookup://$domain
ipv4_doh_list=$(mktemp $TMPDIR/miniooni-domain-check.XXXXXX)
log $ipv4_doh_list
getipv4list $domain > $ipv4_doh_list
log $(cat $ipv4_doh_list)

log "* checking for DNS consistency"
ipv4_overlap_list=$(comm -13 $ipv4_system_list $ipv4_doh_list)
if [ "$ipv4_overlap_list" != "" ]; then
  output "MINIOONI_DNS_CONSISTENCY=1"
fi

log "* selecting IP address provided by system resolver"
ipv4_system_candidate=$(cat $ipv4_system_list|sort -uR|head -n1)
log $ipv4_system_candidate

log "* selecting IP address provided by DoH resolver"
ipv4_doh_candidate=$(cat $ipv4_doh_list|sort -uR|head -n1)
log $ipv4_doh_candidate

exit 0

if [ "$ipv4_system_candidate" != "" ]; then
  log "* using $ipv4_system_candidate as a server for $domain"
  urlgetter -ODNSCache="$ipv4_system_candidate $domain" -ONoTLSVerify=true -i http://$domain
fi

log "* checking for HTTP consistency"
urlgetter -ONoFollowRedirects=true -i http://$domain
body_vanilla=$(getbody)
log $body_vanilla
urlgetter -ONoFollowRedirects=true -OTunnel=psiphon -i http://$domain
body_tunnel=$(getbody)
log $body_tunnel
if [ "$body_vanilla" == "$body_tunnel" ]; then
  output "MINIOONI_HTTP_CONSISTENCY=1"
fi
