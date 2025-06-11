apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: host-gateway-admission-webhook
  labels:
    app: host-gateway-admission-webhook
  annotations:
    cert-manager.io/inject-ca-from: host-gateway-admission-webhook-certificate
webhooks:
  - name: host-gateway-admission-webhook.sj26.github.io
    admissionReviewVersions: ["v1"]
    clientConfig:
      caBundle: |
        ${CA_BUNDLE}
      service:
        name: host-gateway-admission-webhook
        namespace: default
        path: "/mutate"
        port: 443
    objectSelector:
      matchExpressions:
        - {
            key: app,
            operator: NotIn,
            values: [host-gateway-admission-webhook],
          }
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    sideEffects: None
