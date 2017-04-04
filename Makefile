test:
	bash -c 'rm -f coverage.txt && ls -d */ | while read dir; do go test -coverprofile=$${dir:0:-1}.cover.out -covermode=atomic ./$$dir; done && cat *.cover.out >> coverage.txt && rm *cover.out'
build:
	go get -t -v ./...