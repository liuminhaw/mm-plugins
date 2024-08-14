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
            # TODO: not implemented yet
            encoding = "SSH | PEM" 
        }
    }
    equipment "policies" "list" {
        attributes = {
            # The scope of the policy to mine for. Can be Local (default), AWS, or All.
            scope = "Local (default) | AWS | All"
        }
    }
    equipment "virtualMFADevices" "mine" {
        attributes = {
            # The status of the virtual MFA device to mine for. 
            # Can be Any (default), Assigned, or Unassigned.
            assignmentStatus = "Any (default) | Assigned | Unassigned"
        }
    }
}
```

## Building plugin
```bash
go build -o /path/to/mist-miner/plugins/bin/mm-iam .
```
