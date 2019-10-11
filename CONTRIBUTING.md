# Contributing to ooni/probe-engine

This is an open source project, and contributions are welcome! You are welcome
to open pull requests. An open pull request will be reviewed by a core
developer. The review may request you to apply changes. Once the assigned
reviewer is satisfied, they will merge the pull request.

## PR requirements

Every pull request that introduces new functionality should at least
include a basic integration test. Any pull request that modifies existing
functionality should pass existing tests. What's more, any new pull
request that modifies existing functionality should not decrease the
existing code coverage.

If there is a top-level DESIGN.md document, make sure such document is
kept in sync with code changes you have applied.

Do not submit large PRs. A reviewer can best service your PR if the
code changes are around 200-600 lines. (It is okay to add more changes
afterwards, if the reviewer asks you to do more work; the key point
here is that the PR should be reasonably sized when the review starts.)

In this vein, we'd rather structure a complex issue as a sequence of
small PRs, than have a single large PR address it all.

As a general rule, a PR is reviewed by reading the whole diff. Let us
know if you want us to read each diff individually, if you think that's
functional to better understand your changes.

## Code style requirements

You should use `go fmt`, `go vet`, and `golint` to vet your code
contribution before submitting a pull request. Make sure your code
is documented. At the minimum document all the exported symbols.

Make sure you commit `go.mod` and `go.sum` changes. Make sure you
run `go mod tidy` to minimize such changes.

## Code testing requirements

Make sure all tests pass with `go test -tags nopsiphon ./...` run
from the top-level directory. Omit `-tags nopsiphon` if you have
access to the secret required by the Psiphon experiment tests.

## Writing a new OONI experiment

When you are implementing a new experiment (aka nettest), make sure
you have read the relevant spec from the [ooni/spec](
https://github.com/ooni/spec) repository. If the spec is missing,
please help the pull request reviewer to create it. If the spec is
not clear, please let us know during the review.

When you write a new experiment, keep the measurement phase and the
results analysis phases as separate functions. This helps us a lot
to write better unit tests for our code.

To get a sense of what we expect from an experiment, see:

- the experiment/example experiment

- the experiment/dash experiment

Thank you!

