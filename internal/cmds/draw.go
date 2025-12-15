package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/qbot"
)

const drawHelpMsg string = `Generate images from text prompts.
Usage: /draw <prompt> [--size <size>]
Supported sizes: 1328x1328, 1584x1056, 1140x1472, 1664x928, 928x1664
Example: /draw a cat --size 1328x1328`

type ImageGenerationRequest struct {
	Model         string  `json:"model"`
	Prompt        string  `json:"prompt"`
	ImageSize     string  `json:"image_size"`
	BatchSize     int     `json:"batch_size"`
	GuidanceScale float64 `json:"guidance_scale"`
}

type ImageGenerationResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	Timings struct {
		Inference float64 `json:"inference"`
	} `json:"timings"`
	Seed int64 `json:"seed"`
}

type DrawCommand struct {
	cmdBase
}

func NewDrawCommand() *DrawCommand {
	return &DrawCommand{
		cmdBase: cmdBase{
			Name:       "draw",
			HelpMsg:    drawHelpMsg,
			Permission: getCmdPermLevel("draw"),

			NeedRawMsg: false,
			MinArgs:    2,
		},
	}
}

func (cmd *DrawCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *DrawCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	if config.Cfg.ApiKeys.DrawApiKey == "" {
		b.SendGroupMsg(msg.GroupID, "No API key")
		return
	}

	// msg.Array[0] is "/draw"
	// Parse following args
	var args []string
	for i := 1; i < len(msg.Array); i++ {
		if txt := msg.Array[i].GetTextItem(); txt != nil {
			args = append(args, txt.Content)
		}
	}

	prompt, imageSize, err := parseDrawArgs(args)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}

	if prompt == "" {
		b.SendGroupMsg(msg.GroupID, "Please provide a prompt")
		return
	}

	b.SendGroupMsg(msg.GroupID, "Image generating...")

	reqData := ImageGenerationRequest{
		Model:         "Qwen/Qwen-Image",
		Prompt:        prompt,
		ImageSize:     imageSize,
		BatchSize:     1,
		GuidanceScale: 7.5,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	req, err := http.NewRequest("POST", config.Cfg.ApiKeys.DrawUrlBase, bytes.NewBuffer(jsonData))
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.Cfg.ApiKeys.DrawApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	if resp.StatusCode != 200 {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%d\n%s", resp.StatusCode, string(body)))
		return
	}

	var imgResp ImageGenerationResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	if len(imgResp.Images) == 0 {
		b.SendGroupMsg(msg.GroupID, "error: no images generated")
		return
	}

	imageURL := imgResp.Images[0].URL
	// Use b.SendGroupReplyMsg instead of manual qbot.Reply/SendMsg
	b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, qbot.Image(imageURL))
}

func parseDrawArgs(args []string) (prompt, imageSize string, err error) {
	imageSize = "1328x1328" // default

	var promptParts []string
	i := 0

	for i < len(args) {
		arg := args[i]

		switch arg {
		case "--size":
			if i+1 < len(args) {
				size := args[i+1]
				if !isValidSize(size) {
					return "", "", fmt.Errorf("unsupported image size: %s\nSupported sizes: 1328x1328, 1584x1056, 1140x1472, 1664x928, 928x1664", size)
				}
				imageSize = size
				i += 2
			} else {
				return "", "", fmt.Errorf("--size: size value required")
			}
		default:
			promptParts = append(promptParts, arg)
			i++
		}
	}

	prompt = strings.Join(promptParts, " ")
	return prompt, imageSize, nil
}

func isValidSize(size string) bool {
	validSizes := []string{"1328x1328", "1584x1056", "1140x1472", "1664x928", "928x1664"}
	for _, valid := range validSizes {
		if size == valid {
			return true
		}
	}
	return false
}
