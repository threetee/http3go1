[![Build Status](https://travis-ci.org/threetee/http3go1.svg?branch=master)](https://travis-ci.org/threetee/http3go1)

# http3go1

http3go1 is a configurable HTTP redirector built in go. It currently uses redis (presumably with persistence) for all storage.

## Prerequisites

* go
* [redis](http://redis.io/)
* [forego](https://github.com/ddollar/forego) (optional, for development convenience)

## Getting Started

With forego:

    $ make
    $ forego start

Or, without forego:

    $ make
    $ ./admin &
    $ ./redirector &

## Configuration

Both the redirector and the admin component are intended to be configured via environment variables (see http://12factor.net/config). The following environment variables are available for use:

| Variable         | Description                                       | Type   | Default Value      |
|------------------|---------------------------------------------------|--------|--------------------|
| REDIRECTOR_DEBUG | Redirector: Enables debugging output              | string | false              |
| REDIRECTOR_HOST  | Redirector: Host to use for the TCP listener      | string | 0.0.0.0            |
| REDIRECTOR_PORT  | Redirector: TCP port to listen on                 | string | 9000               |
| ADMIN_DEBUG      | Admin interface: Enables debugging output         | string | false              |
| ADMIN_HOST       | Admin interface: Host to use for the TCP listener | string | 0.0.0.0            |
| ADMIN_PORT       | Admin interface: TCP port to listen on            | string | 9001               |
| REDIS_HOST       | Redis host connection string                      | string | tcp:localhost:6379 |
| REDIS_DB         | Redis DB to use                                   | string | 0                  |
| REDIS_PREFIX     | Prefix for redis keys                             | string | h3g1               |
| REDIS_PASS       | Optional redis password                           | string |                    |

## Details

The redirector will respond to all requests with a redirect, the target of which is determined by user configuration. For example, a request to www.domain.com/path can be configured to redirect to host.anotherdomain.net/newpath.

## Data

Redirects are stored in redis hashes, with fields:
* TargetUrl
* CreationDate
* Clicks

Hashes are stored using the source URL as the key.

## Containerization

You can build docker containers for both components by running:

  $ make docker-dist

## Acknowledgements

I based this tool on kurz.go (https://github.com/fs111/kurz.go), although much has been modified. Thanks to the kurz.go author for a good reference point.
