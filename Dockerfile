FROM alpine:3 as certs
RUN apk --update add ca-certificates

#checkov:skip=CKV_DOCKER_2: healthchecks are set in K8s
FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1001:1001
ENTRYPOINT ["/cost-exporter"]
COPY cost-exporter /
