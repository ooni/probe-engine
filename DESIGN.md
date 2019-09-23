# github.com/ooni/probe-engine design

| Author       | Simone Basso
|--------------|---------------|
| Last-Updated | 2019-09-23    |
| Status       | Self approved |

Using [Measurement Kit](github.com/measurement-kit), aka MK, as the engine
for OONI mobile apps is causing us to do too much work. The upcoming release
of Desktop OONI apps is increasing such work. We will need to support,
in fact, several additional platforms.

The implementation of a OONI experiment for detecting the blocking of
Psiphon is creating an interesting opportunity. We are going to pay
the cost of depending from the Go core anyway, because Psiphon is written
in Go. Therefore, why not rewriting in Go significant portions of our
engine as well?

The [ooni/probe-engine](https://github.com/ooni/probe-engine) has been
created to host an initial PoC to explore this possibility. It is in turn
based on an earlier experiment, namely [measurement-kit/engine](
https://github.com/measurement-kit/engine). This document adopts the
point of view we had when we started working on it in May, 2019.

## Development plan

We will merge into ooni/probe-engine existing Go code living in the
github.com/measurement-kit namespace that is used for running MK
tests in Go. Then, we will refactor ooni/probe-cli code and put it
into ooni/probe-engine. The general goal will be that of having a
generic Go library that can be used by ooni/probe-cli as well as
by ooni/probe-ios and ooni/probe-android.

As regards ooni/probe-cli, we will base the first public release on
ooni/probe-engine. This first public release should use MK for
running all the experiments. However, all the other functionality
that ooni/probe-cli may need shall be implemented in Go. We will
gradually move more Go code from probe-cli to probe-engine, and
we will gradually reimplement existing tests in Go. Future releases
of probe-cli shall therefore start running tests written in Go.

We will gradually aim for very high code coverage. We will use
writing unit tests as a tool for structuring reasonable boundaries
between internal modules. We will also gradually aim for better
end-to-end testing, where core functionality like submitting
measurements will test against running instances of our backends
so we can verify that the functionality is correct.

As regards the mobile apps, we will use the [Go mobile](
https://godoc.org/golang.org/x/mobile/cmd/gomobile) tool to
automatically generate bindings for ObjectiveC and Java. We'll
create Go code that is close enough to the API currently
used by the mobile apps that its bindings could be drop-in
replacements for existing code. At that point we will modify
the apps to use the Go engine, rather than MK. Later on, we
will refactor the apps to use directly the engine API.

We will probably not ship releases of the mobile apps where
some code is written in Go and some code is written in C++
because that may make the apps very bloated. We may however
revisit this decision in case we can apply techniques like
[APK splits](
https://developer.android.com/studio/build/configure-apk-splits)
and [on-demand resources](
  https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/On_Demand_Resources_Guide/index.html
) successfully.

## Code architecture

This repository should provide Go interfaces allowing one
to tap into the following functionality:

* Discover OONI backend services

* Discover the probe's IP, ASN, country code

* Run any OONI experiment

* Submit OONI measurements

* Interact with OONI orchestra

* Receive real-time logs for all the above

Internal code should be internal. Delegating functionality
to other repositories is also okay, as long as the core
API described above is part of this repository. Specific
tests, for example, could be implemented in other repositories.

To migrate the mobile apps to use ooni/probe-engine rather
than MK, we need to write an MK compatibility layer, as mentioned
above. Such compatibility layer should be the starting point
to start refactoring. However, the long term goal is to get rid
of such compatibility layer and have the apps use directly
the Go-mobile bindings of probe-engine's main API. For this
reason, we may put this compat layer in another repo.