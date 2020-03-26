# transcoder-go

```
transcoder is an opinionated wrapper around ffmpeg

Usage:
  transcoder [flags] <path> ...

Flags:
      --colors               Force output with colors
      --early-exit           Early exit if transcoded version is larger than original (requires keep-old) (default true)
  -e, --extensions strings   Transcoded file extensions (default [.mp4,.mkv,.flv])
  -f, --flags string         The base flags used for all transcodes (default "-map 0 -c:v libx265 -preset ultrafast -x265-params crf=16 -c:a aac -strict -2 -b:a 256k")
  -h, --help                 help for transcoder
      --interval int         How often to output transcoding status (default 5)
      --keep-old             Keep old version of video if transcoded version is larger (default true)
      --log string           The log level to output (default "info")
      --stderr               Whether to output ffmpeg stderr stream
      --tg-bot-key string    Telegram Bot API Key
      --tg-chat-id int       Telegram Bot Chat ID
```