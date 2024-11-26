package test

import (
	"context"
	_ "embed"
	"log"
	"testing"
	"time"

	gBuilder "github.com/bilibili/gengine/builder"
	gContext "github.com/bilibili/gengine/context"
	gEngine "github.com/bilibili/gengine/engine"
	"gitlab.miliantech.com/go/client-go/proto/userInfo"
)

//go:embed rule.txt
var rule string

func TestDemo(t *testing.T) {
	dataContext := gContext.NewDataContext()

	// 添加基础数据
	ctx := context.Background()
	uInfo := getUserInfo(t, context.Background(), 5017744)
	dataContext.Add("ctx", ctx)
	dataContext.Add("uInfo", uInfo)
	dataContext.Add("log", log.Println)

	ruleBuilder := gBuilder.NewRuleBuilder(dataContext)
	err := ruleBuilder.BuildRuleFromString(rule)
	if err != nil {
		t.Error(err)
	}

	selectedRules := []string{"demo1"}
	// 创建引擎
	eng := gEngine.NewGengine()
	sTag := &gEngine.Stag{StopTag: true}
	err = eng.ExecuteSelectedRulesWithControlAndStopTagAsGivenSortedName(
		ruleBuilder,
		true, // 当规则执行发生错误时, 是否继续执行后续的规则
		sTag,
		selectedRules, // 选择执行的规则
	)
	if err != nil {
		t.Fatal(err)
	}
	result, _ := eng.GetRulesResultMap()
	for _, name := range selectedRules {
		t.Log(name, "===>", result[name])
	}

}

func getUserInfo(t *testing.T, ctx context.Context, memId int64) *userInfo.UserInfo {
	client := userInfo.NewUserServerClient(grpcConnClient)
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	resp, err := client.GetUserInfo(ctx, &userInfo.GetUserInfoRq{
		UserId: memId, NeedMoreStatus: 0,
	})
	if err != nil {
		t.Error(err)
	}
	return resp.GetUserInfo()
}

func TestDemo1(t *testing.T) {

}
