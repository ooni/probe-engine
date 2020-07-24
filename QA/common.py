""" ./QA/common.py - common code for QA """

import contextlib
import json
import os
import shlex
import subprocess
import sys
import time
import urllib.parse


def execute(args):
    """ Execute a specified command """
    subprocess.run(args)


def execute_jafar_and_miniooni(ooni_exe, outfile, experiment, tag, args):
    """ Executes jafar and miniooni. Returns the test keys. """
    with contextlib.suppress(FileNotFoundError):
        os.remove(outfile)  # just in case
    execute(
        [
            "./jafar",
            "-main-command",
            "%s -no '%s' %s" % (ooni_exe, outfile, experiment),
            "-main-user",
            "ooniprobe",  # created in cmd/jafar/Dockerfile
            "-tag", tag,
        ]
        + args
    )
    result = read_result(outfile)
    assert isinstance(result, dict)
    assert isinstance(result["test_keys"], dict)
    return result["test_keys"]


def read_result(outfile):
    """ Reads the result of an experiment """
    return json.load(open(outfile, "rb"))


def test_keys(result):
    """ Returns just the test keys of a specific result """
    return result["test_keys"]


def check_maybe_binary_value(value):
    """ Make sure a maybe binary value is correct """
    assert isinstance(value, str) or (
        isinstance(value, dict)
        and value["format"] == "base64"
        and isinstance(value["data"], str)
    )
