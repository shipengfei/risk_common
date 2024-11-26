package test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	member_status "gitlab.miliantech.com/go/client-go/proto/member_status_server"
	"gitlab.miliantech.com/go/client-go/proto/risk_manager"
	"gitlab.miliantech.com/go/client-go/proto/userInfo"
	"gitlab.miliantech.com/go/common/util/member_crypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var grpcConnClient *grpc.ClientConn
var err error

func init() {
	grpcConnClient, err = grpc.Dial("test1-internal.miliantech.com:80", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
}

func TestUserInfo(t *testing.T) {
	client := userInfo.NewUserServerClient(grpcConnClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	resp, err := client.GetUserInfo(ctx, &userInfo.GetUserInfoRq{
		UserId: 5016909, NeedMoreStatus: 1,
	})
	t.Log(resp.GetUserInfo(), err)
}

func TestMemberStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	statusClient := member_status.NewMemberStatusClient(grpcConnClient)
	respStatus, errSta := statusClient.GetMemberStatus(ctx, &member_status.GetMemberStatusReq{MemberId: 5016909})
	t.Log(respStatus, errSta)
}

func TestTag(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	//member_id:5003920 tag_id:6 source_type:"sex_check_histories" source_id:12728 admin_email:"wangcaiting@miliantech.com" admin_id:1428 comment:"https://test1-sh.miliantech.com/admin/live_detail?id=12728" scene_type:7 sex:1 role:"teach_cupid" wealth:1001.05
	client := risk_manager.NewRiskManagerClient(grpcConnClient)

	req := &risk_manager.AddViolationTagRecordReq{
		MemberId:   5003920,
		TagId:      6,
		SourceType: "sex_check_histories",
		SourceId:   12728,
		AdminEmail: "wangcaiting@miliantech.com",
		AdminId:    1428,
		Comment:    "https://test1-sh.miliantech.com/admin/live_detail?id=12728",
		SceneType:  7,
		Sex:        1,
		Role:       "teach_cupid",
		Wealth:     1001.05,
	}
	resp, err := client.AddViolationTagRecord(ctx, req)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(resp)
	}
}

func TestFamilies(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	client := risk_manager.NewDecisionClient(grpcConnClient)

	header := make(http.Header)
	header.Set("umid", "0c89f3ad2804795256f9516048acf8ea")
	header.Set("Noncestr", "33f183fe4e0a4e02b11e127da98f90931")
	bs, _ := json.Marshal(header)

	req := &risk_manager.FamiliesRiskCheckReq{
		MemberId:    5016909,
		Headers:     bs,
		Version:     1,
		SourceType:  "Box",
		SourceCount: 1,
	}

	if true {
		ctx = metadata.AppendToOutgoingContext(ctx, "recordId", "4")
	}

	resp, err := client.FamiliesRiskCheck(ctx, req)

	if err != nil {
		t.Error(err)
	}

	t.Log(resp)
}

func TestNewYear2023(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	client := risk_manager.NewDecisionClient(grpcConnClient)
	resp, err := client.MemberIncrNewYear2023(ctx, &risk_manager.MemberIncrNewYear2023Req{MemberId: 5018000, ActivedAt: 1673335513})
	t.Log(resp, err)
}

type CustomHandler func()

func TestHandler(t *testing.T) {
	content, err := ioutil.ReadFile("/Users/shipengfei/Desktop/member_uid.txt")
	if err != nil {
		return
	}

	ids := make([]string, 0)
	for _, idStr := range strings.Split(string(content), "\n") {
		id, _ := member_crypt.Decrypt(idStr)
		if id > 0 {
			ids = append(ids, strconv.Itoa(id))
		}
	}

	ioutil.WriteFile("xxx.txt", []byte(strings.Join(ids, ",")), 0777)
}

func TestVoteAuth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	client := risk_manager.NewDecisionClient(grpcConnClient)
	resp, err := client.RiskCheck(ctx, &risk_manager.VoteAuthReq{
		MemberId:    109586,
		DeviceToken: "McZjjcQvocDZStlERcZZo_QOytAQy1gnRRJRyRS_RK_5j7SEoRMQ55eFRcSjj7yRjcSRjkxZR5JO5cyWjcSv5c_5yZDRjcMQ57o5oZRGj1ZWo7JQjtoNV5tsVQFsQOgsn_7SF_lZgK7JV5kQVFRoo1MZJWgPQ_osWHFnZQS3n7oVW1g4WKyRyRJRoZB4yRD_o1ZooRLGjcLjoZQ3oKR_o7yQy_oFjtxWF7xiyFo1sfD3SfYjRko1sfD3SfYjRko4Z5xnyK_qWHyQyKeoVKg4y5+ZoFRRoRRGQRDvQRMZRkRvQRM0",
		Scene:       "CheckDeviceVote",
		VoteCount:   10,
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func TestXxx(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{PoolSize: 2})

	size := 100000
	var wg sync.WaitGroup
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			doRedis(t, redisClient)
			wg.Done()
		}()
	}
	wg.Wait()
}

func doRedis(t *testing.T, client *redis.Client) {
	ctx := context.Background()
	_, err := client.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.HSet(ctx, "abc", "key", "value")
		p.HIncrBy(ctx, "abc", "count", 1)
		p.Expire(ctx, "abc", time.Minute)
		return nil
	})
	if err != nil {
		t.Log(err)
	}
}
