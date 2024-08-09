# POSH rclone provider

## Usage

### Plugin

```go
func New(l log.Logger) (plugin.Plugin, error) {
	// ...
  inst.commands.MustAdd(rcloud.NewCommand(l, inst.cache))
	// ...
}
```

### Config

```yaml
rclone:
  path: devops/config/rclone.conf
  config: |
    [cloudflare]
    type = s3
    provider = Cloudflare
    access_key_id = {{ op://<vault>/<item>/access key id }}
    secret_access_key = {{ op://<vault>/<item>/secret access key }}
    endpoint = {{ op://<vault>/<item>/endpoint }}
    no_check_bucket = true
    acl = private
```
