#!/usr/bin/env bash

NUM=${NUM:-6}

if [ "$MACKEREL_AGENT_PLUGIN_META" = 1 ]; then
  cat << EOF
# mackerel-agent-plugin
{
  "graphs": {
    "dice": {
      "label": "My Dice $NUM",
      "unit": "integer",
      "metrics": [
        {
          "label": "Die $NUM",
          "name": "d$NUM"
        }
      ]
    }
  }
}
EOF
  exit 0
fi

printf 'dice.d%s\t%s\t%s\n' "$NUM" $((RANDOM % NUM + 1)) "$(date +%s)"
