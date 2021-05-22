package utils

import (
	// "fmt"
	"gopkg.in/gomail.v2"
	//"strings"
  structs "github.com/yefriddavid/AccountsReceivable/src/structs"

)


func Send(body string, email structs.FormatEmail, fileName string) {

	m := gomail.NewMessage()
	m.SetAddressHeader("From", email.EmailFrom, email.EmailFromFullName)
	m.SetAddressHeader("Cc", email.EmailCc, email.EmailCcFullName)

	m.SetHeader("To", email.EmailTo)
	m.SetHeader("Subject", email.Subject)
	m.SetBody("text/html", body)
	m.Attach(fileName)

	d := gomail.NewPlainDialer(email.Smtp, email.Port, email.Username, email.Pass)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
