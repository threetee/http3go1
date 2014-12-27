# http3go1

## Description

http3go1 is a configurable HTTP redirector built in go. It currently uses redis for all storage.

## Details

The redirector will respond to all requests with a redirect, the target of which is determined by user configuration. For example, a request to www.domain.com/path can be configured to redirect to host.anotherdomain.net/newpath.

## Data

Redirects are stored in redis hashes, with fields:
* TargetUrl
* CreationDate
* Clicks

Hashes will be stored using the source URL as the key.
