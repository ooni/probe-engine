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
  log "$@"
  "$@"
}

function urlgetter() {
  #run ./miniooni -A session=$uuid $@ urlgetter
  run ./miniooni -n --no-bouncer -A session=$uuid "$@" urlgetter
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

log "* getting IP address of $testdomain"
urlgetter -ODNSCache="8.8.4.4 dns.google" -OResolverURL=doh://google -i dnslookup://$testdomain
if [ "$(getfailure)" != "null" ]; then
  fatal "cannot determine IP address of $testdomain"
fi
testip=$(getipv4first $testdomain)
if [ -z "$testip" ]; then
  fatal "no available IPv4 for $testdomain"
fi
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
# XXX: maybe make these two checks separate checks for now?
urlgetter -OResolverURL=system:/// -ORejectDNSBogons=true -i dnslookup://$domain
if [ "$(getfailure)" = "dns_bogon_error" ]; then
  output "MINIOONI_DNS_BOGONS=1"
fi
ipv4_system_list=$(mktemp $TMPDIR/miniooni-domain-check.XXXXXX)
log $ipv4_system_list
getipv4list $domain > $ipv4_system_list
log $(cat $ipv4_system_list)

log "* resolving $domain using a DoH resolver"
urlgetter -ODNSCache="8.8.4.4 dns.google" -OResolverURL=doh://google -i dnslookup://$domain
ipv4_doh_list=$(mktemp $TMPDIR/miniooni-domain-check.XXXXXX)
log $ipv4_doh_list
getipv4list $domain > $ipv4_doh_list
log $(cat $ipv4_doh_list)

log "* checking for DNS consistency"
ipv4_overlap_list=$(comm -12 $ipv4_system_list $ipv4_doh_list)
if [ "$ipv4_overlap_list" != "" ]; then
  output "MINIOONI_DNS_CONSISTENCY=1"
fi

log "* checking for IPs accessibility and validity for $domain"
valid=0
total=0
for ip in $(cat $ipv4_system_list); do
  log "* checking whether $ip is valid for $domain"
  urlgetter -OTLSServerName=$domain -i tlshandshake://$ip:443
  if [ "$(getfailure)" == "null" ]; then
    valid=$(($valid + 1))
  fi
  total=$(($total + 1))
done
if [ $total -gt 0 ]; then
  output "MINIOONI_DNS_REPLY_VALID=$(echo $valid/$total|bc -l)"
fi

log "* checking for HTTP body consistency"
urlgetter -ONoFollowRedirects=true -i http://$domain
body_vanilla=$(getbody)
log $body_vanilla
urlgetter -ONoFollowRedirects=true -OTunnel=psiphon -i http://$domain
body_tunnel=$(getbody)
log $body_tunnel
if [ "$body_vanilla" == "$body_tunnel" ]; then
  output "MINIOONI_HTTP_CONSISTENCY=1"
fi
