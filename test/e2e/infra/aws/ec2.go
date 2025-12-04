package aws

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EC2Instance(ctx *pulumi.Context) error {
	// Create a new security group for the instance.
	group, err := ec2.NewSecurityGroup(ctx, "web-secgrp", &ec2.SecurityGroupArgs{
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(80),
				ToPort:     pulumi.Int(80),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	})
	if err != nil {
		return err
	}

	// Look up the latest Amazon Linux 2 AMI.
	ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		MostRecent: pulumi.BoolRef(true),
		Owners:     []string{"amazon"},
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"amzn2-ami-hvm-*-x86_64-gp2"},
			},
		},
	})
	if err != nil {
		return err
	}

	// Create a new EC2 instance.
	_, err = ec2.NewInstance(ctx, "web-server-www", &ec2.InstanceArgs{
		InstanceType:        pulumi.String("t3.micro"),
		SecurityGroups:      pulumi.StringArray{group.Name},
		Ami:                 pulumi.String(ami.Id),
		Tags:                pulumi.StringMap{"Name": pulumi.String("web-server-www")},
	})
	if err != nil {
		return err
	}

	return nil
}
