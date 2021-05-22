package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
	//"reflect"
	"fmt"
	"github.com/divan/num2words"
	"github.com/jung-kurt/gofpdf"
	"strconv"
	"strings"
  structs "github.com/yefriddavid/AccountsReceivable/src/structs"
)

const (
	// layoutISO              = "2006-01-02"
	// layoutUS               = "January 2, 2006"
	presentationFormatDate = "2 January 2006"
)

var secret string = "yedcakKormOnesEn"

func GetSecret() string {
  return secret
}

func getHtmlTemplate(pdf *gofpdf.Fpdf, html gofpdf.HTMLBasicType, config structs.Config) (bool, error) {

	setting := config.Setting
	saleOrder := config.SaleOrder
	total, _ := strconv.Atoi(fmt.Sprintf("%.0f", saleOrder.Total))

	saleOrderBody := saleOrder.Content.Body.Raw
	saleOrderBodySign := saleOrder.Content.Sign.Raw
	saleOrderTitle := saleOrder.Content.Title.Raw
	saleOrderOwnTo := saleOrder.Content.OwnTo.Raw

	saleOrderTitle = strings.Replace(saleOrderTitle, "{SaleOrder.Number}", saleOrder.Number, -1)
	saleOrderTitle = strings.Replace(saleOrderTitle, "{SaleOrder.City}", saleOrder.City, -1)

	pdf.SetFont(saleOrder.Content.Title.Font.Family, saleOrder.Content.Title.Font.Style, saleOrder.Content.Title.Font.Size)
	_, lineHt := pdf.GetFontSize()

	html.Write(lineHt, saleOrderTitle)

	saleOrderOwnTo = strings.Replace(saleOrderOwnTo, "{Employee.FullName}", saleOrder.Employee.FullName, -1)
	saleOrderOwnTo = strings.Replace(saleOrderOwnTo, "{Employee.DocumentNumber}", saleOrder.Employee.DocumentNumber, -1)
	saleOrderOwnTo = strings.Replace(saleOrderOwnTo, "{Employee.PhoneNumber}", saleOrder.Employee.PhoneNumber, -1)
	saleOrderOwnTo = strings.Replace(saleOrderOwnTo, "{SaleOrder.Address}", saleOrder.Address, -1)

	html.Write(lineHt, saleOrderOwnTo)

	pdf.SetFont(saleOrder.Content.Body.Font.Family,
		saleOrder.Content.Body.Font.Style, saleOrder.Content.Body.Font.Size)
	_, lineHt = pdf.GetFontSize()

	saleOrderBody = strings.Replace(saleOrderBody, "{SaleOrder.NameOfMonth}", saleOrder.FormatDate.Format("January"), -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{SaleOrder.AmountLetters}", num2words.ConvertAnd(total), -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{SaleOrder.Amount}", fmt.Sprintf("%.0f", saleOrder.Total), -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{SaleOrder.TotalHours}", strconv.Itoa(saleOrder.TotalHours), -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{SaleOrder.CostPerHour}", fmt.Sprintf("%.2f", setting.CostPerHour), -1)

	if saleOrder.Bonus <= 0 {
		saleOrderBody = strings.Replace(saleOrderBody, "{BonusDescription}", "", -1)
	} else {
		bonusLetters := num2words.ConvertAnd(int(saleOrder.Bonus))
		bonusDescription := ` and ` + bonusLetters + ` dollars equivalent to ` + saleOrder.BonusDescription //+ "."

		saleOrderBody = strings.Replace(saleOrderBody, "{BonusDescription}", bonusDescription, -1)
	}

	saleOrderBody = strings.Replace(saleOrderBody, "{upper:AccountSaving.BankName}", strings.ToUpper(setting.AccountSaving.BankName), -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{title:AccountSaving.BankName}", strings.Title(strings.ToLower(setting.AccountSaving.BankName)), -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{AccountSaving.AccountNumber}", setting.AccountSaving.AccountNumber, -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{Employee.FullName}", saleOrder.Employee.FullName, -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{AccountSaving.SwiftCode}", setting.AccountSaving.SwiftCode, -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{AccountSaving.FullSwiftCode}", setting.AccountSaving.FullSwiftCode, -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{AccountSaving.Address}", setting.AccountSaving.FullSwiftCode, -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{SaleOrder.City}", saleOrder.City, -1)
	saleOrderBody = strings.Replace(saleOrderBody, "{Setting.PresentationFormatDate}", setting.FormatDate.Format(presentationFormatDate), -1)

	html.Write(lineHt, saleOrderBody)
	url := setting.Email.Template.SignImage.Path
	pdf.Image(url, setting.Email.Template.SignImage.Position.AxisX,
		setting.Email.Template.SignImage.Position.AxisY,
		setting.Email.Template.SignImage.Position.Rect,
		0, false, "", 0, "")

	saleOrderBodySign = strings.Replace(saleOrderBodySign, "{upper:Employee.FullName}", strings.ToUpper(saleOrder.Employee.FullName), -1)
	saleOrderBodySign = strings.Replace(saleOrderBodySign, "{Employee.DocumentNumber}", saleOrder.Employee.DocumentNumber, -1)
	saleOrderBodySign = strings.Replace(saleOrderBodySign, "{Employee.DocumentCity}", saleOrder.Employee.DocumentCity, -1)
	saleOrderBodySign = strings.Replace(saleOrderBodySign, "{upper:Employee.Position}", strings.ToUpper(saleOrder.Employee.Position), -1)

	pdf.SetFont(saleOrder.Content.Sign.Font.Family, saleOrder.Content.Sign.Font.Style, saleOrder.Content.Sign.Font.Size)
	html.Write(lineHt, saleOrderBodySign)

	return true, nil
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func deriveKey(passphrase string, salt []byte) ([]byte, []byte) {
	if salt == nil {
		salt = make([]byte, 8)
		// http://www.ietf.org/rfc/rfc2898.txt
		// Salt.
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(passphrase), salt, 1000, 32, sha256.New), salt
}

func Encrypt(passphrase, plaintext string) string {
	key, salt := deriveKey(passphrase, nil)
	iv := make([]byte, 12)
	// http://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf
	// Section 8.2
	rand.Read(iv)
	b, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(b)
	data := aesgcm.Seal(nil, iv, []byte(plaintext), nil)
	return hex.EncodeToString(salt) + "-" + hex.EncodeToString(iv) + "-" + hex.EncodeToString(data)
}

func Decrypt(passphrase, ciphertext string) string {
	arr := strings.Split(ciphertext, "-")
	salt, _ := hex.DecodeString(arr[0])
	iv, _ := hex.DecodeString(arr[1])
	data, _ := hex.DecodeString(arr[2])
	key, _ := deriveKey(passphrase, salt)
	b, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(b)
	data, _ = aesgcm.Open(nil, iv, data, nil)
	return string(data)
}
