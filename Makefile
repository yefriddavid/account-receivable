include .env

Author := yefriddavid
Version := $(shell git describe --abbrev=0 --tags | head -1) ##$(shell date "+%Y%m%d%H%M")
ReleaseDate := $(shell date "+%Y/%m/%d-%H:%M")
GitCommit := $(shell git rev-parse HEAD)
GitShortCommit := $(shell git rev-parse --short HEAD)
SysConfigFile := $(SYS_DEFAULT_TARGET_CONFIG_FILE)
DevConfigFile := $(DEV_SOURCE_CONFIG_FILE)
MyEmailPassword := $(MY_EMAIL_PASSWORD)
MyBinarySecret := $(shell apg -m16 | head -n 1)

LDFLAGS := -s -w -X main.secret=$(MyBinarySecret) -X main.Version=$(Version) -X main.GitCommit=$(GitCommit) -X main.Author=$(Author) -X main.GitShortCommit=$(GitShortCommit) -X main.ReleaseDate='$(ReleaseDate)'

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ".:*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run:
run:
	go run -ldflags "$(LDFLAGS)" cmd/main.go -configFile=$(shell pwd)/config.yml

decryptPass:
decryptPass:
	go run cmd/main.go -just-decrypt -secret-decrypt-pass=mySecret -raw-unsecured-text=myEncryptedText

encryptPass:
encryptPass:
	go run cmd/main.go -just-encrypt -raw-unsecured-text=myRawText

setConfigPass:
setConfigPass:
	$(eval hashedPassword := $(shell AccountsRecievable -just-encrypt -raw-unsecured-text=$(MyEmailPassword)))
	@echo $(hashedPassword)
	sed -i 's/\(password:\).*/\1 $(hashedPassword)/' config.yml

openPdf:
openPdf:
	xdg-open ../history/charge_account_2020_Jun.pdf

## local-publish: copy-local-config build binMv
local-publish: build binMv setConfigPass copy-local-config
local-publish:
	@echo published

configureVersion:
configureVersion:
	gvm use go1.15.2

#binMv:
#binMv:
#	sudo mv main /usr/local/bin/AccountsRecievable

build:
build: ## build application
	@go build -ldflags "$(LDFLAGS) -X main.SysConfigFile=$(SysConfigFile)" cmd/main.go

#copy-local-config:
#copy-local-config: ## Copy file settings
#	cp ./config.yml $(SysConfigFile)

show:
show:
	echo $(MyBinarySecret)

RestartService:
RestartService:
	sudo supervisorctl restart AccountsReceivable

StopService:
StopService:
	sudo supervisorctl stop AccountsReceivable

StartService:
StartService:
	sudo supervisorctl start AccountsReceivable

UpdateService:
UpdateService:
	sudo supervisorctl update


setExampleConfigFile:
setExampleConfigFile:
	cp config.yml example.config.yml
	sed -i 's/\(:\).*/\1 value/' example.config.yml


copy-configs:
copy-configs:
	sudo cp ./sign.png $(shell echo $(SysConfigFile) | xargs dirname)



local-release:
local-release:
	sudo rm -rf /usr/local/bin/AccountsRecievable
	sudo ln -s $(shell pwd)/dist/account-receivable_linux_amd64/account-receivable /usr/local/bin/AccountsRecievable
#	sudo mv main /usr/local/bin/AccountsRecievable
#sudo cp ./dist/traze-installer_linux_amd64/traze-installer /usr/local/bin/traze-installer

copy-local-config:
copy-local-config:
	sudo rm -rf /etc/AccountReceivable.yml
	sudo ln -s $(shell pwd)/config.yml /etc/AccountReceivable.yml


