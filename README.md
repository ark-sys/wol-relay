# WOL Relay
`wol-relay` is a lightweight HTTP-to-UDP Wake-on-LAN relay designed for containerized environments (Kubernetes, Docker) where the initiating service may not have direct broadcast permissions or L2 connectivity to the target segment.
## Features
- **Stateless HTTP API**: Exposes a simple JSON structure to trigger a WoL magic packet.
- **Micro-Footprint**: Written in pure Go without heavy external dependencies.
- **Cloud-Native**: Easily deployed via Kubernetes/Helm. 
- **Multi-Arch**: Pre-built Docker images for AMD64 and ARM64.
## Quickstart
```bash
docker run -p 8089:8089 ghcr.io/ark-sys/wol-relay:latest
```
Send a trigger:
```bash
curl -X POST http://localhost:8089/wake \
  -H "Content-Type: application/json" \
  -d '{"mac":"AA:BB:CC:DD:EE:FF"}'
```
## Helm Deployment
`wol-relay` provides an official Helm chart for seamless Kubernetes deployment.
```bash
helm upgrade --install wol-relay ./charts/wol-relay \
  --namespace wol-relay --create-namespace \
  --set env.DEFAULT_BROADCAST="192.168.1.255"
```
*Note: Due to the nature of Wake-On-LAN Layer 2 broadcasts, the container uses `hostNetwork: true` by default.*
Check the `examples/helm-deployment.yaml` for an ArgoCD Application deployment snippet!
## Acknowledgements
The multi-arch CI/CD pipeline natively executing across separated ARM64 and AMD64 Github Runners was heavily inspired by the exceptional work from [sredevopsorg/multi-arch-docker-github-workflow](https://github.com/sredevopsorg/multi-arch-docker-github-workflow/blob/main/.github/workflows/multi-build.yaml).
