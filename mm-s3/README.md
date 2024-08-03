# Plugin: mm-s3
This plugin get information from aws s3 service

## Config
Configuration settings for mist-miner
```hcl
plug "mm-s3" "GROUP_NAME" {
    authenticator = {
        profile = "aws profile name for accessing aws account"
    }
}
```

## Building plugin
```bash
go build -o /path/to/mist-miner/plugins/bin/mm-s3 .
```
