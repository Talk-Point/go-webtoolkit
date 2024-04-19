setup:
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

test:
	go test ./...

test-cover:
	go test -cover ./...

doc:
	pkgsite -open .