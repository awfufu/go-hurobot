package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"go-hurobot/qbot"
)

const llmHelpMsg string = `Configure LLM settings.
Usage:
  /llm prompt [new_prompt]       - Set or view system prompt
  /llm max-history [number]      - Set or view max history messages
  /llm enable|disable            - Enable or disable LLM
  /llm status                    - View current settings
  /llm model [model_name]        - Set or view model
  /llm supplier [supplier_name]  - Set or view API supplier`

type LlmCommand struct {
	cmdBase
}

func NewLlmCommand() *LlmCommand {
	return &LlmCommand{
		cmdBase: cmdBase{
			Name:        "llm",
			HelpMsg:     llmHelpMsg,
			Permission:  getCmdPermLevel("llm"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MinArgs:     2,
		},
	}
}

func (cmd *LlmCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *LlmCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	if len(args) < 2 {
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
		return
	}

	var llmConfig struct {
		Prompt     string
		MaxHistory int
		Enabled    bool
		Debug      bool
		Supplier   string
		Model      string
	}

	err := qbot.PsqlDB.Table("group_llm_configs").
		Where("group_id = ?", src.GroupID).
		First(&llmConfig).Error

	if err != nil {
		llmConfig = struct {
			Prompt     string
			MaxHistory int
			Enabled    bool
			Debug      bool
			Supplier   string
			Model      string
		}{
			Prompt:     "",
			MaxHistory: 200,
			Enabled:    true,
			Debug:      false,
			Supplier:   "siliconflow",
			Model:      "deepseek-ai/DeepSeek-V3",
		}
		qbot.PsqlDB.Table("group_llm_configs").Create(map[string]any{
			"group_id":    src.GroupID,
			"prompt":      llmConfig.Prompt,
			"max_history": llmConfig.MaxHistory,
			"enabled":     llmConfig.Enabled,
			"debug":       llmConfig.Debug,
			"supplier":    llmConfig.Supplier,
			"model":       llmConfig.Model,
		})
	}

	switch args[1] {
	case "prompt":
		if len(args) == 2 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("prompt: %s", llmConfig.Prompt))
		} else {
			newPrompt := strings.Join(args[2:], " ")
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", src.GroupID).
				Update("prompt", newPrompt).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, err.Error())
			} else {
				c.SendMsg(src.GroupID, src.UserID, "prompt updated")
			}
		}

	case "max-history":
		if len(args) == 2 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("max-history: %d", llmConfig.MaxHistory))
		} else {
			maxHistory, err := strconv.Atoi(args[2])
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, "Enter a valid number")
				return
			}
			if maxHistory < 0 {
				c.SendMsg(src.GroupID, src.UserID, "max-history cannot be negative")
				return
			}
			if maxHistory > 300 {
				c.SendMsg(src.GroupID, src.UserID, "max-history cannot exceed 300")
				return
			}
			err = qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", src.GroupID).
				Update("max_history", maxHistory).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, "Failed: "+err.Error())
			} else {
				c.SendMsg(src.GroupID, src.UserID, "max-history updated")
			}
		}

	case "enable":
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", src.GroupID).
			Update("enabled", true).Error
		if err != nil {
			c.SendMsg(src.GroupID, src.UserID, err.Error())
		} else {
			c.SendMsg(src.GroupID, src.UserID, "Enabled LLM")
		}

	case "disable":
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", src.GroupID).
			Update("enabled", false).Error
		if err != nil {
			c.SendMsg(src.GroupID, src.UserID, err.Error())
		} else {
			c.SendMsg(src.GroupID, src.UserID, "Disabled LLM")
		}

	case "status":
		status := fmt.Sprintf("enabled: %v\nmax-history: %d\nsupplier: %q\nmodel: %q\nprompt: %q",
			llmConfig.Enabled,
			llmConfig.MaxHistory,
			llmConfig.Supplier,
			llmConfig.Model,
			llmConfig.Prompt,
		)
		c.SendMsg(src.GroupID, src.UserID, status)

	case "tokens":
		var user qbot.Users
		if len(args) == 2 {
			err := qbot.PsqlDB.Where("user_id = ?", src.UserID).First(&user).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, "Failed to get token usage")
				return
			}
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Token usage: %d", user.TokenUsage))
		} else if len(args) == 3 && strings.HasPrefix(args[2], "--at=") {
			targetID := str2uin64(strings.TrimPrefix(args[2], "--at="))
			err := qbot.PsqlDB.Where("user_id = ?", targetID).First(&user).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, "Failed to get token usage")
				return
			}
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Token usage for %s: %d", args[2], user.TokenUsage))
		} else {
			c.SendMsg(src.GroupID, src.UserID, "Usage:\nllm tokens\nllm tokens @user")
		}

	case "debug":
		if len(args) == 2 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("debug: %v", llmConfig.Debug))
		} else {
			debugValue := strings.ToLower(args[2])
			if debugValue != "on" && debugValue != "off" {
				return
			}
			newDebug := debugValue == "on"
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", src.GroupID).
				Update("debug", newDebug).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, err.Error())
			} else {
				c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("debug = %v", newDebug))
			}
		}

	case "model":
		if len(args) == 2 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("model: %s", llmConfig.Model))
		} else {
			newModel := args[2]
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", src.GroupID).
				Update("model", newModel).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, err.Error())
			} else {
				c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("model updated to %s", newModel))
			}
		}

	case "supplier":
		if len(args) == 2 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("supplier: %s", llmConfig.Supplier))
		} else {
			newSupplier := args[2]

			var exists int64
			qbot.PsqlDB.Table("suppliers").
				Where("name = ?", newSupplier).
				Count(&exists)
			if exists == 0 {
				c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("unknown supplier: %s", newSupplier))
				return
			}

			var sup struct {
				DefaultModel string `psql:"default_model"`
			}
			qbot.PsqlDB.Table("suppliers").
				Select("default_model").
				Where("name = ?", newSupplier).
				Scan(&sup)

			// Update supplier
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", src.GroupID).
				Update("supplier", newSupplier).Error
			if err != nil {
				c.SendMsg(src.GroupID, src.UserID, err.Error())
				return
			}

			// Auto-switch model to supplier default if provided
			if strings.TrimSpace(sup.DefaultModel) != "" {
				_ = qbot.PsqlDB.Table("group_llm_configs").
					Where("group_id = ?", src.GroupID).
					Update("model", sup.DefaultModel).Error
				c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("supplier updated to %s, model -> %s", newSupplier, sup.DefaultModel))
			} else {
				c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("supplier updated to %s", newSupplier))
			}
		}

	default:
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Unrecognized parameter >>%s<<", args[1]))
	}
}
