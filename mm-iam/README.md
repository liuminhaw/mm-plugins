# Plugin: mm-iam
This plugin get information from aws iam service

## Config
Configuration settings for mist-miner
```hcl
plug "mm-iam" "GROUP" {
    authenticator {
        profile = "aws profile name with access to iam service"
    }
}
```
