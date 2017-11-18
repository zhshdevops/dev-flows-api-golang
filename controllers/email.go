package controllers

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"dev-flows-api-golang/models/cluster"
	"dev-flows-api-golang/models"
	"github.com/golang/glog"
	"gopkg.in/gomail.v2"
)

// EmailNotifier is an interface to send notify by email

var ErrMultiOrNoMailConfig = errors.New("multi or no mail config")
var ErrWrongConfigFormat = errors.New("config in wrong format")

func getMailConfig() (config cluster.MailConfig, err error) {
	var configs []cluster.Configs
	if configs, err = new(cluster.Configs).GetByType("mail"); err != nil {
		return
	}
	if configs == nil || len(configs) != 1 {
		err = ErrMultiOrNoMailConfig
		return
	}
	err = json.Unmarshal([]byte(configs[0].ConfigDetail), &config)
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
			port = 465 // use SSL
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

type EmailDetail struct {
	Type    string //ci cd
	Result  string //failed success
	Subject string
	Body    string
}

// Notify implement Notifier.Notify
func (detail *EmailDetail) SendEmailUsingFlowConfig(namespace, flowId string) error {

	method := "SendEmailUsingFlowConfig"
	flowInfo, err := models.NewCiFlows().FindFlowById(namespace, flowId)
	if err != nil {
		glog.Errorf("%s get flow info failed from database:%v\n", method, err)
		return err
	}

	if flowInfo.Name == "" {
		return fmt.Errorf("not found the flow flowId:%s", flowId)
	}

	if flowInfo.NotificationConfig == "" {
		glog.Infof("%s the flow %s does not have any notification config", method, flowId)
		return fmt.Errorf("the flow %s does not have any notification config", flowId)
	}

	var notificationConfig models.NotificationConfig
	err = json.Unmarshal([]byte(flowInfo.NotificationConfig), &notificationConfig)
	if err != nil {
		glog.Errorf("%s json unmashal failed:%v\n", method, err)
		return err
	}

	if detail.Type == "ci" && detail.Result == "failed" && notificationConfig.Ci.Failed_notification {

	} else if detail.Type == "ci" && detail.Result == "success" && notificationConfig.Ci.Success_notification {

	} else if detail.Type == "cd" && detail.Result == "success" && notificationConfig.Cd.Success_notification {

	} else if detail.Type == "cd" && detail.Result == "failed" && notificationConfig.Cd.Failed_notification {

	} else {
		glog.Infof("%s the flow %s does not have any notification config", method, flowId)
		return fmt.Errorf("%s the flow %s does not have any notification config", method, flowId)
	}

	config, err := getMailConfig()
	if err != nil {
		return err
	}

	from := config.SenderMail
	user := config.SenderMail
	pwd := config.SenderPassword

	host, port, err := spiltHostPort(config.MailServer, config.Secure)
	if err != nil {
		return err
	}

	glog.Infof("%s sending notification to %s\n", method, notificationConfig.Email_list)

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", notificationConfig.Email_list...)
	m.SetHeader("Subject", fmt.Sprintf("EnnFlow(%s): %s", flowInfo.Name, detail.Subject))
	m.SetBody("text/html", detail.Body)

	d := gomail.NewDialer(host, port, user, pwd)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send notification email failed, err: %s", err)
	}
	return nil
}
