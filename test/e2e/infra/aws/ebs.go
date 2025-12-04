package aws

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ebs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EBSVolume(ctx *pulumi.Context) error {
	// Create a new EBS volume.
	_, err := ebs.NewVolume(ctx, "example-volume", &ebs.VolumeArgs{
		AvailabilityZone: pulumi.String("us-east-1a"),
		Size:             pulumi.Int(8),
		Type:             pulumi.String("gp3"),
		Tags:             pulumi.StringMap{"Name": pulumi.String("example-volume")},
	})
	if err != nil {
		return err
	}

	return nil
}
