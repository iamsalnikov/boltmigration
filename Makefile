test:
	go test ./... -v

lint:
	golint -set_exit_status