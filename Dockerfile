FROM ghcr.io/vilsol/ffmpeg-alpine:latest@sha256:1bc053b7b5148efea76a6f4ace36247e80e470f24e70ba65e75cfe2d24842ff1 as ffmpeg

FROM alpine:edge@sha256:e64a0b2fc7ff870c2b22506009288daa5134da2b8c541440694b629fc22d792e

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

COPY transcoder-go /transcoder

ENTRYPOINT ["/transcoder"]