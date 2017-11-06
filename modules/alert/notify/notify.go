/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 * 2017-03-06  @author mengyuan
 */
package notify

import (
	"errors"
	"strings"
	"sync"

	"dev-flows-api-golang/models/alert"

	"github.com/golang/glog"
	"github.com/prometheus/alertmanager/notify"
)

// Notifier is an interface to send alert
type Notifier interface {
	Notify(message notify.WebhookMessage) error
}

type Receiver struct {
	Email []string
}

func Notify(strategy alert.NotifyStrategy, message notify.WebhookMessage, r Receiver) error {
	method := "Notify"
	var notifiers []Notifier
	var errMsg = make([]string, 0, 1)
	if len(r.Email) != 0 {
		glog.V(4).Infof("%s add email notifier. email receivers: %s\n", method, r.Email)
		notifiers = append(notifiers, &EmailNotifier{Receivers: r.Email, strategy: strategy})
	}

	// notify user
	var wg sync.WaitGroup
	for i := range notifiers {
		wg.Add(1)
		go func(i int) {
			if err := notifiers[i].Notify(message); err != nil {
				errMsg = append(errMsg, err.Error())
				glog.Errorf("%s failed. error: %s\n", method, err)
			}

			wg.Done()
		}(i)
	}
	wg.Wait()

	if len(errMsg) > 0 {
		return errors.New(strings.Join(errMsg, "\n"))
	}
	return nil
}
