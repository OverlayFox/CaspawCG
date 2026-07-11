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
	curl -sSfL https://golangci-lint.run/install.sh | sudo sh -s -- -b $(go env GOPATH)/bin v2.11.2
	golangci-lint --version
	@echo "--> Installing Wails dependencies..."
	sudo apt update
	sudo apt install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev libsoup-3.0-dev
	@echo "==> Installing Wails..."
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "Make sure to add `$(go env GOPATH)/bin` to your PATH if it's not already there."
	

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
