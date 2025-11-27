package main

import (
	"fmt"
	"time"
)

type Alert struct {
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Status      string            `json:"status"`
	StartsAt    string            `json:"startsAt"`
	EndsAt      string            `json:"endsAt"`
}

type AlertData struct {
	Alerts []Alert `json:"alerts"`
	Status string  `json:"status"`
}

func formatAlertMessage(alertData AlertData) string {
	message := "🚨 *SigNoz алерт*\n"

	if len(alertData.Alerts) > 0 {
		for _, alert := range alertData.Alerts {
			title := getValue(alert.Annotations, "info", "Нет названия")
			description := getValue(alert.Annotations, "description", "Нет описания")
			severity := getValue(alert.Labels, "severity", "unknown")

			startTime, _ := time.Parse(time.RFC3339, alert.StartsAt)
			formattedStart := startTime.Format("2006-01-02 15:04:05")

			message += fmt.Sprintf("*Алерт:* %s\n", escapeStringForTelega(title))
			message += fmt.Sprintf("*Описание:* %s\n", escapeStringForTelega(description))
			message += fmt.Sprintf("*Важность:* %s\n", escapeStringForTelega(severity))
			message += fmt.Sprintf("*Статус:* %s\n", escapeStringForTelega(alert.Status))
			message += fmt.Sprintf("*Начался:* %s\n", escapeStringForTelega(formattedStart))

			if alert.Status == "resolved" {
				endTime, _ := time.Parse(time.RFC3339, alert.EndsAt)
				formattedEnd := endTime.Format("2006-01-02 15:04:05")
				message += fmt.Sprintf("*Разрешён:* %s\n", escapeStringForTelega(formattedEnd))
			}
			message += "\n"
		}
	}

	return message
}
