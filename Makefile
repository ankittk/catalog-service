PROTO_DIR=proto
CMD_MAIN=./cmd/main.go

.PHONY: generate run

generate:
	cd $(PROTO_DIR) && buf generate

run:
	go run $(CMD_MAIN)
