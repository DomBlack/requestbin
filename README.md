# Requestbin

# Overview

This is a debugging tool to view HTTP requests made by a client.

# Use cases

* Investigate file upload services. You want to know what client is being used to fetch your files. Using some file types that can have embedded links, you can also detect what is used to process your file once downloaded.
* Data exfiltration

# How to use

## Basic usage

Assuming your server is at http://example.com

Any requests to http://example.com/_ will be logged and stored with the following info
* Url path
* Full url
* HTTP Method
* Headers
* Remote address
* Raw Body

The first folder in the path (the bin id) is used to group requests together.
All requests to http://example.com/_/abc will be visible at http://example.com/abc.

Bins don't need to pre-exist, just use a new one if you want.
Requests by default expire after 24h.

# File types

If your client is expecting a specific file type for your testing, the tool will return dummy data for the file types currently supported based on the file extension in the path.
Only the file extension is used, you can choose any valid filename you want, for example to make it look more legit or to tag specific queries.

## Static file types

The following file types are currently supported

* `.png`
* `.bmp`
* `.gif`
* `.jpg`
* `.mp3`
* `.css`

Just use `http://example.com/_/abc/foo.jpg` to get a valid jpg file.
The files are under `static/files`.

## Dynamic file types

Some file types let you embed links that get automatically fetched when the file is opened.
A url related to the dynamic file type is generated and used as the ping back url, with the appropriate file format.

The following file types are currently supported

* `.odt`: Links to a `.jpg` file
* `.torrent`: The pingback url is returned in the `announce`, `announce-list` and `http-peers` fields.
* `.svg`: Includes a `.css` stylesheet
* `.m3u`: Links to a `.mp3` file

**Example**

```
$ curl http://example.com/_/foo/playlist.m3u
#EXTM3U
http://example.com/_/foo/playlist.mp3
```

# How to setup

The tool is meant to be run inside docker containers (one for the HTTP endpoint, one for redis).
You can either build your own image or pull from the docker hub, see below.

## With Docker

## Use pre-built image

Just run

```
docker-compose up
```

You only need the `docker-compose.yml` file, no need to checkout the full repo.

### Build your own image

```
make build
make run-docker
```

You can also use

```
make run-local
```

To run it without docker.
You will need to update the `Makefile` to make it point to your redis server if not running locally.

You need to install the dependencies manually

```
go get "github.com/garyburd/redigo/redis"
go get "github.com/gorilla/mux"
go get "github.com/jackpal/bencode-go"
go get github.com/codegangsta/gin
```

# References

* [Dropbox Opening my docs? (archived)](https://archive.is/GCWoO)
* BSides Toronto 2014 talk "Honeydocs and Offensive Countermeasures" by Roy Firestein
    * [DocPing](https://docping.me/)
	* [Video](https://www.youtube.com/watch?v=a-b6uyDL1Rg)
	* [Slides](http://sector.ca/Portals/17/Presentations14/Roy%20Firestein%20-%20SecTor%202014.pdf)
* Defcon 23 talk "Bugged Files, Is your Document Telling on You?" by Daniel Crowley and Damon Smith
    * [Video](https://www.youtube.com/watch?v=M3vP-wughPo)
	* [Slides](https://media.defcon.org/DEF%20CON%2023/DEF%20CON%2023%20presentations/DEFCON-23-Daniel-Crowley-Damon-Smith-Bugged-Files.pdf)
