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

## Recv Log

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/recv-log cmd/recv-log/main.go
```

#### Runtime

```
[Unit]
Description=Recv Log Server
After=network-online.target

[Service]
Type=simple
ExecStart=/opt/recv-log/bin/recv-log /opt/recv-log/recv-log.conf

[Install]
WantedBy=multi-user.target
```

## Nats Decrypt

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/nats-decrypt cmd/nats-decrypt/main.go
```

#### Runtime

```
touch /usr/lib/systemd/system/nats-decrypt.service@.service
vim /usr/lib/systemd/system/nats-decrypt.service@.service
systemctl daemon-reload
systemctl start nats-decrypt.service@{1..2}
```

```
[Unit]
Description=Nats Client nats-decrypt %i
After=network-online.target
#StartLimitIntervalSec=300
#StartLimitBurst=5

[Service]
Type=simple
ExecStart=/opt/nats-decrypt/bin/nats-decrypt /opt/nats-decrypt/nats-decrypt.conf
#Restart=on-failure
#RestartSec=600

[Install]
WantedBy=multi-user.target
```

## Build Key

![image](https://i.imgur.com/xRlGf0p.jpg)

## Nats Server

#### Add Stream
```
Stream Name: LOG
Subjects: encrypt.>
```