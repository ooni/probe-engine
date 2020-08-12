""" ./QA/common.py - common code for QA """

import contextlib
import json
import os
import shlex
import shutil
import subprocess
import sys
import time
import urllib.parse


def execute(args):
    """ Execute a specified command """
    subprocess.run(args)


def execute_jafar_and_miniooni(ooni_exe, outfile, experiment, tag, args):
    """ Executes jafar and miniooni. Returns the test keys. """
    tmpoutfile = "/tmp/{}".format(outfile)
    with contextlib.suppress(FileNotFoundError):
        os.remove(tmpoutfile)  # just in case
    execute(
        [
            "./jafar",
            "-main-command",
            "{} -no '{}' --home ./QA/HOME {}".format(ooni_exe, tmpoutfile, experiment),
            "-main-user",
            "nobody",  # should be present on Unix
            "-tag",
            tag,
        ]
        + args
    )
    shutil.move(tmpoutfile, outfile)
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
