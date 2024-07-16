FROM photon:4.0
WORKDIR /app
COPY  bin/webhook ./
EXPOSE 8443
ENTRYPOINT ["./webhook"]
