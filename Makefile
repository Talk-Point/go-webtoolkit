setup:
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

test:
	go test ./...

test-cover:
	go test -cover ./...

doc:
	pkgsite -open .

update:
	git pull
	go get -u ./...
	go mod tidy
	# Prüfen ob es Änderungen gibt und diese committen
	@if [ -n "$$(git status --porcelain)" ]; then \
		git add .; \
		git commit -m "chore: auto-update dependencies"; \
		gh pr create --base master --head develop --title "chore: auto-update dependencies" --label auto-merge; \
	else \
		echo "No changes to commit"; \
	fi