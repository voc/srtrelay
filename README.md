# srtrelay
Streaming-Relay for the SRT-protocol

**EXPERIMENTAL AT BEST, use at your own risk**

Credit goes to github.com/asticode/go-astits

## Design Ideas
  - Just an 1:n relay, one publisher, multiple subscribers
  - No decoding -> use ffmpeg
  - No remuxing -> use ffmpeg
  - Allow any data, not just MPEG-TS
