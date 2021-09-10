package notifier

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
)

func init() {
	registerNotifier("email", NewEmailNotifier())
}

type EmailNotifier struct {
	sender     string
	name       string
	receiver   string
	smtpServer string
	smtpPort   int
	password   string
}

func NewEmailNotifier() Notifier {

	smtpPort, err := strconv.Atoi(os.Getenv("EMAIL_SMTPPORT"))
	if err != nil {
		smtpPort = 25
	}

	return &EmailNotifier{
		sender:     os.Getenv("EMAIL_SENDER"),
		name:       os.Getenv("EMAIL_SENDERNAME"),
		receiver:   os.Getenv("EMAIL_RECEIVER"),
		smtpServer: os.Getenv("EMAIL_SMTPSERVER"),
		smtpPort:   smtpPort,
		password:   os.Getenv("EMAIL_PASSWORD"),
	}
}

func (n *EmailNotifier) SendMessage(events []*Event) error {

	if len(events) == 0 {
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", n.sender)
	//m.SetHeader("From", fmt.Sprintf("%s <%s>", n.name, n.sender))
	m.SetHeader("To", strings.Split(n.receiver, ",")...)
	m.SetHeader("Subject", "腾讯云轻量服务器监控提醒通知")

	htmlContent := fmt.Sprintf(`
	<div style="margin: 10px 20px 20px 20px;text-align:center;">
	<span style="font-size:25px;line-height: 45px">腾讯云轻量服务器监控提醒通知</span><br>
	<span  style="font-size:14px;color: #5e6d82">%s</span>
	</div>
	<div style="display:flex;flex-wrap: wrap">
	`, time.Now().Format("2006年01月02日 15:04"))

	for _, event := range events {
		htmlContent = htmlContent +
			fmt.Sprintf(`<div style="width:340px;font-size: 14px;color: #5e6d82;border-radius:4px;margin:10px;border:1px solid #ebeef5;overflow:hidden;box-shadow: 0 2px 12px 0 rgb(0 0 0 / 10%%);">
			<div style="padding: 18px 20px;border-bottom: 1px solid #ebeef5;box-sizing: border-box;font-size: 16px">%s
			<span style="display:inline-block;padding:2px 10px;font-size:12px;background:%s;margin-left:10px;border-radius:4px;color:white">%s</span>
			</div>
			<div style="padding:10px 20px 30px 20px">
				<div style="margin:15px 0px 5px 0px">
					流量一共%s，当前使用%s，%s
				</div>
				<div style="width:300px;height:20px;margin:15px 0px;background:#ebeef5;position:relative;border-radius: 100px;overflow:hidden"> 
					<div style="width:%.2f%%;height:20px;background:%s;border-radius: 100px;"></div>
				</div>`, event.Name, GetStateColor(event.State), GetCNState(event.State), wellSize(event.Total), wellSize(event.Used),
				getTips(event.Used, event.Total), (float64(event.Used)/float64(event.Total))*100, getColor(event.Used, event.Total))

		if event.Action != Statistics {
			htmlContent = htmlContent + fmt.Sprintf(`<div>
			当前执行操作: 【%s】
		</div>`, event.Action)
		}

		htmlContent = htmlContent + `</div>
			</div>`

	}

	htmlContent = htmlContent + `</div>`

	m.SetBody("text/html", htmlContent)
	d := gomail.NewDialer(n.smtpServer, n.smtpPort, n.sender, n.password)

	return d.DialAndSend(m)

}
