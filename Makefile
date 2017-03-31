test:
	go test -v --coverprofile coverage.txt ./query
	go tool cover -func=coverage
	go test -v --coverprofile coverage.txt ./connector
	go tool cover -func=coverage