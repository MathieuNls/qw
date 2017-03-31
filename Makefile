test:
	go test -v --coverprofile coverage ./query
	go tool cover -func=coverage
	go test -v --coverprofile coverage ./connector
	go tool cover -func=coverage