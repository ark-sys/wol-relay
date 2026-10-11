# wol-relay

A small HTTP service that sends Wake-on-LAN magic packets. A pod in Kubernetes usually can't send a Wake-on-LAN
packet itself: the packet is a UDP broadcast on the physical LAN, and pod networks don't forward broadcasts.
`wol-relay` runs on a node's host network and turns an HTTP call into that broadcast.

I use it in my homelab to start a GPU workstation on demand. The local LLMs behind my AIOps tooling run on that
machine, and it stays powered off until a service in the cluster calls `/wake` before sending inference requests.

- About 150 lines of Go, standard library only
- Distroless, non-root image for `amd64` and `arm64` (`ghcr.io/ark-sys/wol-relay`)
- Helm chart in `charts/wol-relay`

## API

`POST /wake`

```json
{ "mac": "AA:BB:CC:DD:EE:FF", "broadcast_address": "192.168.1.255", "port": 9 }
```

Only `mac` is required (`:` or `-` separators). `broadcast_address` and `port` override the defaults below.
The response is `200 {"status":"ok", ...}` once the packet is sent, `400` for a malformed body or MAC address, and
`500` if the packet can't be sent.

`GET /healthz` returns `200 {"status":"ok"}`.

There is no authentication. Keep the service internal to the cluster (the chart uses a `ClusterIP` service).

## Configuration

| Variable            | Default            | Description                               |
|---------------------|--------------------|-------------------------------------------|
| `LISTEN_ADDR`       | `:8089`            | HTTP listen address                       |
| `DEFAULT_BROADCAST` | `255.255.255.255`  | Broadcast address used when none is given |
| `DEFAULT_PORT`      | `9`                | UDP port used when none is given          |

Set `DEFAULT_BROADCAST` to your LAN's directed broadcast address (for example `192.168.1.255`) when the node has
several interfaces.

## Run

With Docker, the container needs the host network to reach the LAN:

```shell
docker run --rm --network host ghcr.io/ark-sys/wol-relay:latest

curl -X POST http://localhost:8089/wake \
  -H "Content-Type: application/json" \
  -d '{"mac":"AA:BB:CC:DD:EE:FF"}'
```

On Kubernetes, with Helm:

```shell
helm upgrade --install wol-relay ./charts/wol-relay \
  --namespace wol-relay --create-namespace \
  --set env.DEFAULT_BROADCAST="192.168.1.255"
```

The chart enables `hostNetwork` by default and runs the container as non-root with all capabilities dropped. Sending
to a broadcast address doesn't need `NET_RAW` or `NET_ADMIN`. Pin the pod with `nodeSelector` to a node that sits on
the same L2 segment as the machines you want to wake. `examples/helm-deployment.yaml` shows the same deployment as an
Argo CD `Application`.

## Development

```shell
go test ./...
go build -o wol-relay .
```

## Acknowledgements

The multi-arch build, with native `amd64` and `arm64` runners merged into one manifest, is adapted from
[sredevopsorg/multi-arch-docker-github-workflow](https://github.com/sredevopsorg/multi-arch-docker-github-workflow).
