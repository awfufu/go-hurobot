package qbot

import "log"

func (c *Client) SendPrivateMsg(userID uint64, message string, autoEscape bool) (uint64, error) {
	if message == "" {
		message = " "
	}
	req := cqRequest{
		Action: "send_private_msg",
		Params: map[string]any{
			"user_id":     userID,
			"message":     message,
			"auto_escape": autoEscape,
		},
	}
	resp, err := c.sendWithResponse(&req)
	if err != nil {
		return 0, err
	}
	log.Println("send-private: ", message)
	return resp.Data.MessageId, nil
}

func (c *Client) SendGroupMsg(groupID uint64, message string, autoEscape bool) (uint64, error) {
	if message == "" {
		message = " "
	}
	req := cqRequest{
		Action: "send_group_msg",
		Params: map[string]any{
			"group_id":    groupID,
			"message":     message,
			"auto_escape": autoEscape,
		},
	}

	resp, err := c.sendWithResponse(&req)
	if err != nil {
		return 0, err
	}
	log.Println("send-group: ", message)
	return resp.Data.MessageId, nil
}

func (c *Client) SetGroupSpecialTitle(groupID uint64, userID uint64, specialTitle string) error {
	req := cqRequest{
		Action: "set_group_special_title",
		Params: map[string]any{
			"group_id":      groupID,
			"user_id":       userID,
			"special_title": specialTitle,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupName(groupID uint64, groupName string) error {
	req := cqRequest{
		Action: "set_group_name",
		Params: map[string]any{
			"group_id":   groupID,
			"group_name": groupName,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupAdmin(groupID uint64, userID uint64, enable bool) error {
	req := cqRequest{
		Action: "set_group_admin",
		Params: map[string]any{
			"group_id": groupID,
			"user_id":  userID,
			"enable":   enable,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupBan(groupID uint64, userID uint64, duration int) error {
	req := cqRequest{
		Action: "set_group_ban",
		Params: map[string]any{
			"group_id": groupID,
			"user_id":  userID,
			"duration": duration,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupEssence(msgID uint64) error {
	req := cqRequest{
		Action: "set_essence_msg",
		Params: map[string]any{
			"message_id": msgID,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) DeleteGroupEssence(msgID uint64) error {
	req := cqRequest{
		Action: "delete_essence_msg",
		Params: map[string]any{
			"message_id": msgID,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) DeleteMsg(msgID uint64) error {
	req := cqRequest{
		Action: "delete_msg",
		Params: map[string]any{
			"message_id": msgID,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SendMsg(groupID uint64, userID uint64, message string) {
	if groupID == 0 {
		c.SendPrivateMsg(userID, message, false)
	} else {
		c.SendGroupMsg(groupID, message, false)
	}
}

func (c *Client) GetGroupMemberInfo(groupID uint64, userID uint64, noCache bool) (*GroupMemberInfo, error) {
	req := cqRequest{
		Action: "get_group_member_info",
		Params: map[string]any{
			"group_id": groupID,
			"user_id":  userID,
			"no_cache": noCache,
		},
	}
	resp, err := c.sendWithResponse(&req)
	if err != nil {
		return nil, err
	}
	return &resp.Data.GroupMemberInfo, nil
}

func (c *Client) GetGroupFileUrl(groupID uint64, fileID string, busid int32) (string, error) {
	req := cqRequest{
		Action: "get_group_file_url",
		Params: map[string]any{
			"group_id": groupID,
			"file_id":  fileID,
			"busid":    busid,
		},
	}
	resp, err := c.sendWithResponse(&req)
	if err != nil {
		return "", err
	}
	return resp.Data.Url, nil
}

func (c *Client) SendTestAPIRequest(action string, params map[string]interface{}) (string, error) {
	req := cqRequest{
		Action: action,
		Params: params,
	}

	resp, err := c.sendWithJSONResponse(&req)
	if err != nil {
		return "", err
	}

	return resp, nil
}
