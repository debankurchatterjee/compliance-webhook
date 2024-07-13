FROM tcx-docker-local.artifactory.eng.vmware.com/release/images/photon4-baseimage:3.1.0-123
WORKDIR /app
COPY certs/tls.crt certs/tls.key ./  bin/webhook ./
EXPOSE 8443
ENTRYPOINT ["./webhook"]
