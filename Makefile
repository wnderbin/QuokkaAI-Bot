quokka-run:
	AES_KEY=52af68c49d7c22ae9f23e489de6c6639bf20d930bceca4c52ed6983e94399996 CONFIG_PATH=./config/config.yaml go run main.go
quokka-build:
	go build main.go