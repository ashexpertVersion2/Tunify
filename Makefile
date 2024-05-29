
fmt:
	go fmt -mod=vendor ./...

lint:
	go vet -mod=vendor ./...

tunify:
	go build -a -mod=vendor -o bin/tunify