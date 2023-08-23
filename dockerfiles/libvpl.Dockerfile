FROM akashisn/ffmpeg:6.0-libvpl
RUN apt update && apt install -y ca-certificates
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]