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
    assert isinstance(tk["requests"], list)
    assert len(tk["requests"]) > 0
    for entry in tk["requests"]:
        assert isinstance(entry, dict)
        failure = entry["failure"]
        assert isinstance(failure, str) or failure is None
        assert isinstance(entry["request"], dict)
        req = entry["request"]
        common.check_maybe_binary_value(req["body"])
        assert isinstance(req["headers"], dict)
        for key, value in req["headers"].items():
            assert isinstance(key, str)
            common.check_maybe_binary_value(value)
        assert isinstance(req["method"], str)
        assert isinstance(entry["response"], dict)
        resp = entry["response"]
        common.check_maybe_binary_value(resp["body"])
        assert isinstance(resp["code"], int)
        if resp["headers"] is not None:
            for key, value in resp["headers"].items():
                assert isinstance(key, str)
                common.check_maybe_binary_value(value)
    assert isinstance(tk["tcp_connect"], list)
    assert len(tk["tcp_connect"]) > 0
    for entry in tk["tcp_connect"]:
        assert isinstance(entry, dict)
        assert isinstance(entry["ip"], str)
        assert isinstance(entry["port"], int)
        assert isinstance(entry["status"], dict)
        failure = entry["status"]["failure"]
        success = entry["status"]["success"]
        assert isinstance(failure, str) or failure is None
        assert isinstance(success, bool)
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


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: %s /path/to/ooniprobelegacy-like/binary" % sys.argv[0])
    outfile = "webconnectivity.jsonl"
    ooni_exe = sys.argv[1]
    tests = [
        webconnectivity_transparent_http_proxy,
        webconnectivity_dns_hijacking,
        webconnectivity_control_unreachable_http,
    ]
    for test in tests:
        test(ooni_exe, outfile)
        time.sleep(7)


if __name__ == "__main__":
    main()
