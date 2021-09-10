package notifier

import (
	"os"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

func init() {
	registerNotifier("sms", NewSMSNotifier())
}

type SMSNotifier struct {
	secretId   string
	secretKey  string
	receiver   string
	appId      string
	signName   string
	templateId string
	region     string
	Subscribe  string
}

func NewSMSNotifier() Notifier {
	return &SMSNotifier{
		secretId:   os.Getenv("SECRET_ID"),
		secretKey:  os.Getenv("SECRET_KEY"),
		receiver:   os.Getenv("SMS_RECEIVER"),
		signName:   os.Getenv("SMS_SIGNNAME"),
		appId:      os.Getenv("SMS_APPID"),
		templateId: os.Getenv("SMS_TEMPLATEID"),
		region:     os.Getenv("SMS_REGION"),
		Subscribe:  os.Getenv("SMS_SUBSCRIBE"),
	}
}

func (m *SMSNotifier) SendMessage(events []*Event) error {

	if len(events) == 0 {
		return nil
	}

	credential := common.NewCredential(
		m.secretId,
		m.secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	client, _ := sms.NewClient(credential, m.region, cpf)

	for _, event := range events {

		if strings.Contains(m.Subscribe, string(event.Action)) {

			request := sms.NewSendSmsRequest()

			request.PhoneNumberSet = common.StringPtrs(strings.Split(m.receiver, ","))
			request.SmsSdkAppId = common.StringPtr(m.appId)
			request.SignName = common.StringPtr(m.signName)
			request.TemplateId = common.StringPtr(m.templateId)
			request.TemplateParamSet = common.StringPtrs([]string{event.Name, wellSize(event.Total), wellSize(event.Total - event.Used),
				string(event.Action)})
			_, err := client.SendSms(request)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
