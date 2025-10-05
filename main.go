package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	//"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type SystemPrompt struct {
	Text string `json:"text"`
}

type Content struct {
	Text string `json:"text"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// NovaRequest
type NovaRequest struct {
	System   []SystemPrompt `json:"system"`
	Messages []Message      `json:"messages"`
}

type ResponseMsg struct {
	Message []Content `json:"message"`
}
type NovaResponse struct {
	Output string `json:"output"`
}

type CommitMsg struct {
	CommitMsg string `json:"commit_message"`
}

type GitAgent struct {
	SystemMsg string
	UserMsg   string
}

func (ga GitAgent) Run() []byte {
	// Configuration
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("bshelling"))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
	}

	// System Prompt
	systemPrompt := []SystemPrompt{
		{Text: ga.SystemMsg},
	}
	// Message
	messages := []Message{
		{Role: "user", Content: []Content{{Text: ga.UserMsg}}},
	}

	// Agent Request
	request := NovaRequest{
		System:   systemPrompt,
		Messages: messages,
	}

	body, err := json.Marshal(request)
	if err != nil {
		log.Panicln("Couldn't marshal the request: ", err)
	}
	brclient := bedrockruntime.NewFromConfig(cfg)

	//parsedReq := bytes.NewBuffer(body)
	//fmt.Println(parsedReq.String())

	modelId := "amazon.nova-lite-v1:0"
	contentType := "application/json"
	modelresponse, err := brclient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &modelId,
		ContentType: &contentType,
		Body:        body,
	})
	if err != nil {
		fmt.Printf("Bedrock error: %v", err)
	}
	return modelresponse.Body

}

// Write a commit message exec.
func commitfile() {

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("")
	fmt.Print("Hi! What would you like to say? ")
	for scanner.Scan() {
		line := scanner.Text()
		if line == fmt.Sprintf("%s", line) {
			break
		}

	}

	ga := GitAgent{
    SystemMsg: "You write git an expanded commit message from the message provided. Only return the subject on one line and message on the next. Don't annotate the with `Subject:` and `Message:`",
		UserMsg:   scanner.Text(),
	}

	buffer := bytes.NewBuffer(ga.Run())
	var result map[string]interface{}
	jsonerr := json.Unmarshal([]byte(buffer.String()), &result)
	if jsonerr != nil {
		fmt.Println(jsonerr)
	}

	res, ok := result["output"].(map[string]interface{})
	if !ok {
		fmt.Println("No message available")
	}

	msg, ok := res["message"].(map[string]interface{})
	if !ok {
		fmt.Println("No message available")
	}

	content, ok := msg["content"].([]interface{})
	if !ok {
		fmt.Println("No content available")
	}

	textoutput, ok := content[0].(map[string]interface{})
	if !ok {
		fmt.Println("No content available")
	}

	commitmsg, ok := textoutput["text"].(string)
	if !ok {
		fmt.Println("No content available")
	}

	fmt.Print(commitmsg)

	cmdout, cmderr := exec.Command("git", "commit", "-m", fmt.Sprint(commitmsg)).Output()
	if cmderr != nil {
		fmt.Println("----------------------------------------------")
		fmt.Println("⚠️ Something went wrong please try again")
		fmt.Println("----------------------------------------------")
	}
	fmt.Println("")
	cmdOutput := bytes.NewBuffer(cmdout)
	fmt.Println(cmdOutput.String())

}

func main() {

	commitfile()

}
