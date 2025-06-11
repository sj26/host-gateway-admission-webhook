# `host-gateway` Admission Webhook for Kubernetes

There is a [closed feature request for `host-gateway` support in Kubernetes][feature-request]. I really want this to exist. I use [Docker's `host-gateway` support][host-gateway] all the time to connect host services to container services and back again in my development environment. So I made it work.

This is a Kubernetes [mutating admission webhook][mutating-admission-webhook] which mutates [Pod resources][pod] when a [HostAlias][host-alias] specifies an IP of `host-gateway`, replacing it with the resolved IP for `host.docker.internal`.

  [feature-request]: https://github.com/kubernetes/kubernetes/issues/107079
  [host-gateway]: https://docs.docker.com/reference/cli/docker/container/run/#add-host
  [mutating-admission-webhook]: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook
  [pod]: https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/
  [host-alias]: https://kubernetes.io/docs/concepts/services-networking/add-entries-to-pod-etc-hosts-with-host-aliases/

> [!WARNING]
> **This is an experiment.** It's good enough for use in a development environment. But do not ship this to production.

## Usage

Create the admission webhook:

```
kubectl apply -f k8s
```

Then use it from a pod like so:

```
cat <<EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: shell-demo
spec:
  containers:
  - name: bash
    image: bash
    command: [ sleep, infinity]
  hostAliases:
  - ip: host-gateway
    hostnames: [ example.com ]
EOF
```

Confirm it worked:

```
kubectl exec -it pod/shell-demo -- ping -c 1 example.com
PING example.com (0.250.250.254): 56 data bytes
64 bytes from 0.250.250.254: seq=0 ttl=62 time=0.250 ms

--- example.com ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 0.250/0.250/0.250 ms
```

You can also monitor the logs to confirm it's working:

```
kubectl logs -f deployment/host-gateway-admission-webhook
2025/06/11 02:35:49 [default] Pod/shell-demo: replacing `host-gateway` with `0.250.250.254`
```

If it fails, `kubectl create ...` should fail with a useful error message.

## Caveats

- TLS is required for admissions webhooks. An example key and cert is supplied for ease of use. This is a terrible idea. But it's fine for experimental purposes. You can regenerate these with `rm tls.* && make tls.crt`, then change the resources within the `k8s` directory. (This could be automated.)
- `host-gateway` is resolved by the webhook service, so will use the IP it sees for `host.docker.internal`. This is fine for [Kubernetes running in Docker Desktop][docker-desktop-kubernetes], or alternatives like [OrbStack][orbstack-kubernetes]. But it will break on multi-node clusters, or when `host.docker.internal` is not available.

  [docker-desktop-kubernetes]: https://docs.docker.com/desktop/kubernetes/
  [orbstack-kubernetes]: https://docs.orbstack.dev/kubernetes/
