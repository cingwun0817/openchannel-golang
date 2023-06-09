# oc-go

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
Message TTL: 1d
```

## Prometheus API

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/prometheus-api cmd/prometheus-api/main.go
```

#### Runtime

```
[Unit]
Description=Prometheus API
After=network-online.target

[Service]
Type=simple
ExecStart=/opt/prometheus-api/bin/prometheus-api /opt/prometheus-api/prometheus-api.conf

[Install]
WantedBy=multi-user.target
```

## People hour analyze

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/people-hour-analyze cmd/people-hour-analyze/main.go
```

## Scylla API

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/scylla-api cmd/scylla-api/main.go
```

## Target Audience Store Analyze

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/ta-analyze cmd/ta-analyze/main.go
```

## Insert Media Data

#### Build

(Linux/AMD64)
```
GOOS=linux GOARCH=amd64 go build -o bin/insert-media-data cmd/insert-media-data/main.go
```