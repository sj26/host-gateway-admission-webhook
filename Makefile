default:
	docker build --push --tag ghcr.io/sj26/host-gateway-admission-webhook .

tls.crt: tls.key
	openssl req -x509 -key tls.key -sha256 -days 3650 -out tls.crt \
	    -subj "/CN=host-gateway-admission-webhook.default.svc" \
	    -addext "subjectAltName=DNS:host-gateway-admission-webhook.default.svc,DNS:host-gateway-admission-webhook.default,DNS:host-gateway-admission-webhook"
tls.key:
	openssl genrsa -out tls.key 4096
