# raspi-live-libcamera
📷 raspi-live-libcamera is a command-line application that streams video from the Raspberry Pi Camera module using the new libcamera software to the web. This was forked and updated from an older version by https://github.com/jaredpetersen in which the older version used the `raspivid` application, which is now deprecated.

## Usage
```
raspi-live-libcamera streams video from the Raspberry Pi Camera Module to the web

For more information visit https://github.com/amd940/raspi-live-libcamera

Usage:
  raspi-live-libcamera [command]

Available Commands:
  hls         Stream video using HLS
  dash        Stream video using DASH
  help        Help about any command

Flags:
      --debug             enable debug logging
      --fps int           video framerate (default 30)
      --height int        video height (default 720)
  -h, --help              help for raspilive
      --horizontal-flip   horizontally flip video
  -v, --version           version for raspilive
      --vertical-flip     vertically flip video
      --width int         video width (default 1280)

Use "raspilive [command] --help" for more information about a command.
```

### Commands
#### HLS
The `hls` command muxes the video stream into the [HLS](https://en.wikipedia.org/wiki/HTTP_Live_Streaming) video
streaming format and serves the produced content by starting a static file server.

If you're not familiar with HLS, the technology works by splitting the video stream into small, consumable segments.
These segments are arranged into a constantly updating playlist of files. Clients periodically read these playlists,
download the listed videos, and queue up the segments to produce a seamless playback experience.
[Twitch uses it](https://blog.twitch.tv/en/2015/12/18/twitch-engineering-an-introduction-and-overview-a23917b71a25/)
to distribute streaming video to all of its viewers.

```
Stream video using HLS

Usage:
  raspi-live-libcamera hls [flags]

Flags:
      --port int              static file server port
      --directory string      static file server directory
      --tls-cert string       static file server TLS certificate
      --tls-key string        static file server TLS key
      --segment-type string   format of the video segments (valid ["mpegts", "fmp4"], default "mpegts")
      --segment-time int      target segment duration in seconds (default 2)
      --playlist-size int     maximum number of playlist entries (default 10)
      --storage-size int      maximum number of unreferenced segments to keep on disk before removal (default 1)
      --cors string           allow cross domain HTTP requests, https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS (default "", supports * wildcard (e.g. `--cors "*"`))
  -h, --help                  help for hls

Global Flags:
      --debug             enable debug logging
      --fps int           video framerate (default 30)
      --height int        video height (default 720)
      --horizontal-flip   horizontally flip video
      --vertical-flip     vertically flip video
      --width int         video width (default 1280)
```

#### DASH
The `dash` command muxes the video stream into the
[DASH](https://en.wikipedia.org/wiki/Dynamic_Adaptive_Streaming_over_HTTP) video streaming format and serves the
produced content by starting a static file server.

DASH effectively utilizes the same mechanism for streaming video as HLS. The video is split into small segments and
listed in a changing playlist file. Clients download the playlist and the videos listed in it to piece the video
together seamlessly.

```
Stream video using DASH

Usage:
  raspi-live-libcamera dash [flags]

Flags:
      --port int            static file server port
      --directory string    static file server directory
      --tls-cert string     static file server TLS certificate
      --tls-key string      static file server TLS key
      --segment-time int    target segment duration in seconds (default 2)
      --playlist-size int   maximum number of playlist entries (default 10)
      --storage-size int    maximum number of unreferenced segments to keep on disk before removal (default 1)
      --cors string         allow cross domain HTTP requests, https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS (default "", supports * wildcard (e.g. `--cors "*"`))
  -h, --help                help for dash

Global Flags:
      --debug             enable debug logging
      --fps int           video framerate (default 30)
      --height int        video height (default 720)
      --horizontal-flip   horizontally flip video
      --vertical-flip     vertically flip video
      --width int         video width (default 1280)
```

### Performance Tips
#### HLS & DASH
HLS and DASH are inherently latent streaming technologies. However, you can still produce some lower latency video
streams.

The general recommendations seem to be:
- Reduce the segment size
- Increase the number of segments in the playlist to build up a buffer

Experiment with the flags and see what seems to work best for your Pi. We try to provide "sane" defaults but Raspberry
Pis are computationally diverse so you may find better performance with some tweaking.

Additionally, you may find that the SD card on the Raspberry Pi is a limitation. Fast disk read/writes are important
and SD cards can only perform so many in their lifetime. For better performance and longevity, you may consider setting
up a [RAM drive](https://en.wikipedia.org/wiki/RAM_drive) so that the files are stored in memory instead.

## Installation
raspi-live-libcamera uses [libcamera](https://www.raspberrypi.com/documentation/computers/camera_software.html#rpicam-vid) to operate the
Raspberry Pi Camera Module (Note: the documentation needs to be updated, `rpicam-*` should instead be `libcamera-*`). This is already available on the Raspbian operating system and can be enabled via
[raspi-config](https://www.raspberrypi.org/documentation/configuration/raspi-config.md).

raspi-live-libcamera also uses [Ffmpeg](https://ffmpeg.org/), a prominent video conversion command line utility, to process the
streaming video that the Raspberry Pi Camera Module outputs. Version 4.0 or higher is required.
```zsh
sudo apt-get install ffmpeg
```

Download the latest version of raspi-live-libcamera from the [Releases page](https://github.com/amd940/raspi-live-libcamera/releases). All of the release binaries are compiled for ARM 6 and are compatible with Raspberry Pi.
