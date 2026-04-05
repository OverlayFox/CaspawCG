.PHONY: dev
dev:
	wails dev -tags webkit2_41

.PHONY: build
build:
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

.PHONY: caspar-server
caspar-server:
	@echo "==> Starting CasparCG Server..."
	cd caspar && \
	export QT_X11_NO_MITSHM=1 && \
	export DISPLAY=:0 && \
	casparcg-server-2.5 casparcg.config

.PHONY: caspar-scanner
caspar-scanner:
	@echo "==> Starting CasparCG Scanner..."
	cd caspar && \
	casparcg-scanner casparcg.config
