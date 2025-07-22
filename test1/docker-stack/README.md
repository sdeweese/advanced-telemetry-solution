# TorIX Docker Stack

## Deploying on remote host

### torix-tats-ingest-1

```bash
# create docker context
docker context create torix-tats-ingest-1 --docker "host=ssh://torix-tats-ingest-1"

# copy configs dir to remote host
rsync -azP --delete configs/ingest torix-tats-ingest-1:configs/

# deploy stack
DOCKER_CONTEXT=torix-tats-ingest-1 docker compose -f docker-compose.ingest.yml --env-file .env.torix-tats-ingest-1.dev up -d
```

### torix-tats-ingest-2

```bash
# create docker context
docker context create torix-tats-ingest-2 --docker "host=ssh://torix-tats-ingest-2"

# copy configs dir to remote host
rsync -azP --delete configs/ingest torix-tats-ingest-2:configs/

# deploy stack
DOCKER_CONTEXT=torix-tats-ingest-2 docker compose -f docker-compose.ingest.yml --env-file .env.torix-tats-ingest-2.dev up -d
```


### torix-tats-influx-1

```bash
# create docker context
docker context create torix-tats-influx-1 --docker "host=ssh://torix-tats-influx-1"

# copy configs dir to remote host
rsync -azP --delete configs/influx torix-tats-influx-1:configs/

# deploy stack
DOCKER_CONTEXT=torix-tats-influx-1 docker compose -f docker-compose.influx.yml --env-file .env.torix-tats-influx-1.dev up -d
```
