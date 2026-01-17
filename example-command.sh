#!/bin/bash
#
# example-command.sh
#
# -> https://github.com/awfufu/go-hurobot/blob/main/example-command.sh
#
# This is an example of how to create an external command for the bot.
#
# 1. Place this script (or a symlink to it) in the `cmds/` directory inside the bot's working directory.
#    The filename of the script/symlink will be the command name.
#    For example, if you name it `hello`, the command will be `/hello`.
#    Ensure the script is executable: `chmod +x cmds/hello`.
#
# 2. Accessing Arguments:
#    Arguments passed to the command are available as standard bash arguments ($1, $2, ...).
#
# 3. Accessing Message Context:
#    The bot injects the original message context as environment variables:
#    - QBOT_USER_ID    : Sender's QQ number
#    - QBOT_GROUP_ID   : Group ID (if in group)
#    - QBOT_MSG_ID     : Message ID
#    - QBOT_CHAT_TYPE  : Chat Type (numeric)
#    - QBOT_NAME       : Sender's Nickname
#    - QBOT_GROUP_CARD : Sender's Group Card name
#
# 4. Returning Responses:
#    Whatever you print to stdout (echo) will be sent back to the chat as a reply.

USER_ID="${QBOT_USER_ID}"
NAME="${QBOT_NAME}"
ARG_1="$1"

if [ -z "$ARG_1" ]; then
    echo "Usage: /example-command.sh <something>"
    exit 0
fi

echo "Hello, ${NAME} (${USER_ID})!"
echo "You said: ${ARG_1}"
echo "This message was processed by a script."
echo "Learn more: https://github.com/awfufu/go-hurobot/blob/main/example-command.sh"
