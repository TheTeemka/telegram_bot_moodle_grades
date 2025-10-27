.PHONY: load unload reload test

dev:
	go run main.go  --debug

test:
	@echo "Running go tests..."
	go test ./internal/model

load: test
	@echo "loading binary..."
	go build -o telegram_bot_moodle_grades .
	launchctl load ~/Library/LaunchAgents/com.temirlan_bayangazy.telegram_bot_moodle_grades.plist

reload: test
	@echo "Building binary..."
	go build -o telegram_bot_moodle_grades .
	$(MAKE) unload
	$(MAKE) load

unload:
	@echo "unloading binary..."
	launchctl unload ~/Library/LaunchAgents/com.temirlan_bayangazy.telegram_bot_moodle_grades.plist

status:
	launchctl list ~/Library/LaunchAgents/com.temirlan_bayangazy.telegram_bot_moodle_grades.plist