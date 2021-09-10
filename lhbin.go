package main

import (
	"os"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

var regions []string = []string{"ap-beijing", "ap-chengdu", "ap-guangzhou", "ap-hongkong", "ap-shanghai", "ap-singapore", "ap-tokyo", "ap-nanjing"}
var secretId string
var secretKey string
var Endpoint string = "lighthouse.tencentcloudapi.com"

func init() {
	secretId = os.Getenv("SECRET_ID")
	secretKey = os.Getenv("SECRET_KEY")
}

type LHInstance struct {
	Region       string `json:"region"`
	InstanceId   string `json:"instanceID"`
	InstanceName string `json:"instanceName"`
	State        string `json:"state"`
}

type TrafficPackageInfo struct {
	Total      int64 `json:"total"`
	Used       int64 `json:"used"`
	CreateTime int64 `json:"createTime"`
	ExipreTime int64 `json:"expireTime"`
}

type LHBin interface {
	ListInstances() ([]*LHInstance, error)
	QueryTrafficPackages(region, instanceID string) ([]*TrafficPackageInfo, error)
	StartInstance(region, instanceID string) (bool, error)
	StopInstance(region, instanceID string) (bool, error)
}

func NewLHBin() LHBin {
	return &libin{}
}

type libin struct {
}

func (bin *libin) ListInstances() ([]*LHInstance, error) {

	instances := []*LHInstance{}

	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = Endpoint

	for _, region := range regions {
		client, _ := lighthouse.NewClient(credential, region, cpf)
		request := lighthouse.NewDescribeInstancesRequest()
		response, err := client.DescribeInstances(request)
		if err != nil {
			return nil, err
		}

		for _, instance := range response.Response.InstanceSet {
			instances = append(instances, &LHInstance{
				InstanceId:   *instance.InstanceId,
				InstanceName: *instance.InstanceName,
				Region:       region,
				State:        *instance.InstanceState,
			})
		}

	}

	return instances, nil

}

func (bin *libin) QueryTrafficPackages(region, instanceID string) ([]*TrafficPackageInfo, error) {

	packages := []*TrafficPackageInfo{}

	credential := common.NewCredential(
		secretId,
		secretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = Endpoint
	request := lighthouse.NewDescribeInstancesTrafficPackagesRequest()
	request.InstanceIds = common.StringPtrs([]string{instanceID})

	client, _ := lighthouse.NewClient(credential, region, cpf)
	response, err := client.DescribeInstancesTrafficPackages(request)
	if err != nil {
		return nil, err
	}

	if response.Response.InstanceTrafficPackageSet != nil {
		for _, instanceTrafficPackage := range response.Response.InstanceTrafficPackageSet {
			if *(instanceTrafficPackage.InstanceId) == instanceID {
				for _, trafficPackage := range instanceTrafficPackage.TrafficPackageSet {
					createTime, _ := time.Parse("2006-01-02T15:04:05Z", *trafficPackage.StartTime)
					expireTime, _ := time.Parse("2006-01-02T15:04:05Z", *trafficPackage.EndTime)

					packages = append(packages, &TrafficPackageInfo{
						Total:      *trafficPackage.TrafficPackageTotal,
						Used:       *trafficPackage.TrafficUsed,
						CreateTime: createTime.Unix(),
						ExipreTime: expireTime.Unix(),
					})
				}
			}
		}

	}
	return packages, nil
}

func (bin *libin) StartInstance(region, instanceID string) (bool, error) {
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = Endpoint
	client, _ := lighthouse.NewClient(credential, region, cpf)
	request := lighthouse.NewStartInstancesRequest()
	request.InstanceIds = []*string{&instanceID}

	_, err := client.StartInstances(request)
	if err != nil {
		return false, nil
	}
	return true, nil
}
func (bin *libin) StopInstance(region, instanceID string) (bool, error) {
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = Endpoint
	client, _ := lighthouse.NewClient(credential, region, cpf)
	request := lighthouse.NewStopInstancesRequest()
	request.InstanceIds = []*string{&instanceID}

	_, err := client.StopInstances(request)
	if err != nil {
		return false, nil
	}
	return true, nil
}
