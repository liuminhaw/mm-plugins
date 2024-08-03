# Plugin: mm-iam
This plugin get information from aws iam service

## Config
Configuration settings for mist-miner
```hcl
plug "mm-iam" "GROUP_NAME" {
    authenticator = {
        profile = "aws profile name for accessing aws account"
    }
    equipment "user" "sshPublicKey" {
        attributes = {
            encoding = "SSH | PEM" (Todo)
        }
    }
    equipment "policies" "list" {
        attributes = {
            scope = "Local (default) | AWS | All"
        }
    }
    equipment "virtualMFADevices" "mine" {
        attributes = {
            assignmentStatus = "Any (default) | Assigned | Unassigned"
        }
    }
}
```

## Building plugin
```bash
go build -o /path/to/mist-miner/plugins/bin/mm-iam .
```
