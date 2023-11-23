# POSH etcd provider

## Usage

## Configuration

```yaml
etcd:
  configPath: .posh/config/etcd
  clusters:
    - name: prod
      podName: etcd-0
      namespace: etcd
      paths: [ "cluster-prod.yaml" ]
    - name: stage
      podName: etcd-0
      namespace: etcd
      paths: [ "cluster-stage.yaml" ]
```
