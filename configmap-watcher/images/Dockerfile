FROM ${gcr_endpoint}/distroless/static

WORKDIR /

COPY dist/remediator_linux_amd64/remediator .

ENTRYPOINT ["/remediator"]
