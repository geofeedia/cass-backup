package main


import(
	"os"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
    awssession "github.com/aws/aws-sdk-go/aws/session"
    gcemetadata "google.golang.org/cloud/compute/metadata"
)

const (
	POD_ENV_NAME = "K8S_POD_NAME"
)


func introspectGCE(channel chan<- *CommonMetadata) {
	if gcemetadata.OnGCE() {
		cmdd := &CommonMetadata{}
		cmdd.cloud = "gce"

		if zone, err := gcemetadata.Zone(); err == nil {
			cmdd.zone = zone
		}
		if cmdd.zone != "" && len(cmdd.zone) > 1 {
			// remove the last character
			cmdd.region = cmdd.zone[:len(cmdd.zone)-2]
		}
		if hostname, err := gcemetadata.Hostname(); err == nil {
			cmdd.hostname = hostname
		}
		if instance_id, err := gcemetadata.InstanceID(); err == nil {
			cmdd.instance_id = instance_id
		}

		cmdd.pod_name = os.Getenv(POD_ENV_NAME)

		channel <- cmdd
	}
}

func introspectAWS(channel chan<- *CommonMetadata) {
	svc := ec2metadata.New(awssession.New())

	if svc.Available() {
		cmdd := &CommonMetadata{}
		cmdd.cloud = "aws"

		if zone, err := svc.GetMetadata("placement/availability-zone"); err == nil {
			cmdd.zone = zone
		}
		if cmdd.zone != "" && len(cmdd.zone) > 2 {
			// remove the last 2 characters
			cmdd.region = cmdd.zone[:len(cmdd.zone)-1]
		}
		if hostname, err := svc.GetMetadata("hostname"); err == nil {
			cmdd.hostname = hostname
		}
		if instance_id, err := svc.GetMetadata("instance-id"); err == nil {
			cmdd.instance_id = instance_id
		}

		cmdd.pod_name = os.Getenv(POD_ENV_NAME)

		channel <- cmdd
	}
}