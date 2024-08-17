# Plugin: mm-sqs
This plugin get information from aws sqs service

## Config
Configuration settings for mist-miner
```hcl
plug "mm-sqs" "group" {
    authenticator = {
        profile = "aws profile name for accessing aws account"
        regions = "us-east1,ap-northeast-1,... (default read all available regions if not set)"
    }
}
```

## Building plugin
```bash
go build -o /path/to/mist-miner/plugins/bin/mm-sqs .
```
