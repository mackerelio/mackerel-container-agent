#!/usr/bin/env bash

if [ "$MACKEREL_AGENT_PLUGIN_META" = 1 ]; then
  cat << EOF
# mackerel-agent-plugin
{
  "graphs": {
    "dice": {
      "label": "My Dice",
      "unit": "integer",
      "metrics": [
        {
          "label": "Die 6",
          "name": "d6"
        },
        {
          "label": "Die 20",
          "name": "d20"
        }
      ]
    }
  }
}
EOF
  exit 0
fi

printf 'dice.d6\t%s\t%s\n' $((RANDOM % 6 + 1)) "$(date +%s)"
printf 'dice.d20\t%s\t%s\n' $((RANDOM % 20 + 1)) "$(date +%s)"
