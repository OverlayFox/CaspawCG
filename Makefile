.PHONY: run-wails
run-wails:
	wails dev -tags webkit2_41

.PHONY: build-wails
build-wails:
	wails build -tags webkit2_41