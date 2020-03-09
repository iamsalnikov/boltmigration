test:
	go test ./... -v

lint:
	docker run --rm -it \
		-v `pwd`:/mig \
		golang:latest \
		bash -c 'go get -u golang.org/x/lint/golint && cd /mig && golint -set_exit_status'