# Package github.com/ooni/probe-engine/oonimkall

Package oonimkall implements API used by OONI mobile apps. We
expose this API to mobile apps using gomobile.

This package is named oonimkall because it's a ooni/probe-engine
mplementation of the mkall API implemented by Measurement Kit
in, e.g., [mkall-ios](https://github.com/measurement-kit/mkall-ios).

The basic tenet of oonimkall is that you define an experiment
task you wanna run using a JSON, then you start such task, and
you receive events as serialized JSONs. In addition to this
functionality, we also include extra APIs used by OONI mobile.
