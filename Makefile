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
	@if [ -n "$$(git status --porcelain)" ]; then \
		git add .; \
		git commit -m "chore: auto-update dependencies"; \
		git push origin develop; \
		echo "# Automated Dependency Update\n\n## Summary\n" > /tmp/pr_body.txt; \
		git diff HEAD~1 | ollama run qwen2.5-coder:32b "Analyze this git diff and provide a concise summary of the dependency updates. Focus on major version changes and breaking changes if any. Format the response in markdown." >> /tmp/pr_body.txt; \
		echo "\n## Detailed Changes\n\`\`\`" >> /tmp/pr_body.txt; \
		git diff HEAD~1 go.mod >> /tmp/pr_body.txt; \
		echo "\`\`\`" >> /tmp/pr_body.txt; \
		gh pr create --base master --head develop \
			--title "chore: auto-update dependencies" \
			--body-file /tmp/pr_body.txt \
			--label auto-merge; \
		rm /tmp/pr_body.txt; \
	else \
		echo "No changes to commit"; \
	fi