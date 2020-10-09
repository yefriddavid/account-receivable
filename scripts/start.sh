#!/bin/sh

notify-send 'Cuenta de cobro' 'Se esta enviando la cuenta de cobro'

#node createAttachment.js
notify-send 'Cuenta de cobro' 'Se esta enviando el correo'

go run cmd/main.go -configFile=/mnt/Zeus/Workspace/traze/docs/ChargeAccounts/config.yml


