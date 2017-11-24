/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06  @author mengyuan
 */
package notify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"dev-flows-api-golang/models"

	"dev-flows-api-golang/models/alert"

	"github.com/astaxie/beego"
	"github.com/golang/glog"
	"github.com/prometheus/alertmanager/notify"
	"gopkg.in/gomail.v2"
)

// annotations in rules
const (
	conditionKey string = "condition"
	currentValue string = "currentValue"
)

var emailTemplate string

// EmailNotifier is an interface to send notify by email
type EmailNotifier struct {
	Receivers []string
	strategy  alert.NotifyStrategy
}

var ErrMultiOrNoMailConfig = errors.New("multi or no mail config")
var ErrWrongConfigFormat = errors.New("config in wrong format")

func getMailConfig() (config *models.MailConfig, err error) {
	var configs []models.Configs
	if configs, err = new(models.Configs).GetByType("mail"); err != nil {
		return
	}
	if configs == nil || len(configs) != 1 {
		err = ErrMultiOrNoMailConfig
		return
	}
	config = new(models.MailConfig)
	err = json.Unmarshal([]byte(configs[0].ConfigDetail), config)
	return
}

func spiltHostPort(mailServer string, secure bool) (host string, port int, err error) {
	hp := strings.Split(mailServer, ":")
	length := len(hp)
	if length < 1 || length > 2 {
		err = ErrWrongConfigFormat
		return
	}
	host = hp[0]
	if length == 1 {
		if secure {
			port = 465
		} else {
			port = 25
		}
	} else {
		var p64 int64
		if p64, err = strconv.ParseInt(hp[1], 10, 32); err != nil {
			return
		}
		port = int(p64)
	}
	return
}

// Notify implement Notifier.Notify

func (a *EmailNotifier) Notify(message notify.WebhookMessage) error {
	config, err := getMailConfig()
	if err != nil {
		return err
	}
	from := config.SenderMail
	user := config.SenderMail
	pwd := config.SenderPassword
	if err != nil {
		return err
	}
	host, port, err := spiltHostPort(config.MailServer, config.Secure)
	if err != nil {
		return err
	}
	glog.Infof("sending notification to %s\n", a.Receivers)
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", a.Receivers...)
	m.SetHeader("Subject", "[新智云]告警提醒")
	m.SetBody("text/html", a.genEmailBody(message))

	d := gomail.NewDialer(host, port, user, pwd)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send notification email failed, err: %s", err)
	}
	return nil
}

func init() {
	// get email template
	if emailBytes, err := ioutil.ReadFile("templates/emails/alert_warning.html"); err != nil {
		panic(err)
	} else {
		emailTemplate = string(emailBytes)
	}
}

func (a *EmailNotifier) genEmailBody(message notify.WebhookMessage) string {
	targetType := "服务"
	if a.strategy.TargetType == alert.StrategyTypeNode {
		targetType = "节点"
	}
	tableContent := bytes.Buffer{}
	tableContent.WriteString(`<table cellpadding="0" cellspacing="1" style="width:100%;background: grey;" border="0">
                         <tbody>
                           <tr style="text-align: center">
                             <td style="background:#ffffff;color: black"><b>监控对象</b></td>
                             <td style="background:#ffffff;color: black"><b>触发规则</b></td>
                             <td style="background:#ffffff;padding: 5px;color: black"><b>最近数据</b></td>
                           </tr>`)
	rowTemplate := `<tr style="text-align: center">
			   <td style="background:#ffffff;padding: 5px;padding: 5px;">%s</td>
			   <td style="background:#ffffff;padding: 5px;padding: 5px;">%s</td>
			   <td style="background:#ffffff;padding: 5px;">%s</td>
			  </tr>`
	for i := range message.Alerts {
		row := fmt.Sprintf(rowTemplate,
			a.strategy.TargetName,
			html.EscapeString(message.Alerts[i].Annotations[conditionKey]),
			html.EscapeString(message.Alerts[i].Annotations[currentValue]))
		tableContent.WriteString(row)
	}
	tableContent.WriteString(`</tbody></table>`)
	oldNew := []string{
		"%strageName%",
		a.strategy.StrategyName,
		"%targetName%",
		a.strategy.TargetName,
		"%targetType%",
		targetType,
		"%tableContent%",
		tableContent.String(),
		"%date%",
		time.Now().Format("2006-01-02"),
		"%systemEmail%",
		beego.AppConfig.String("email_service_email"),
	}
	r := strings.NewReplacer(oldNew...)
	return r.Replace(emailTemplate)
}
