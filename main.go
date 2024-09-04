package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	ClientID = "mtxanqxed4emk8jsx35e"
	Secret   = "xxxxx_xxxxx"
)

func main() {
	onlyToken := flag.Bool("onlyToken", false, "./openapi-demo --onlyToken=true")
	flag.Parse()

	//除非 token 失效，否则不必要每次都获取 token, 这里只为演示用法
	token, err := GetToken(ClientID, Secret)
	if token == "" || err != nil {
		fmt.Printf("GetToken failed token:%s, err:%v", token, err)
		return
	}

	if *onlyToken {
		fmt.Printf("only GetToken onlyToken:%v, done", *onlyToken)
		return
	}

	//调用流式 API
	doAIChat(ClientID, Secret, token)
}

type chatRequest struct {
	Input string `json:"input"`
}

func doAIChat(clientId string, secret string, token string) {
	chatReq := chatRequest{
		Input: "今天天气怎么样",
	}
	chatReqs, _ := json.Marshal(chatReq)

	api := "/v1.0/cloud/iot/panel/studio/ai/chat"
	payload := strings.NewReader(string(chatReqs))
	req, _ := http.NewRequest("POST", Host+api, payload)

	buildHeader(req, clientId, secret, token, chatReqs)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("doChat http DO err:%s\n", err)
		return
	}
	defer resp.Body.Close()

	bufReader := bufio.NewReader(resp.Body)
	for {
		rawLine, readErr := bufReader.ReadBytes('\n')
		if readErr != nil {
			fmt.Printf("bufReader.ReadBytes error: %v\n", readErr)
			break
		}

		fmt.Printf("aiChatLine:\n%s\n", string(rawLine))
	}

	//bodys, _ := io.ReadAll(resp.Body)
	//fmt.Printf("StatusCode:%v, bodys:%s\n", resp.StatusCode, bodys)
}
