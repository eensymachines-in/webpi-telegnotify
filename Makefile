BINARY_NAME = telegnotify
.DEFAULT_GOAL = run
export FLOG=0
export SILENT=0
export DEVICEREG_URL=http://aqua.eensymachines.in:30001/api/devices
export BOT_BASEURL=https://api.telegram.org/bot
export BOT_TOK=7003243457:AAFdkeOEWTXakLxz7HjyJFBkJiL8ME-tZvE
export BOT_UNAME=raspb_notifybot

test:
	go clean --testcache 
	go test -v -timeout 30s -run TestTelegGetMe
	go test -v -timeout 30s -run TestTelegSendMessage
	
build:
	go build -o ./${BINARY_NAME}

run: build
	./${BINARY_NAME}