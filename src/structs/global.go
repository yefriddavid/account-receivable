package structs

import (
	"time"
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
	To       []EmailConfig
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
	Cron             string
	CronReminders    []string `mapstructure:"cron-reminders"`
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



type FormatEmail struct {
  EmailFrom string
  EmailTo string
  EmailCc string
  Subject string
	Pass string
  Username string
  Port int
  Smtp string
  EmailFromFullName string
  EmailCcFullName string
}

