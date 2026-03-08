package builtin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DaWesen/Pandora/pkg/core"
)

type TimeTool struct{}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

func (t *TimeTool) Name() string {
	return "time"
}

func (t *TimeTool) Description() string {
	return "Provides the current time and date information."
}

func (t *TimeTool) Schema() core.ToolSchema {
	return core.ToolSchema{
		Type: "object",
		Properties: map[string]core.Property{
			"format": {
				Type:        "string",
				Description: "时间格式：'full'(完整), 'date'(仅日期), 'time'(仅时间), 'timestamp'(时间戳)",
				Enum:        []string{"full", "date", "time", "timestamp"},
			},
			"timezone": {
				Type:        "string",
				Description: "时区，如 'Asia/Shanghai', 'UTC'。默认本地时区",
			},
		},
		Required: []string{"format"},
	}
}

func (t *TimeTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	var params struct {
		Format   string `json:"format"`
		Timezone string `json:"timezone"`
	}

	if err := json.Unmarshal(input.Arguments, &params); err != nil {
		return core.ToolOutput{}, err
	}

	var now time.Time

	if params.Timezone != "" {
		loc, err := time.LoadLocation(params.Timezone)
		if err != nil {
			return core.ToolOutput{}, fmt.Errorf("invalid timezone: %v", err)
		}
		now = time.Now().In(loc)
	} else {
		now = time.Now()
	}

	var formattedTime string
	switch params.Format {
	case "full":
		formattedTime = now.Format("2006-01-02 15:04:05 MST")
	case "date":
		formattedTime = now.Format("2006-01-02")
	case "time":
		formattedTime = now.Format("15:04:05")
	case "timestamp":
		formattedTime = fmt.Sprintf("%d", now.Unix())
	default:
		formattedTime = now.Format(time.RFC3339)
	}

	return core.ToolOutput{
		Content: formattedTime,
		Data: map[string]any{
			"timestamp": now.Unix(),
			"formatted": formattedTime,
			"format":    params.Format,
			"timezone":  params.Timezone,
		},
	}, nil
}
