FROM jrottenberg/ffmpeg:5.1-vaapi
RUN apt update && apt install -y ca-certificates
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]