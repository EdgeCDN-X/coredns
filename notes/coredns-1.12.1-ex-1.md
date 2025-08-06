+++
title = "CoreDNS-1.12.1-ex-1.1 Release"
description = "CoreDNS-1.12.1-ex-1.1 Release Notes."
tags = ["Release", "1.12.1-ex-1.1", "Notes"]
release = "1.12.1-ex-1.1"
date = "2025-03-24T00:00:00+00:00"
author = "coredns"
+++

In this release:
* kubernetes: Revert recent change to only create PTR records for endpoints with hostname defined.
* forward: added option to return SERVFAIL immediately if all upstreams are unhealthy.

## Brought to You By

Adrian Moisey,
Arthur Outhenin-Chalandre,
Bartosz Borkowski,
Ben Kochie,
Chris O'Haver,
Min Woo Kim,
Puneet Loya,
Rich,
Viktor,
momantech


## Noteworthy Changes

* Added EdgeCDN-X Components