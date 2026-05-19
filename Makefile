.PHONY: dev seed

dev:
	air -c .air.linux.toml

seed:
	go run cmd/seed/main.go
