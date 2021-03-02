
## How to use

1. Install osquery

```
brew install osquery
```

2. Compile extension

```
go build -o ksql cmd/extension/main.go
```

3. Run 

```
KSQL_LOG_LEVEL=error osqueryi --extension ./ksql    
```

4. Try query

```
osquery> select * from k8s_contexts;
```

## Troubleshooting access problems

Some k8s clusters are discoverable through kube config, but are behind some kind of firewall.
Many queries will try to access those clusters. There is no 100% working workaround for that yet.

## Available tables discovery

All statically defined tables have `k8s_` prefix, to discover those, run 

```
osquery> .tables k8s_
```

## Warning

`k8s_env_vars` table will show secrets (from env vars) in plaintext.
