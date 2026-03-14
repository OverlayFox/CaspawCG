.PHONY: run-wails
run-wails:
	wails dev -tags webkit2_41

.PHONY: build-wails
build-wails:
	wails build -tags webkit2_41

.PHONY: install-dev
install-dev:
	@echo "==> Installing development dependencies..."
	@echo "--> Installing golangci-lint..."
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.11.2
	golangci-lint --version

.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix