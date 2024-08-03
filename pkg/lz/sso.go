package lz

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/fatih/color"
)

func ListSSO(ctx context.Context, cfg aws.Config) {
	// Print the list of SSO instances - SDK returns array!
	fmt.Println("-------------------------------")
	fmt.Println("List of deployed SSO instances:")
	fmt.Println("-------------------------------")

	for _, instance := range getSSOinstances(ctx, cfg) {
		fmt.Printf("ARN: %s\n", color.HiMagentaString(aws.ToString(instance.InstanceArn)))
		fmt.Printf("Name: %s\n", color.HiRedString(aws.ToString(instance.Name)))
		fmt.Printf("Owner account ID: %s\n", color.HiBlueString(aws.ToString(instance.OwnerAccountId)))
		if instance.Status == types.InstanceStatusActive {
			fmt.Printf("Status: %s\n", color.HiGreenString(fmt.Sprintf("%s", instance.Status)))
		} else {
			fmt.Printf("Status: %s\n", color.HiYellowString(fmt.Sprintf("%s", instance.Status)))
		}
		fmt.Printf("IdentityStoreId: %s\n", color.HiYellowString(aws.ToString(instance.IdentityStoreId)))
		fmt.Printf("CreatedDate: %s\n", color.HiCyanString(instance.CreatedDate.Format(time.RFC1123)))
	}
}

func ListPermissionSets(ctx context.Context, cfg aws.Config) {
	// Create an SSOadmin client from the configuration
	svc := ssoadmin.NewFromConfig(cfg)
	instances := getSSOinstances(ctx, cfg)

	// Get info about the IAM IC
	r, err := svc.ListPermissionSets(ctx, &ssoadmin.ListPermissionSetsInput{InstanceArn: instances[0].InstanceArn})
	if err != nil {
		log.Fatalf("failed to get SSO PermissionSets: %v", err)
	}
	// Print the list of SSO PermissionSets - SDK returns array!
	fmt.Println("List of PermissionSets:")
	fmt.Println("-----------------------")

	for _, ps := range r.PermissionSets {
		fmt.Printf("%s\n", color.HiYellowString(ps))
		res, err := svc.DescribePermissionSet(ctx, &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      instances[0].InstanceArn,
			PermissionSetArn: &ps,
		})
		if err != nil {
			log.Fatalf("failed to describe SSO PermissionSet: %v", err)
		}
		fmt.Printf("Name: %s\n", color.HiGreenString(aws.ToString(res.PermissionSet.Name)))
		fmt.Printf("Description: %s\n", aws.ToString(res.PermissionSet.Description))
		fmt.Printf("SessionDuration: %s\n", aws.ToString(res.PermissionSet.SessionDuration))
		fmt.Printf("CreatedDate: %s\n", res.PermissionSet.CreatedDate.Format(time.RFC1123))
		fmt.Println("--")
	}
}

func getSSOinstances(ctx context.Context, cfg aws.Config) (Instances []types.InstanceMetadata) {
	// Create an SSOadmin client from the configuration
	svc := ssoadmin.NewFromConfig(cfg)

	// Get info about the IAM IC
	res, err := svc.ListInstances(ctx, &ssoadmin.ListInstancesInput{})
	if err != nil {
		log.Fatalf("failed to get SSO instances: %v", err)
	}

	return res.Instances
}
