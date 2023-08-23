FROM akashisn/ffmpeg:6.0-libvpl
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]