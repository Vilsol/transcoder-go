FROM ghcr.io/vilsol/ffmpeg-alpine:latest as ffmpeg

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

COPY transcoder-go /transcoder

ENTRYPOINT ["/transcoder"]