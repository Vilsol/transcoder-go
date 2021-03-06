FROM golang:1.15-alpine AS builder

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

# ffmpeg
COPY --from=ffmpeg /root/bin/ffmpeg /bin/ffmpeg
COPY --from=ffmpeg /root/bin/ffprobe /bin/ffprobe

# x265
COPY --from=ffmpeg /usr/local/ /usr/local/

RUN apk add --no-cache \
	libtheora \
	libvorbis \
	x264-libs \
	fdk-aac \
	lame \
	opus \
	libvpx \
	libstdc++ \
	numactl \
	nasm

COPY --from=builder /transcoder /transcoder

ENTRYPOINT ["/transcoder"]