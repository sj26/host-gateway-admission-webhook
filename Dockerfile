FROM golang:alpine AS builder
ADD . /go/src/github.com/sj26/host-gateway-admission-webhook
WORKDIR /go/src/github.com/sj26/host-gateway-admission-webhook
RUN go build -o host-gateway-admission-webhook .

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/sj26/host-gateway-admission-webhook"
LABEL org.opencontainers.image.licenses="MIT"
COPY --from=builder /go/src/github.com/sj26/host-gateway-admission-webhook/host-gateway-admission-webhook  /
ENTRYPOINT ["/host-gateway-admission-webhook"]
