package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

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

func cmd_draw(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if args.Size < 2 {
		helpMsg := `Usage: draw <prompt> [--size <1328x1328|1584x1056|1140x1472|1664x928|928x1664>]`
		c.SendMsg(msg, helpMsg)
		return
	}

	if config.ApiKey == "" {
		c.SendMsg(msg, "No API key")
		return
	}

	prompt, imageSize, err := parseDrawArgs(args.Contents[1:])
	if err != nil {
		c.SendMsg(msg, err.Error())
		return
	}

	if prompt == "" {
		c.SendMsg(msg, "Please provide a prompt")
		return
	}

	c.SendMsg(msg, "Image generating...")

	reqData := ImageGenerationRequest{
		Model:         "Qwen/Qwen-Image",
		Prompt:        prompt,
		ImageSize:     imageSize,
		BatchSize:     1,
		GuidanceScale: 7.5,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	req, err := http.NewRequest("POST", "https://api.siliconflow.cn/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	if resp.StatusCode != 200 {
		c.SendMsg(msg, fmt.Sprintf("%d\n%s", resp.StatusCode, string(body)))
		return
	}

	var imgResp ImageGenerationResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	if len(imgResp.Images) == 0 {
		c.SendMsg(msg, "error: 未生成任何图片")
		return
	}

	imageURL := imgResp.Images[0].URL
	c.SendMsg(msg, qbot.CQReply(msg.MsgID)+qbot.CQImage(imageURL))
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
					return "", "", fmt.Errorf("不支持的图片尺寸: %s\n支持的尺寸: 1328x1328, 1584x1056, 1140x1472, 1664x928, 928x1664", size)
				}
				imageSize = size
				i += 2
			} else {
				return "", "", fmt.Errorf("--size: 需要指定尺寸值")
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
