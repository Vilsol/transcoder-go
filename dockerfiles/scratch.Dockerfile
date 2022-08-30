FROM jrottenberg/ffmpeg:5.1-scratch
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY transcoder-go /transcoder
ENTRYPOINT ["/transcoder"]