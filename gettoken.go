package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	Host = "https://openapi-cn.wgine.com"
)

type TokenResponse struct {
	Result struct {
		AccessToken  string `json:"access_token"`
		ExpireTime   int    `json:"expire_time"`
		RefreshToken string `json:"refresh_token"`
		UID          string `json:"uid"`
	} `json:"result"`
	Success bool  `json:"success"`
	T       int64 `json:"t"`
}

func GetToken(clientId string, secret string) (string, error) {
	body := []byte(``)
	req, _ := http.NewRequest("GET", Host+"/v1.0/token?grant_type=1", bytes.NewReader(body))

	buildHeader(req, clientId, secret, "", body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GetToken http DO err:%s\n", err)
		return "", err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("GetToken http DO err:%s\n", err)
		return "", err
	}

	ret := TokenResponse{}
	json.Unmarshal(bs, &ret)
	fmt.Printf("GetToken resp:%s\n", string(bs))

	if v := ret.Result.AccessToken; v != "" {
		fmt.Printf("Return token:%s\n\n", v)
		return v, nil
	}
	return "", fmt.Errorf("accessToken is nil")
}

func buildHeader(req *http.Request, clientId string, secret string, token string, body []byte) {
	req.Header.Set("client_id", clientId)
	req.Header.Set("sign_method", "HMAC-SHA256")

	ts := fmt.Sprint(time.Now().UnixNano() / 1e6)
	req.Header.Set("t", ts)

	if token != "" {
		req.Header.Set("access_token", token)
	}

	sign := buildSign(req, clientId, secret, token, body, ts)
	req.Header.Set("sign", sign)
}

func buildSign(req *http.Request, clientId string, secret string, token string, body []byte, t string) string {
	headers := getHeaderStr(req)
	urlStr := getUrlStr(req)
	contentSha256 := Sha256(body)
	stringToSign := req.Method + "\n" + contentSha256 + "\n" + headers + "\n" + urlStr
	signStr := clientId + token + t + stringToSign
	sign := strings.ToUpper(HmacSha256(signStr, secret))
	return sign
}

func Sha256(data []byte) string {
	sha256Contain := sha256.New()
	sha256Contain.Write(data)
	return hex.EncodeToString(sha256Contain.Sum(nil))
}

func getUrlStr(req *http.Request) string {
	url := req.URL.Path
	keys := make([]string, 0, 10)

	query := req.URL.Query()
	for key, _ := range query {
		keys = append(keys, key)
	}
	if len(keys) > 0 {
		url += "?"
		sort.Strings(keys)
		for _, keyName := range keys {
			value := query.Get(keyName)
			url += keyName + "=" + value + "&"
		}
	}

	if url[len(url)-1] == '&' {
		url = url[:len(url)-1]
	}
	return url
}

func getHeaderStr(req *http.Request) string {
	signHeaderKeys := req.Header.Get("Signature-Headers")
	if signHeaderKeys == "" {
		return ""
	}
	keys := strings.Split(signHeaderKeys, ":")
	headers := ""
	for _, key := range keys {
		headers += key + ":" + req.Header.Get(key) + "\n"
	}
	return headers
}

func HmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
