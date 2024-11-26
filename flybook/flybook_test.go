package flybook

import (
	"context"
	"testing"
)

func TestXxx(t *testing.T) {
	client := NewFlyBookClient() //.WithSecret("YNftocWuNrEst8sIbfPUwf")
	err := client.SendTextMessage(context.Background(), "https://open.feishu.cn/open-apis/bot/v2/hook/5b3b9962-6ee1-4ccb-ac11-1fc3efe5ef1f", "hello")
	t.Log(err)
}
