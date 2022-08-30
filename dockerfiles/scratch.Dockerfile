FROM jrottenberg/ffmpeg:5.1-ubuntu
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]