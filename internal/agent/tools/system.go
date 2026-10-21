package tools

import (
	"time"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

type SystemToolkit struct {
}

func NewSystemToolkit() *SystemToolkit {
	return &SystemToolkit{}
}

func (t *SystemToolkit) Tools() ([]tool.Tool, error) {
	currentDate, err := functiontool.New(functiontool.Config{
		Name:        "get_current_date",
		Description: "Текущая дата в формате YYYY-MM-DD.",
	}, func(ctx adkagent.Context, _ struct{}) (string, error) {
		return time.Now().Format(time.DateOnly), nil
	})
	if err != nil {
		return nil, err
	}
	return []tool.Tool{currentDate}, nil
}
