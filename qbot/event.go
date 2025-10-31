package qbot

import (
	"encoding/json"
	"fmt"
	"log"
)

func parseContent(msgarr *[]MsgItem) (result string) {
	result = ""
	for _, item := range *msgarr {
		switch item.Type {
		case Text:
			result += item.Content
		case At:
			result += "[CQ:at,qq=" + item.Content + "]"
		case Face:
			faceName := GetQFaceNameByStrID(item.Content)
			log.Println(faceName)
			result += "[CQ:face,id=" + item.Content + "](" + faceName + ")"
		case Image:
			result += "[图片(无法查看)]"
		case Record:
			result = "[语音(无法查看)]"
		case Reply:
			result = "[CQ:reply,id=" + item.Content + "]"
		case File:
			result = "[文件(无法查看)]"
		case Forward:
			result = "[合并转发(无法查看)]"
		case Json:
			result = "[不支持的消息(无法查看)]"
		}
	}
	return
}

func parseMsgJson(raw *messageJson) *Message {
	if raw == nil {
		return nil
	}
	result := Message{
		GroupID:  raw.GroupID,
		UserID:   raw.Sender.UserID,
		Nickname: raw.Sender.Nickname,
		Card:     raw.Sender.Card,
		Role:     raw.Sender.Role,
		Time:     raw.Time,
		MsgID:    raw.MessageID,
		Raw:      raw.RawMessage,
	}
	for _, msg := range raw.Message {
		var jsonData map[string]any
		if err := json.Unmarshal(msg.Data, &jsonData); err != nil {
			log.Printf("解析消息数据失败: %v, 原始数据: %s", err, string(msg.Data))
			continue
		}
		switch msg.Type {
		case "text":
			if text, ok := jsonData["text"].(string); ok {
				result.Array = append(result.Array, MsgItem{
					Type:    Text,
					Content: text,
				})
			}
		case "at":
			// qq 可能是 string 或 number
			var qqStr string
			if qq, ok := jsonData["qq"].(string); ok {
				qqStr = qq
			} else if qq, ok := jsonData["qq"].(float64); ok {
				qqStr = fmt.Sprintf("%.0f", qq)
			}
			if qqStr != "" {
				result.Array = append(result.Array, MsgItem{
					Type:    At,
					Content: qqStr,
				})
			}
		case "face":
			// id 可能是 string 或 number
			var idStr string
			if id, ok := jsonData["id"].(string); ok {
				idStr = id
			} else if id, ok := jsonData["id"].(float64); ok {
				idStr = fmt.Sprintf("%.0f", id)
			}
			if idStr != "" {
				result.Array = append(result.Array, MsgItem{
					Type:    Face,
					Content: idStr,
				})
			}
		case "image":
			if url, ok := jsonData["url"].(string); ok {
				result.Array = append(result.Array, MsgItem{
					Type:    Image,
					Content: url,
				})
			}
		case "record":
			if path, ok := jsonData["path"].(string); ok {
				result.Array = append(result.Array, MsgItem{
					Type:    Record,
					Content: path,
				})
			}
		case "reply":
			// reply 的 id 可能是 string 或 number
			var replyId string
			if id, ok := jsonData["id"].(string); ok {
				replyId = id
			} else if id, ok := jsonData["id"].(float64); ok {
				replyId = fmt.Sprintf("%.0f", id)
			}
			if replyId != "" {
				result.Array = append(result.Array, MsgItem{
					Type:    Reply,
					Content: replyId,
				})
			}
		case "file":
			result.Array = append(result.Array, MsgItem{
				Type:    File,
				Content: string(msg.Data),
			})
		case "forward":
			result.Array = append(result.Array, MsgItem{
				Type:    Forward,
				Content: string(msg.Data),
			})
		case "json":
			result.Array = append(result.Array, MsgItem{
				Type:    Json,
				Content: string(msg.Data),
			})
		default:
			result.Array = append(result.Array, MsgItem{
				Type:    Other,
				Content: string(msg.Data),
			})
		}
	}
	result.Content = parseContent(&result.Array)
	log.Println(result.Content)
	return &result
}

func (c *Client) handleEvents(postType *string, msgStr *[]byte, jsonMap *map[string]any) {
	switch *postType {
	case "meta_event":
		// heartbeat, connection state..
	case "notice":
		// TODO
		switch (*jsonMap)["notice_type"] {
		case "group_recall":
			// TODO
		}
	case "message":
		switch (*jsonMap)["message_type"] {
		case "private":
			fallthrough
		case "group":
			if c.eventHandlers.onMessage != nil {
				msgJson := &messageJson{}
				if json.Unmarshal(*msgStr, msgJson) != nil {
					return
				}
				if msg := parseMsgJson(msgJson); msg != nil {
					c.eventHandlers.onMessage(c, msg)
				}
			}
		}
	}
}

func (c *Client) HandleMessage(handler func(c *Client, msg *Message)) {
	c.eventHandlers.onMessage = handler
}
