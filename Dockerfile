FROM scratch

COPY pgm-linux-64 /pgm

ENTRYPOINT ["/pgm"]
