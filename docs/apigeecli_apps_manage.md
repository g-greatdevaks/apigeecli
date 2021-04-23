## apigeecli apps manage

Approve or revoke a developer app

### Synopsis

Approve or revoke a developer app

```
apigeecli apps manage [flags]
```

### Options

```
  -x, --action string   Action to perform - revoke or approve (default "revoke")
  -d, --email string    Developer Email
  -h, --help            help for manage
  -n, --name string     Developer app name
```

### Options inherited from parent commands

```
  -a, --account string   Path Service Account private key in JSON
      --disable-check    Disable check for newer versions
  -o, --org string       Apigee organization name
  -t, --token string     Google OAuth Token
```

### SEE ALSO

* [apigeecli apps](apigeecli_apps.md)	 - Manage Apigee Developer Applications

###### Auto generated by spf13/cobra on 25-Feb-2021