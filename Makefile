.PHONY: default build install uninstall

default:
	@echo "Use 'make build' to build and push a new version of the controller image"
	@echo "Use 'make install' to install to your local k8s cluster"
	@echo "Use 'make uninstall' to uninstall from your local k8s cluster"
	@echo "Use 'make clean' to remove the self-signed cert and generated template files"

build:
	docker build --push --tag ghcr.io/sj26/host-gateway-admission-webhook .

install: tls.key tls.crt k8s/mutatingwebhook.yaml
	kubectl create secret tls host-gateway-admission-webhook --key tls.key --cert tls.crt
	kubectl create -f k8s

k8s/mutatingwebhook.yaml: k8s/mutatingwebhook.yaml.tpl tls.crt
	CA_BUNDLE="`cat tls.crt | base64 | tr -d '\n'`" envsubst < k8s/mutatingwebhook.yaml.tpl > k8s/mutatingwebhook.yaml

tls.crt: tls.key
	openssl req -x509 -key tls.key -sha256 -days 3650 -out tls.crt \
	    -subj "/CN=host-gateway-admission-webhook.default.svc" \
	    -addext "subjectAltName=DNS:host-gateway-admission-webhook.default.svc,DNS:host-gateway-admission-webhook.default,DNS:host-gateway-admission-webhook"
tls.key:
	openssl genrsa -out tls.key 4096

uninstall:
	kubectl delete -f k8s; 	kubectl delete secret host-gateway-admission-webhook

clean:
	rm -f tls.key tls.crt k8s/mutatingwebhook.yaml
