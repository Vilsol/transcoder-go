FROM jrottenberg/ffmpeg:5.1-scratch
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]