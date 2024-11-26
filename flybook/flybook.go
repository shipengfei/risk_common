package flybook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitlab.miliantech.com/infrastructure/log"
	"gitlab.miliantech.com/risk/base/risk_common/utils"
	"go.uber.org/zap"
)

type FlyBookClient struct {
	Client *http.Client
	secret string
}

func NewFlyBookClient() *FlyBookClient {
	client := &http.Client{Timeout: 6 * time.Second}
	return &FlyBookClient{Client: client}
}

var DefaultFlyBookClient = NewFlyBookClient()

func (client *FlyBookClient) SendTextMessage(ctx context.Context, url, text string) error {
	defer utils.SimpleRecover(ctx)

	timestamp := time.Now().Unix()
	bodyContent := map[string]any{
		"timestamp": fmt.Sprintf("%v", timestamp),
		"msg_type":  "text",
		"content":   map[string]any{"text": text},
	}
	if client.secret != "" {
		bodyContent["sign"], _ = client.signString(timestamp)
	}

	bs, err := json.Marshal(bodyContent)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBodyBs, errRespBody := io.ReadAll(resp.Body)
	if errRespBody != nil {
		return errRespBody
	}

	responseObj := make(map[string]any)
	if err := json.Unmarshal(respBodyBs, &responseObj); err != nil {
		return err
	}

	if msg, load := responseObj["msg"]; load {
		if msgStr, ok := msg.(string); ok && msgStr == "success" {
			return nil
		} else {
			return errors.New(msgStr)
		}
	}

	log.Info(ctx, "sendTextMessage", zap.String("go_text", string(respBodyBs)))
	return errors.New("unknown error")
}

func (client *FlyBookClient) WithSecret(secret string) *FlyBookClient {
	client.secret = secret
	return client
}

func (client *FlyBookClient) signString(timestamp int64) (sign string, err error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + client.secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
