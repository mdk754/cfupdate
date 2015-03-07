# cfupdate #

A fast and simple CloudFlare DNS updater utility written in Go.

## Overview ##

cfupdate will check your public IPv4 address against the last set value and update the CloudFlare API as necessary.

Currently the project will properly update the API, but I *do not* recommend production use. There is not enough error checking on the responses returned from the API.
