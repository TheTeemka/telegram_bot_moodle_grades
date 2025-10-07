.PHONY: load unload reload

load:
	launchctl load ~/Library/LaunchAgents/com.temirlan_bayangazy.telegram_bot_moodle_grades.plist

unload:
	launchctl unload ~/Library/LaunchAgents/com.temirlan_bayangazy.telegram_bot_moodle_grades.plist

reload:
	go build -o telegram_bot_moodle_grades . 
	$(MAKE) unload
	$(MAKE) load