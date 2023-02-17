# oc-go

## Dev

#### nats-server
```
nats-server -m 8222 --js -D
```

#### create stream
```
nats stream add
> LOG
> log.*
> ... (default)
```

#### pub
```
nats pub log.decrypt "94fe793a9f55097134b6dbc0d55faaa3ab88aeede51e97ffbab3b785f5" --count=1000 --sleep=200ms
```