package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/robfig/cron"
	"golang.org/x/crypto/pbkdf2"
	//"reflect"
	"flag"
	"fmt"
	"github.com/divan/num2words"
	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var SysConfigFile = ""
var secret string = "yedcakKormOnesEn"

var (
	configFile       = flag.String("configFile", "", "Configuration file")
	stringFormatDate = flag.String("format-date", "2006-01-02", "Format of date")

	// arguments for command tools
	secretDecriptPass = flag.String("secret-decrypt-pass", "", "Secret for decript pass")
	justEncrypt       = flag.Bool("just-encrypt", false, "Just Encrypt passwowrd")
	justDecrypt       = flag.Bool("just-decrypt", false, "Just Decrypt passwowrd")
	rawUnsecureText   = flag.String("raw-unsecured-text", "", "Raw unsecured text")
)

func init() {

	flag.Parse()

	if *configFile == "" && SysConfigFile != "" {
		*configFile = SysConfigFile
	}

	if fileExists(*configFile) {
		// fmt.Println("File exist")
	} else {

		if !fileExists("config.yml") {
			*configFile = "config.yml"
		} else {
			fmt.Println("Configuration File does not exist!")
			//os.Exit(2)
		}
	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Println(err)
		return false
	}
	return !info.IsDir()
}

const (
	layoutISO              = "2006-01-02"
	layoutUS               = "January 2, 2006"
	presentationFormatDate = "2 January 2006"
)

type ItemFont struct {
	Family string
	Size   float64
	Style  string
}

type SaleOrderContentItem struct {
	Font ItemFont
	Raw  string
}
type SaleOrderContent struct {
	Title SaleOrderContentItem
	OwnTo SaleOrderContentItem `mapstructure:"own-to"`
	Body  SaleOrderContentItem
	Sign  SaleOrderContentItem
}

type SignImagePosition struct {
	AxisY float64 `mapstructure:"axis-y"`
	AxisX float64 `mapstructure:"axis-x"`
	//X float64`mapstructure:"y"`
	//Y float64
	Rect float64
}
type SignImage struct {
	Position SignImagePosition
	Path     string
}
type Template struct {
	SignPathImage string    `mapstructure:"sign-path-image"`
	SignImage     SignImage `mapstructure:"sign-image"`
}
type Email struct {
	From     EmailConfig
	To       EmailConfig
	Cc       EmailConfig
	Subject  string
	Body     string
	Template Template
}
type EmailConfig struct {
	Email    string
	FullName string `mapstructure:"full-name"`
}

type AccountSavings struct {
	BankName      string `mapstructure:"bank-name"`
	AccountNumber string `mapstructure:"account-number"`
	SwiftCode     string `mapstructure:"swift-code"`
	FullSwiftCode string `mapstructure:"full-swift-code"`
	Address       string
}

type Settings struct {
	CronReminder     string `mapstructure:"cron-reminder"`
	TimeZone         string `mapstructure:"time-zone"`
	OutputFilePrefix string `mapstructure:"output-file-prefix"`
	StringFormatDate string `mapstructure:"string-format-date"`
	Email            Email
	CostPerHour      float32 `mapstructure:"cost-per-hour"`
	HistoryFolder    string  `mapstructure:"history-folder"`
	//Date time.Time
	Date          string
	FormatDate    time.Time
	AccountSaving AccountSavings `mapstructure:"account-saving"`
	//WorkFrom time.Time `mapstructure:"work-from"` // date of start
	WorkFrom string `mapstructure:"work-from"` // date of start
}

type SaleOrder struct {
	Content          SaleOrderContent
	TotalHours       int `mapstructure:"total-hours"`
	Bonus            float32
	BonusDescription string `mapstructure:"bonus-description"`
	BodySign         string `mapstructure:"body-sign"`
	Number           string
	Total            float32
	City             string
	Address          string
	Employee         Employee
	Date             string
	Body             string
	FormatDate       time.Time
	// Date time.Time
}

type Employee struct {
	FullName       string `mapstructure:"full-name"`
	DocumentNumber string `mapstructure:"document-number"`
	DocumentCity   string `mapstructure:"document-city"`
	PhoneNumber    string `mapstructure:"phone-number"`
	Position       string
}

type SmtpConfig struct {
	Port     int
	Username string
	Password string
	Smtp     string
}

type Config struct {
	Setting   Settings
	SaleOrder SaleOrder `mapstructure:"sale-order"`
	Smtp      SmtpConfig
}

func main() {

	flag.Parse()
	if *justEncrypt == true {
		result := encrypt(secret, *rawUnsecureText)
		fmt.Println(result)

		return
	}
	if *justDecrypt == true {
		// result := decrypt(secret, *rawUnsecureText)
		result := decrypt(*secretDecriptPass, *rawUnsecureText)
		fmt.Println(result)

		return
	}

	var config Config
	config, _ = loadSetting()

	if config.Setting.CronReminder == "" {
		startProgram(config)
	} else {
		fmt.Println("Start with cron")
		fmt.Println(config.Setting.CronReminder)
		done := make(chan bool)

		nyc, _ := time.LoadLocation(config.Setting.TimeZone)
		c := cron.New(cron.WithLocation(nyc))
		c.AddFunc(config.Setting.CronReminder, func() { startProgram(config) })
		c.Start()

		<-done
	}
}

