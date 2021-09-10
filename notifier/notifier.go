package notifier

import (
	"fmt"
	"strings"
)

func wellSize(size int64) string {

	if size <= 1024 {
		return fmt.Sprintf("%dByte", size)
	}
	if size <= 1024*1024 {
		return fmt.Sprintf("%.2fKB", float32(size)/1024)
	}

	if size <= 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float32(size)/(1024*1024))
	}

	if size <= 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fGB", float32(size)/(1024*1024*1024))
	}

	return fmt.Sprintf("%.2fTB", float32(size)/(1024*1024*1024*1024))

}

func getTips(used int64, total int64) string {

	percent := float64(used) / float64(total)
	if percent <= 60 {
		return "流量充足"
	}
	if percent <= 80 {
		return "流量告警"
	}

	return "即将用完"
}

func getColor(used int64, total int64) string {
	percent := float64(used) / float64(total) * 100
	if percent <= 60 {
		return "#409eff"
	}
	if percent <= 80 {
		return "#e6a23c"
	}

	return "#f56c6c"
}

func GetStateColor(state string) string {
	if state == "RUNNING" {
		return "#409eff"
	}

	if state == "STARTING" || state == "STOPPING" || state == "REBOOTING" {
		return "#e6a23c"
	}

	return "#f56c6c"
}

func GetCNState(state string) string {
	switch state {
	case "PENDING":
		return "创建中"
	case "LAUNCH_FAILED":
		return "创建失败"
	case "RUNNING":
		return "运行中"
	case "STOPPED":
		return "关机"
	case "STARTING":
		return "开机中"
	case "STOPPING":
		return "关机中"
	case "REBOOTING":
		return "重启中"
	case "SHUTDOWN":
		return "停止待销毁"
	case "TERMINATING":
		return "销毁中"
	default:
		return "未知状态"

	}

}

type Action string

const (
	ShutDown   Action = "关机"
	Start      Action = "开机"
	Statistics Action = "流量统计"
)

type Event struct {
	Name   string
	Used   int64
	Total  int64
	Action Action
	State  string
}

type Notifier interface {
	SendMessage(events []*Event) error
}

var notifiers map[string]*Notifier = map[string]*Notifier{}

func registerNotifier(name string, notifier Notifier) {
	notifiers[name] = &notifier
}

func SendMessage(notifyType string, events []*Event) error {
	for name, notifier := range notifiers {
		if strings.Contains(notifyType, name) {
			(*notifier).SendMessage(events)
		}
	}
	return nil
}
