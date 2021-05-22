package main

import (
	"github.com/robfig/cron"
	//"reflect"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
  "github.com/gen2brain/beeep"
  structs "github.com/yefriddavid/AccountsReceivable/src/structs"
  utils "github.com/yefriddavid/AccountsReceivable/src/utils"

)

var Version = "No Provided"
var GitCommit = "No Provided"
var GitShortCommit = "No Provided"
var Date = "No Provided"
var VersionStr = ""

var Author = ""
var Homepage = ""
var ReleaseDate = ""


//var SysConfigFile = ""

var (
	configFile       = flag.String("configFile", "/etc/AccountReceivable.yml", "Configuration file")
	stringFormatDate = flag.String("format-date", "2006-01-02", "Format of date")
  showVersion        = flag.Bool("version", false, "Show version")
  showEmailTo        = flag.Bool("showEmailTo", false, "Show email to")
  specificEmailToId         = flag.Int("specificEmailToId", 0, "Specific email to")


	// arguments for command tools
	secretDecriptPass = flag.String("secret-decrypt-pass", "", "Secret for decript pass")
	justEncrypt       = flag.Bool("just-encrypt", false, "Just Encrypt passwowrd")
	justDecrypt       = flag.Bool("just-decrypt", false, "Just Decrypt passwowrd")
	rawUnsecureText   = flag.String("raw-unsecured-text", "", "Raw unsecured text")
)

func init() {

	flag.Parse()

	//if *configFile == "" && SysConfigFile != "" {
	/*if *configFile == "" && SysConfigFile != "" {
		*configFile = SysConfigFile
	}*/

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

func main() {

	flag.Parse()

	if *showVersion {
		//fmt.Println(Version)
		showAppInfo()
		return
	}


	if *justEncrypt == true {
		result := utils.Encrypt(utils.GetSecret(), *rawUnsecureText)
		fmt.Println(result)

		return
	}
	if *justDecrypt == true {
		// result := decrypt(secret, *rawUnsecureText)
		result := utils.Decrypt(*secretDecriptPass, *rawUnsecureText)
		fmt.Println(result)

		return
	}

	var config structs.Config
	config, _ = loadSetting()
  formatEmail := GetFormatEmail(config, *specificEmailToId)

	if *showEmailTo {
    fmt.Println(formatEmail.EmailTo)
		return
	}


	if config.Setting.Cron == "" {
		startProgram(config, formatEmail)
	} else {
		fmt.Println("Start with cron")
		fmt.Println(config.Setting.Cron)
		done := make(chan bool)

		nyc, _ := time.LoadLocation(config.Setting.TimeZone)
		c := cron.NewWithLocation(nyc)
		//c := cron.New(cron.WithLocation(nyc))

		// load task
    c.AddFunc(config.Setting.Cron, func() { startProgram(config, formatEmail) })

		// Load reminders
    for _, reminder := range config.Setting.CronReminders {
      c.AddFunc(reminder, func() { beeep.Notify("Account Recievables", "next send account receivable: " + config.Setting.Cron, "") })
    }

		c.Start()

		<-done
	}
}

func startProgram(config structs.Config, formatEmail structs.FormatEmail) {

	fmt.Println("Start taks")
	var setting structs.Settings
	var accountSaving structs.AccountSavings
	var employee structs.Employee
	var saleOrder structs.SaleOrder
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

	utils.GeneratePdf(config, fileName)
	utils.Send(body, formatEmail, fileName)

}

func GetFormatEmail(config structs.Config, specificEmailToId int) (email structs.FormatEmail) {

  email.EmailFrom = config.Setting.Email.From.Email
  email.EmailTo = config.Setting.Email.To[specificEmailToId].Email
  email.EmailCc = config.Setting.Email.Cc.Email
  email.Subject = config.Setting.Email.Subject
	email.Pass = utils.Decrypt(utils.GetSecret(), config.Smtp.Password)
  email.Username = config.Smtp.Username
  email.Port = config.Smtp.Port
  email.Smtp = config.Smtp.Smtp
  email.EmailFromFullName = config.Setting.Email.From.FullName
  email.EmailCcFullName = config.Setting.Email.Cc.FullName

  return
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
func loadSetting() (config structs.Config, err error) {
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




func showAppInfo() {
	fmt.Printf("ReleaseDate: %s\n", ReleaseDate)
	fmt.Printf("Revision: %s\n", GitCommit)
	fmt.Printf("Short Revision: %s\n", GitShortCommit)
	fmt.Printf("Author: %s\n", Author)
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Homepage: %s\n", Homepage)
}

