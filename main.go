package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/lixiaofei326/lhmonitor/notifier"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

var threshold int
var noticeType string
var autoStart bool
var runInterval int
var reportStatTime int

func init() {

	var cstZone = time.FixedZone("CST", 8*3600) // East 8 District
	time.Local = cstZone

	thresholdStr := os.Getenv("THRESHOLD")
	var err error
	threshold, err = strconv.Atoi(thresholdStr)
	if err != nil {
		threshold = 90
	}

	noticeType = os.Getenv("NOTICETYPE")

	autoStartStr := os.Getenv("AUTOSTART")
	autoStart, err = strconv.ParseBool(autoStartStr)
	if err != nil {
		autoStart = false
	}

	runIntervalStr := os.Getenv("RUNINTERVAL")
	runInterval, err = strconv.Atoi(runIntervalStr)
	if err != nil {
		runInterval = 60
	}

	reportStatTimeStr := os.Getenv("REPORTTIME")
	reportStatTime, err = strconv.Atoi(reportStatTimeStr)
	if err != nil {
		reportStatTime = 8
	}

}

func run(ctx context.Context) (string, error) {

	needReport := false
	var cstSh, _ = time.LoadLocation("Asia/Shanghai")
	minuteOfDay := time.Now().In(cstSh).Hour()*60 + time.Now().Minute()
	if minuteOfDay >= reportStatTime*60 && minuteOfDay < reportStatTime*60+runInterval {
		needReport = true
	}

	bin := NewLHBin()
	instances, err := bin.ListInstances()
	if err != nil {
		log.Panicf(err.Error())
	}

	events := []*notifier.Event{}

	for _, instance := range instances {
		tos, err := bin.QueryTrafficPackages(instance.Region, instance.InstanceId)
		if err == nil {
			for _, to := range tos {
				userPercent := float64(to.Used) / float64(to.Total) * 100
				if int(userPercent) >= threshold {
					if instance.State == "RUNNING" {
						// 执行关机操作
						_, err := bin.StopInstance(instance.Region, instance.InstanceId)
						if err != nil {
							return "", err
						}
						events = append(events, &notifier.Event{
							Name:   instance.InstanceName,
							Used:   to.Used,
							Total:  to.Total,
							Action: notifier.ShutDown,
							State:  instance.State,
						})
					}
				} else {
					if instance.State == "STOPPED" && autoStart {
						// 执行开机操作
						_, err := bin.StartInstance(instance.Region, instance.InstanceId)
						if err != nil {
							return "", err
						}
						events = append(events, &notifier.Event{
							Name:   instance.InstanceName,
							Used:   to.Used,
							Total:  to.Total,
							Action: notifier.Start,
							State:  instance.State,
						})
					}
				}

				if needReport {
					events = append(events, &notifier.Event{
						Name:   instance.InstanceName,
						Used:   to.Used,
						Total:  to.Total,
						Action: notifier.Statistics,
						State:  instance.State,
					})
				}
			}
		}
	}

	notifier.SendMessage(noticeType, events)

	return "", nil
}

func main() {

	cloudfunction.Start(run)
}
