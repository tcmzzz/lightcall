PWD    := $(shell pwd)
CERT_DIR:= $(PWD)/example/dev/run/proxy-cert
#UID:= $(shell id -u)
#GID:= $(shell id -g)
#TS:= $(shell date +%s)


.PHONY: cert
cert:
	mkdir -p $(CERT_DIR)
	cd $(CERT_DIR) && CAROOT=$(CERT_DIR) mkcert ring.dev.local
	CAROOT=$(CERT_DIR) mkcert -install #Install the local CA in the system trust store.

.PHONY: img
img:
	docker compose -f example/dev/docker-compose.yaml build

.PHONY: dev
dev:
	docker compose -f example/dev/docker-compose.yaml up && docker compose -f example/dev/docker-compose.yaml down

.PHONY: lintfront
lintfront:
	docker compose -f example/dev/docker-compose.yaml exec frontend pnpm lint
	docker compose -f example/dev/docker-compose.yaml exec frontend pnpm format

.PHONY: lintgo
lintgo:
	golangci-lint run --fix > /dev/null || golangci-lint run
