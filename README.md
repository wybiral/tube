# tube [![GoDoc](https://godoc.org/github.com/wybiral/tube?status.svg)](https://godoc.org/github.com/wybiral/tube)

This is a Golang project to build a self hosted "tube"-style video player for watching your own video collection over HTTP or hosting your own channel for others to watch.

Some of the key features include:
- Easy to add videos (just move a file into the folder)
- No database (video info pulled from file metadata)
- No JavaScript (the player UI is entirely HTML)
- Easy to customize CSS and HTML template
- Automatically generates RSS feed (at `/feed.xml`)
- Builtin Tor onion service support
- Clean, simple, familiar UI

Currently only supports MP4 video files so you may need to re-encode your media to MP4 using something like [ffmpeg](https://ffmpeg.org/).

Since all of the video info comes from metadata it's also useful to have a metadata editor such as [EasyTAG](https://github.com/GNOME/easytag) (which supports attaching images as thumbnails too).

By default the server is configured to run on 127.0.0.1:0 which will assign a random port every time you run it. This is to avoid conflicting with other applications and to ensure privacy. You can configure this to be any specific host:port by editing `config.json` before running the server. You can also change the RSS feed details and library path from `config.json`.

# installation

## from release

1. Download [release](https://github.com/wybiral/tube/releases) for your platform
2. Extract zip archive
3. Run `tube` executable to start server (this will output the URL for accessing from a browser)
4. Move videos to `videos` directory
5. Open the URL from step 3 and enjoy!

## from source

1. [Install Golang](https://golang.org/doc/install) if you don't already have it
2. `go get github.com/wybiral/tube`
3. `cd $GOPATH/src/github.com/wybiral/tube`
4. `go run main.go` (this will output the URL for accessing from a browser)
5. Move videos to `$GOPATH/src/github.com/wybiral/tube/videos`
6. Open the URL from step 4 and enjoy!

## from Container
```sh
go get github.com/wybiral/tube
cd $GOPATH/src/github.com/wybiral/tube
docker build . -t  wybiral/tube:v0.x
docker container run  -p 8080:8080   -v  /path/to/your/videos:$GOPATH/src/github.com/wybiral/tube/videos  wybiral/tube:v0.x

```