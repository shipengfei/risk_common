package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt"
)

const SecretKey = "e5bc1bt093f9d39df09665714be98a2be93dc514f86914d81cf678a6f294291b11eeecc8bec091ab4cf7efb1b7a0920267ddbc042958bad730d6249a3ed2b786"

func parseAuthorization(auth string) (jwt.MapClaims, error) {
	auth = strings.TrimSpace(auth)
	token, err := jwt.Parse(auth, func(token *jwt.Token) (i any, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("method %v is unexpected signing", token.Header["alg"])
		}
		return []byte(SecretKey), e
	})
	if err == nil && token != nil {
		claims, _ := token.Claims.(jwt.MapClaims)
		switch e := claims["id"].(type) {
		case float64:
			claims["id"] = int(e)
			return claims, nil
		default:
			return nil, errors.New("unknown type")
		}
	}
	return nil, err
}

func TestParseJwt(t *testing.T) {
	ms, _ := parseAuthorization("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjaGFubmVsX25hbWUiOiJtYXJrZXRfdGFvYmFvIiwiZGV2aWNlX2lkIjoiNGQ3NjlhMjg2Y2UyYmQ5YyIsImV4cGlyZV9hdCI6IjIwMjMtMDYtMDYgMDI6MDE6NDYiLCJnaW9pZCI6IiIsImlkIjoxNDc5NjA2MTgsIm5vbmNlc3RyIjoiIiwicGxhdGZvcm0iOjEsInlkaWQiOiIifQ.xnStOLFEoRqVessRMCMchiVHzJOXC45K_3cURr2C4hc")
	t.Log(ms)
}

func TestParseJwtFromFile(t *testing.T) {
	byteList, _ := os.ReadFile("/Users/shipengfei/Downloads/c77e8a3f-4def-4c0b-8e83-77acf4d90c49.json")
	fields := []string{"Channel", "CodeTag", "ApiKey", "DeviceId"}
	for _, field := range fields {
		file, _ := os.Create("/Users/shipengfei/Downloads/" + strings.ToLower(field) + ".txt")
		for _, lineStr := range strings.Split(string(byteList), "\n") {
			lineObj := make(map[string]any)
			if json.Unmarshal([]byte(lineStr), &lineObj) == nil {
				goTxt, _ := lineObj["go_text"].(string)
				goTxtObj := make(map[string]any)
				if json.Unmarshal([]byte(goTxt), &goTxtObj) == nil {
					if optVals, ok := goTxtObj["opts"].(map[string]any); ok && len(optVals) > 0 {
						authStr, _ := optVals["Authorization"].(string)
						ms, _ := parseAuthorization(authStr)
						if _, ok2 := ms["id"]; !ok2 || len(ms) == 0 {
							continue
						}

						if val, _ := optVals[field].(string); val == "" {
							file.WriteString(fmt.Sprintf("%v", ms["id"]) + "\t" + goTxtObj["uri"].(string))
							file.WriteString("\n")
						}
					}

				} else {
					t.Log(goTxtObj)
				}
			}
		}
		file.Close()
	}
}