func startProgram(config Config) {

	fmt.Println("Start taks")
	var setting Settings
	var accountSaving AccountSavings
	var employee Employee
	var saleOrder SaleOrder
	//var saleOrderDate time.Time
	now := time.Now()

	setting = config.Setting
	saleOrder = config.SaleOrder
	employee = saleOrder.Employee

	if config.SaleOrder.Date == "" {
		config.SaleOrder.FormatDate = now.AddDate(0, -1, 0)
	} else {
		config.SaleOrder.FormatDate, _ = time.Parse(config.Setting.StringFormatDate, config.SaleOrder.Date)
	}

	if setting.StringFormatDate == "" {
		config.Setting.StringFormatDate = *stringFormatDate
	}

	workFrom, _ := time.Parse(config.Setting.StringFormatDate, setting.WorkFrom)

	diffMonths := monthsCountSince(workFrom)
	config.SaleOrder.Number = `0` + strconv.Itoa(diffMonths)

	if config.SaleOrder.Total <= 0 {
		config.SaleOrder.Total = setting.CostPerHour * float32(saleOrder.TotalHours)
		config.SaleOrder.Total += saleOrder.Bonus
	}

	if setting.Date == "" {
		config.Setting.FormatDate = now //now.Format(config.Setting.StringFormatDate)
	} else {
		config.Setting.FormatDate, _ = time.Parse(config.Setting.StringFormatDate, config.Setting.Date)
	}

	setting.AccountSaving = accountSaving
	saleOrderYear := config.SaleOrder.FormatDate.Format("2006")
	saleOrderMonth := config.SaleOrder.FormatDate.Format("Jan")

	fileName := config.Setting.HistoryFolder + "/" + config.Setting.OutputFilePrefix + saleOrderYear + "_" + saleOrderMonth + ".pdf"

	total := fmt.Sprintf("%.0f", config.SaleOrder.Total)
	subject := config.Setting.Email.Subject
	subject = strings.Replace(subject, "{SaleOrder.Month}", saleOrderMonth, -1)
	subject = strings.Replace(subject, "{SaleOrder.Year}", saleOrderYear, -1)
	subject = strings.Replace(subject, "{Employee.FullName}", employee.FullName, -1)
	subject = strings.Replace(subject, "{SaleOrder.Total}", total, -1)

	if saleOrder.Bonus > 0 {
		bonusDescription := " and " + saleOrder.BonusDescription + " "
		subject = strings.Replace(subject, "{AdditionalSubject}", bonusDescription, -1)
	} else {
		subject = strings.Replace(subject, "{AdditionalSubject}", " ", -1)
	}

	config.Setting.Email.Subject = subject

	body := config.Setting.Email.Body
	body = strings.Replace(body, "{SaleOrder.NameOfMonth}", config.SaleOrder.FormatDate.Format("January"), -1)

	config.Setting.Email.Body = body

	generatePdf(config, fileName)
	send(body, config, fileName)

}

func generatePdf(config Config, fileName string) {

	pdf := gofpdf.New("P", "mm", "letter", "")
	pdf.AddPage()

	html := pdf.HTMLBasicNew()
	getHtmlTemplate(pdf, html, config)

	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		fmt.Println(err)
	}

}

func getHtmlTemplate(pdf *gofpdf.Fpdf, html gofpdf.HTMLBasicType, config Config) (bool, error) {

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

func send(body string, config Config, fileName string) {

	m := gomail.NewMessage()
	m.SetAddressHeader("From", config.Setting.Email.From.Email, config.Setting.Email.From.FullName)
	m.SetAddressHeader("Cc", config.Setting.Email.Cc.Email, config.Setting.Email.Cc.FullName)

	m.SetHeader("To", config.Setting.Email.To.Email)
	m.SetHeader("Subject", config.Setting.Email.Subject)
	m.SetBody("text/html", body)
	m.Attach(fileName)

	pass := decrypt(secret, config.Smtp.Password)
	d := gomail.NewPlainDialer(config.Smtp.Smtp, config.Smtp.Port, config.Smtp.Username, pass)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func monthsCountSince(createdAtTime time.Time) int {
	now := time.Now()
	months := 0
	month := createdAtTime.Month()
	for createdAtTime.Before(now) {
		createdAtTime = createdAtTime.Add(time.Hour * 24)
		nextMonth := createdAtTime.Month()
		if nextMonth != month {
			months++
		}
		month = nextMonth
	}

	return months
}

//func loadSetting() (settings Settings, err error) {
func loadSetting() (config Config, err error) {
	v := viper.New()
	if *configFile == "" {
		v.AddConfigPath("./")
	} else {
		dir, file := path.Split(*configFile)
		ext := path.Ext(file)
		var absoluteFileName string
		if ext == "" {
			absoluteFileName = file
		} else {
			absoluteFileName = strings.TrimRight(file, ext)
		}
		v.AddConfigPath(dir)
		v.SetConfigName(absoluteFileName)

	}
	err = v.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	v.Unmarshal(&config)
	return
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

func encrypt(passphrase, plaintext string) string {
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

func decrypt(passphrase, ciphertext string) string {
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
