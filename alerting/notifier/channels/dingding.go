package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
	old_notifiers "github.com/grafana/grafana/pkg/services/alerting/notifiers"
)

const defaultDingdingMsgType = "link"

// NewDingDingNotifier is the constructor for the Dingding notifier
func NewDingDingNotifier(model *NotificationChannelConfig, t *template.Template) (*DingDingNotifier, error) {
	if model.Settings == nil {
		return nil, alerting.ValidationError{Reason: "No Settings Supplied"}
	}

	url := model.Settings.Get("url").MustString()
	if url == "" {
		return nil, alerting.ValidationError{Reason: "Could not find url property in settings"}
	}

	msgType := model.Settings.Get("msgType").MustString(defaultDingdingMsgType)

	return &DingDingNotifier{
		NotifierBase: old_notifiers.NewNotifierBase(&models.AlertNotification{
			Uid:                   model.UID,
			Name:                  model.Name,
			Type:                  model.Type,
			DisableResolveMessage: model.DisableResolveMessage,
			Settings:              model.Settings,
		}),
		MsgType: msgType,
		URL:     url,
		Message: model.Settings.Get("message").MustString(`{{ template "default.message" .}}`),
		log:     log.New("alerting.notifier.dingding"),
		tmpl:    t,
	}, nil
}

// DingDingNotifier is responsible for sending alert notifications to ding ding.
type DingDingNotifier struct {
	old_notifiers.NotifierBase
	MsgType string
	URL     string
	Message string
	tmpl    *template.Template
	log     log.Logger
}

// Notify sends the alert notification to dingding.
func (dd *DingDingNotifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	dd.log.Info("Sending dingding")

	ruleURL, err := joinUrlPath(dd.tmpl.ExternalURL.String(), "/alerting/list")
	if err != nil {
		return false, err
	}

	q := url.Values{
		"pc_slide": {"false"},
		"url":      {ruleURL},
	}

	// Use special link to auto open the message url outside of Dingding
	// Refer: https://open-doc.dingtalk.com/docs/doc.htm?treeId=385&articleId=104972&docType=1#s9
	messageURL := "dingtalk://dingtalkclient/page/link?" + q.Encode()

	var tmplErr error
	tmpl, _, err := TmplText(ctx, dd.tmpl, as, dd.log, &tmplErr)
	if err != nil {
		return false, err
	}

	message := tmpl(dd.Message)
	title := tmpl(`{{ template "default.title" . }}`)

	var bodyMsg map[string]interface{}
	if dd.MsgType == "actionCard" {
		bodyMsg = map[string]interface{}{
			"msgtype": "actionCard",
			"actionCard": map[string]string{
				"text":        message,
				"title":       title,
				"singleTitle": "More",
				"singleURL":   messageURL,
			},
		}
	} else {
		link := map[string]string{
			"text":       message,
			"title":      title,
			"messageUrl": messageURL,
		}

		bodyMsg = map[string]interface{}{
			"msgtype": "link",
			"link":    link,
		}
	}

	if tmplErr != nil {
		return false, fmt.Errorf("failed to template DingDing message: %w", tmplErr)
	}

	body, err := json.Marshal(bodyMsg)
	if err != nil {
		return false, err
	}

	cmd := &models.SendWebhookSync{
		Url:  dd.URL,
		Body: string(body),
	}

	if err := bus.DispatchCtx(ctx, cmd); err != nil {
		return false, fmt.Errorf("send notification to dingding: %w", err)
	}

	return true, nil
}

func (dd *DingDingNotifier) SendResolved() bool {
	return !dd.GetDisableResolveMessage()
}