#!/bin/bash
#
# example-command-cn.sh
#
# -> https://github.com/awfufu/go-hurobot/blob/main/example-command-cn.sh
#
# 这是一个关于如何为机器人创建外部命令的示例。
#
# 1. 将此脚本（或其软链接）放置在机器人工作目录下的 `cmds/` 目录中。
#    脚本或软链接的文件名将成为命令名。
#    例如，如果你将其命名为 `hello`，则命令将为 `/hello`。
#    确保脚本具有可执行权限：`chmod +x cmds/hello`。
#
# 2. 获取参数：
#    传递给命令的参数可以通过标准的 bash 参数（$1, $2, ...）获取。
#
# 3. 获取消息上下文：
#    机器人会将原始消息的上下文注入为环境变量：
#    - QBOT_USER_ID    : 发送者的 QQ 号
#    - QBOT_GROUP_ID   : 群组 ID（如果是群消息）
#    - QBOT_MSG_ID     : 消息 ID
#    - QBOT_CHAT_TYPE  : 聊天类型（数值）
#    - QBOT_NAME       : 发送者昵称
#    - QBOT_GROUP_CARD : 发送者群名片
#
# 4. 返回响应：
#    任何打印到标准输出（echo）的内容都会作为回复发送回聊天。

USER_ID="${QBOT_USER_ID}"
NAME="${QBOT_NAME}"
ARG_1="$1"

if [ -z "$ARG_1" ]; then
    echo "用法: /example-command-cn.sh <内容>"
    exit 0
fi

echo "你好, ${NAME} (${USER_ID})!"
echo "你说了: ${ARG_1}"
echo "这条消息是由外部脚本处理的。"
echo "了解更多: https://github.com/awfufu/go-hurobot/blob/main/example-command-cn.sh"
