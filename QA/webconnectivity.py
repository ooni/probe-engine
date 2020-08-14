#!/usr/bin/env python3


""" ./QA/webconnectivity.py - main QA script for webconnectivity

    This script performs a bunch of webconnectivity tests under censored
    network conditions and verifies that the measurement is consistent
    with the expectations, by parsing the resulting JSONL. """

import contextlib
import json
import os
import shlex
import socket
import subprocess
import sys
import time
import urllib.parse

sys.path.insert(0, ".")
import common


def execute_jafar_and_return_validated_test_keys(
    ooni_exe, outfile, experiment_args, tag, args
):
    """ Executes jafar and returns the validated parsed test keys, or throws
        an AssertionError if the result is not valid. """
    tk = common.execute_jafar_and_miniooni(
        ooni_exe, outfile, experiment_args, tag, args
    )
    return tk


def webconnectivity_transparent_http_proxy(ooni_exe, outfile):
    """ Test case where we pass through a transparent HTTP proxy """
    args = []
    args.append("-iptables-hijack-https-to")
    args.append("127.0.0.1:443")
    tk = execute_jafar_and_return_validated_test_keys(
        ooni_exe,
        outfile,
        "-i https://example.org web_connectivity",
        "webconnectivity_transparent_http_proxy",
        args,
    )
    assert tk["dns_experiment_failure"] == None
    assert tk["dns_consistency"] == "consistent"
    assert tk["control_failure"] == None
    assert tk["http_experiment_failure"] == None
    assert tk["body_length_match"] == True
    assert tk["body_proportion"] == 1
    assert tk["status_code_match"] == True
    assert tk["headers_match"] == True
    assert tk["title_match"] == True
    assert tk["blocking"] == False
    assert tk["accessible"] == True


def webconnectivity_dns_hijacking(ooni_exe, outfile):
    """ Test case where there is DNS hijacking towards a transparent proxy. """
    args = []
    args.append("-iptables-hijack-dns-to")
    args.append("127.0.0.1:53")
    args.append("-dns-proxy-hijack")
    args.append("example.org")
    tk = execute_jafar_and_return_validated_test_keys(
        ooni_exe,
        outfile,
        "-i https://example.org web_connectivity",
        "webconnectivity_dns_hijacking",
        args,
    )
    assert tk["dns_experiment_failure"] == None
    assert tk["dns_consistency"] == "inconsistent"
    assert tk["control_failure"] == None
    assert tk["http_experiment_failure"] == None
    assert tk["body_length_match"] == True
    assert tk["body_proportion"] == 1
    assert tk["status_code_match"] == True
    assert tk["headers_match"] == True
    assert tk["title_match"] == True
    assert tk["blocking"] == False
    assert tk["accessible"] == True


def webconnectivity_control_unreachable_http(ooni_exe, outfile):
    """ Test case where the control is unreachable and we're using the
        plaintext HTTP protocol rather than HTTPS """
    args = []
    args.append("-iptables-reset-keyword")
    args.append("wcth.ooni.io")
    tk = execute_jafar_and_return_validated_test_keys(
        ooni_exe,
        outfile,
        "-i http://example.org web_connectivity",
        "webconnectivity_control_unreachable_http",
        args,
    )
    assert tk["dns_experiment_failure"] == None
    assert tk["dns_consistency"] == None
    assert tk["control_failure"] == "connection_reset"
    assert tk["http_experiment_failure"] == None
    assert tk["body_length_match"] == None
    assert tk["body_proportion"] == 0
    assert tk["status_code_match"] == None
    assert tk["headers_match"] == None
    assert tk["title_match"] == None
    assert tk["blocking"] == None
    assert tk["accessible"] == None


def webconnectivity_nonexistent_domain(ooni_exe, outfile):
    """ Test case where the domain does not exist """
    if "measurement_kit" in ooni_exe:
        return  # MK result does not look correct
    args = []
    tk = execute_jafar_and_return_validated_test_keys(
        ooni_exe,
        outfile,
        "-i http://antani.xyz web_connectivity",
        "webconnectivity_nonexistent_domain",
        args,
    )
    assert tk["dns_experiment_failure"] == "dns_nxdomain_error"
    assert tk["dns_consistency"] == "consistent"
    assert tk["control_failure"] == None
    assert tk["http_experiment_failure"] == None
    assert tk["body_length_match"] == None
    assert tk["body_proportion"] == 0
    assert tk["status_code_match"] == None
    assert tk["headers_match"] == None
    assert tk["title_match"] == None
    assert tk["blocking"] == None
    assert tk["accessible"] == None


def webconnectivity_tcpip_blocking_with_consistent_dns(ooni_exe, outfile):
    """ Test case where there's TCP/IP blocking w/ consistent DNS """
    ip = socket.gethostbyname("nexa.polito.it")
    args = [
        "-iptables-drop-ip",
        ip,
    ]
    tk = execute_jafar_and_return_validated_test_keys(
        ooni_exe,
        outfile,
        "-i http://nexa.polito.it web_connectivity",
        "webconnectivity_tcpip_blocking_with_consistent_dns",
        args,
    )
    assert tk["dns_experiment_failure"] == None
    assert tk["dns_consistency"] == "consistent"
    assert tk["control_failure"] == None
    assert tk["http_experiment_failure"] == "generic_timeout_error"
    assert tk["body_length_match"] == None
    assert tk["body_proportion"] == 0
    assert tk["status_code_match"] == None
    assert tk["headers_match"] == None
    assert tk["title_match"] == None
    assert tk["blocking"] == "tcp_ip"
    assert tk["accessible"] == False


def webconnectivity_tcpip_blocking_with_inconsistent_dns(ooni_exe, outfile):
    """ Test case where there's TCP/IP blocking w/ inconsistent DNS """

    def runner(port):
        args = [
            "-dns-proxy-hijack",
            "nexa.polito.it",
            "-iptables-hijack-dns-to",
            "127.0.0.1:53",
            "-iptables-hijack-http-to",
            "127.0.0.1:{}".format(port),
        ]
        tk = execute_jafar_and_return_validated_test_keys(
            ooni_exe,
            outfile,
            "-i http://nexa.polito.it web_connectivity",
            "webconnectivity_tcpip_blocking_with_inconsistent_dns",
            args,
        )
        assert tk["dns_experiment_failure"] == None
        assert tk["dns_consistency"] == "inconsistent"
        assert tk["control_failure"] == None
        assert tk["http_experiment_failure"] == "connection_refused"
        assert tk["body_length_match"] == None
        assert tk["body_proportion"] == 0
        assert tk["status_code_match"] == None
        assert tk["headers_match"] == None
        assert tk["title_match"] == None
        assert tk["blocking"] == "dns"
        assert tk["accessible"] == False

    common.with_free_port(runner)


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: %s /path/to/ooniprobelegacy-like/binary" % sys.argv[0])
    outfile = "webconnectivity.jsonl"
    ooni_exe = sys.argv[1]
    tests = [
        webconnectivity_transparent_http_proxy,
        webconnectivity_dns_hijacking,
        webconnectivity_control_unreachable_http,
        webconnectivity_nonexistent_domain,
        webconnectivity_tcpip_blocking_with_consistent_dns,
        webconnectivity_tcpip_blocking_with_inconsistent_dns,
    ]
    for test in tests:
        test(ooni_exe, outfile)
        time.sleep(7)


if __name__ == "__main__":
    main()
