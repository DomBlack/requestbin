# Requestbin

# Overview

This is a debugging tool to view HTTP requests made by a client.

![Overview](docs/screenshot.png?raw=true "Overview")

# Use cases

* Investigate file upload services. You want to know what client is being used to fetch your files. Using some file types that can have embedded links, you can also detect what is used to process your file once downloaded.
* Data exfiltration

# How to use

## Basic usage

Assuming your server is at http://example.com:18080

Any requests to http://example.com:18080/ will be logged and stored with the following info
* Url path
* Full url
* HTTP Method
* Headers
* Remote address
* Raw Body

The first folder in the path (the bin id) is used to group requests together.
All requests to http://example.com:18080/abc will be visible at http://example.com:18081/abc.

Bins don't need to pre-exist, just use a new one if you want.
Requests by default expire after 24h.

## File types

If your client is expecting a specific file type for your testing, the tool will return dummy data for the file types currently supported based on the file extension in the path.
Only the file extension is used, you can choose any valid filename you want, for example to make it look more legit or to tag specific queries.

### Static file types

The following file types are currently supported

* `.png`
* `.bmp`
* `.gif`
* `.jpg`
* `.mp3`
* `.css`

Just use `http://example.com/_/abc/foo.jpg` to get a valid jpg file.
The files are under `static/files`.

### Dynamic file types

Some file types let you embed links that get automatically fetched when the file is opened.
A url related to the dynamic file type is generated and used as the ping back url, with the appropriate file format.

The following file types are currently supported

* `.odt`: Links to a `.jpg` file
* `.torrent`: The pingback url is returned in the `announce`, `announce-list` and `http-peers` fields.
* `.svg`: Includes a `.css` stylesheet
* `.m3u`, `.asx`, `.pls`, `.xspf`: Links to a `.mp3` file

**Example**

```
$ curl http://example.com:18080/foo/playlist.m3u
#EXTM3U
http://example.com:18080/foo/playlist.mp3
```

## API

You can use the following endpoints to retrieve info in `json`

* `http://example.com:8081/api/bins` To list all the bins
* `http://example.com:8081/api/bins/<binId>` To list the requests for that bin

```
$ http http://example.com:18081/api/bins
HTTP/1.1 200 OK
Content-Length: 96
Content-Type: application/json; charset=UTF-8
Date: Sun, 27 Mar 2016 01:48:51 GMT

[
    "docs",
    "foo"
]
```

```
$ http http://example.com:18081/api/bins/foo
HTTP/1.1 200 OK
Content-Length: 654
Content-Type: application/json; charset=UTF-8
Date: Sun, 27 Mar 2016 01:49:03 GMT

[
    {
        "body": "",
        "form": {
            "k": [
                "1"
            ]
        },
        "full_url": "http://example.com/_/foo/playlist.m3u?k=1",
        "headers": {
            "Accept": [
                "*/*"
            ],
            "User-Agent": [
                "curl/7.43.0"
            ]
        },
        "host": "example.com",
        "json": null,
        "method": "GET",
        "post_form": {},
        "remote_addr": "10.10.123.62:46372",
        "time": "2016-03-27T01:17:27.78127107Z",
        "url": "/_/foo/playlist.m3u"
    },
    {
        "body": "",
        "form": {},
        "full_url": "http://example.com/_/foo/playlist.mp3",
        "headers": {
            "Accept": [
                "*/*"
            ],
            "User-Agent": [
                "curl/7.43.0"
            ]
        },
        "host": "example.com",
        "json": null,
        "method": "GET",
        "post_form": {},
        "remote_addr": "10.10.123.62:46354",
        "time": "2016-03-27T01:16:32.091116201Z",
        "url": "/_/foo/playlist.m3u"
    }
]
```

# How to setup

The tool is meant to be run inside docker containers

* redis
* elasticsearch
* kibana
* go server (18080 logging, 18081 api/ui, 18082 proxy to kibana with basic auth)
You can either build your own image or pull from the docker hub, see below.

## With Docker

## Use pre-built image

Create a `passwd` file for the passwords to access Kibana.
Format is
`username:md5(passwd)` (NOT htpasswd, too much hassle to parse).

Then run

```
docker-compose build
docker-compose up
```
