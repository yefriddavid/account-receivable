package utils

import (
	//"reflect"
	"fmt"
	"github.com/jung-kurt/gofpdf"
  structs "github.com/yefriddavid/AccountsReceivable/src/structs"

)


func GeneratePdf(config structs.Config, fileName string) {

	pdf := gofpdf.New("P", "mm", "letter", "")
	pdf.AddPage()

	html := pdf.HTMLBasicNew()
	getHtmlTemplate(pdf, html, config)

	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		fmt.Println(err)
	}

}

