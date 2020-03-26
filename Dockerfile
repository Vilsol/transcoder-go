FROM golang:alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR $GOPATH/src/github.com/Vilsol/transcoder-go/

ENV GO111MODULE=on

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /transcoder main.go

FROM vilsol/ffmpeg-alpine as ffmpeg

FROM alpine:edge

RUN apk add --no-cache \
	libtheora \
	libvorbis \
	x264-libs \
	x265 \
	fdk-aac \
	lame \
	opus \
	libvpx \
	nasm

COPY --from=ffmpeg /root/bin/ffmpeg /bin/ffmpeg
COPY --from=ffmpeg /root/bin/ffprobe /bin/ffprobe

COPY --from=builder /transcoder /transcoder

ENTRYPOINT ["/transcoder"]