package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type LoggingRT struct {
}

func (l *LoggingRT) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		reqBodyBytes = []byte{}
	}

	req.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		slog.Error(
			"error sending request",
			slog.String("url", req.URL.String()),
			slog.String("method", req.Method),
			slog.String("request_body", string(reqBodyBytes)),
			slog.String("request_headers", fmt.Sprintf("%+v", req.Header)),
			slog.String(errorLogKey, err.Error()),
		)
		return nil, err
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		respBodyBytes = []byte{}
	}

	resp.Body = io.NopCloser(bytes.NewReader(respBodyBytes))

	slog.Info(
		"got response",
		slog.String("url", req.URL.String()),
		slog.String("method", req.Method),
		slog.String("response_status", resp.Status),
		slog.String("request_body", string(reqBodyBytes)),
		slog.String("response_body", string(respBodyBytes)),
		slog.String("response_headers", fmt.Sprintf("%+v", resp.Header)),
		slog.String("request_headers", fmt.Sprintf("%+v", req.Header)),
	)

	return resp, nil
}

const errorLogKey = "error"

type tgSendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Message   string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func main() {
	lrt := &LoggingRT{}

	client := &http.Client{
		Transport: lrt,
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	tgToken, foundToken := os.LookupEnv("TOKEN")
	if !foundToken {
		slog.Error("not found telegram token")
	}

	chatID, foundToken := os.LookupEnv("CHAT_ID")
	if !foundToken {
		slog.Error("not found chat id")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/alert", func(w http.ResponseWriter, originalReq *http.Request) {
		body, err := io.ReadAll(originalReq.Body)
		if err != nil {
			http.Error(w, "Error reading originalReq body", http.StatusBadRequest)
			return
		}
		defer originalReq.Body.Close()

		// Парсим JSON с алертами
		var alertData AlertData
		if err := json.Unmarshal(body, &alertData); err != nil {
			slog.Error("failed to parse alert json",
				slog.String(errorLogKey, err.Error()),
				slog.String("body", string(body)))
			http.Error(w, "Error parsing JSON", http.StatusBadRequest)
			return
		}

		message := formatAlertMessage(alertData)

		msg := tgSendMessageRequest{
			ChatID:    chatID,
			Message:   message,
			ParseMode: "Markdown",
		}

		targetReqBodyBytes, err := json.Marshal(msg)
		if err != nil {
			slog.Error("failed to marshal tg send",
				slog.String(errorLogKey, err.Error()))
			return
		}

		_, err = client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tgToken), "application/json", bytes.NewReader(targetReqBodyBytes))
		if err != nil {
			slog.Error("failed to make proxied req", slog.String(errorLogKey, err.Error()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	err := http.ListenAndServe("0.0.0.0:8081", mux)
	if err != nil {
		slog.Error("error listening", slog.String(errorLogKey, err.Error()))
	}
}

func escapeStringForTelega(s string) string {
	replaceMap := map[string]string{
		"_": "\\_",
		"*": "\\*",
		"[": "\\[",
		"]": "\\]",
		"~": "\\~",
		"`": "\\`",
		">": "\\>",
		"#": "\\#",
		"+": "\\+",
		// "-": "\\-",
		"=": "\\=",
		"|": "\\|",
		"{": "\\{",
		"}": "\\}",
		".": "\\.",
		"!": "\\!",
	}
	for k, v := range replaceMap {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

func getValue(m map[string]string, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return defaultValue
}
