FROM jrottenberg/ffmpeg:5.1-nvidia
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]