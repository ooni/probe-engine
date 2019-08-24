// Package ootemplate contains base code for OONI measurements.
//
// Traditionally OONI has always had a layer of abstraction between the
// code implementing the measurements and the measurement engine. The name
// originally provided to this layer was "test templates". This package
// is following the original convention, hence its name.
//
// Compared to the original code and specification, this code is more
// strict with respect to typing. There is currently a bunch of places
// in which nullable strings that oughta be `null` according to the
// spec are instead serialized as empty strings.
//
// # TCP connect template
//
// Strictly speaking this is not listed as a template in the OONI spec, yet
// it is widely used, e.g., by IM tests.
//
// This template attempts to connect directly to a specific TCP endpoint
// and produces the following JSON measurement:
//
//         {
//           "ip": "149.154.167.50",
//           "port": 443,
//           "status": {
//             "failure": "generic_timeout_error",
//             "success": false
//           }
//         }
//
// # HTTP template
//
// The HTTP template is the basic building block of Web Connectivity as
// well as of the IM tests. The implementation we have here is following
// the informal specification given in the Web Connectivity test.
//
// The emitted JSON measurement is:
//
//         {
//           "failure": "",
//           "request": {
//             "body": "",
//             "headers": {
//               "Accept": "text/html",
//             },
//             "method": "GET",
//             "tor": {
//               "exit_ip": null,
//               "exit_name": null,
//               "is_tor": false
//             },
//             "url": "http://torproject.org/"
//           },
//           "response": {
//             "body": "....",
//             "code": 200,
//             "headers": {
//               "Accept-Ranges": "bytes",
//             }
//            }
//         }
//
// This should be close enough to the expected format not to create issues
// to the pipeline parser. As mentioned, most string fields that may be
// nullable are actually emitted as empty rather then null, when not set.
package ootemplate
