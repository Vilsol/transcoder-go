FROM jrottenberg/ffmpeg:5.1-vaapi
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]