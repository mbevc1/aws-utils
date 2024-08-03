package lz

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

func FindOUs(ctx context.Context, cfg aws.Config, name string) {
	// Create Organizations client
	orgClient := organizations.NewFromConfig(cfg)

	// Specify the OU name you want to find
	ouName := name

	// Start with the root parent ID
	rootID, err := getRootID(ctx, orgClient)
	if err != nil {
		log.Fatalf("failed to get root ID: %v", err)
	}

	// Call ListOrganizationalUnitsForParent recursively to get details of all OUs
	ouID, err := findOUByName(ctx, orgClient, rootID, ouName)
	if err != nil {
		log.Fatalf("failed to find OU by name: %v", err)
	}

	if ouID == "" {
		fmt.Printf("OU with name '%s' not found!\n", ouName)

		return
	}

	fmt.Printf("Found OU for '%s', ID: %s\n", ouName, ouID)
}

// getRootID retrieves the root ID of the AWS Organization
func getRootID(ctx context.Context, client *organizations.Client) (string, error) {
	input := &organizations.ListRootsInput{}
	output, err := client.ListRoots(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to list roots: %w", err)
	}

	if len(output.Roots) == 0 {
		return "", fmt.Errorf("no roots found in the organization")
	}

	return *output.Roots[0].Id, nil
}

// findOUByName recursively searches for the OU by name
func findOUByName(ctx context.Context, client *organizations.Client, parentID, ouName string) (string, error) {
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parentID),
	}

	output, err := client.ListOrganizationalUnitsForParent(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to list OUs for parent %s: %w", parentID, err)
	}

	for _, ou := range output.OrganizationalUnits {
		if *ou.Name == ouName {
			return *ou.Id, nil
		}
		// Recursively search child OUs
		ouID, err := findOUByName(ctx, client, *ou.Id, ouName)
		if err == nil && ouID != "" {
			return ouID, nil
		}
	}

	return "", nil
}

func ListOUs(ctx context.Context, cfg aws.Config) {
	// Create an Organizations client from the configuration
	svc := organizations.NewFromConfig(cfg)

	// List the root(s) to get their ID(s)
	rootsOutput, err := svc.ListRoots(ctx, &organizations.ListRootsInput{})
	if err != nil {
		log.Fatalf("failed to list roots: %v", err)
	}

	for _, root := range rootsOutput.Roots {
		fmt.Printf("Root ID: %s, Name: %s\n", *root.Id, *root.Name)
		listOUs(ctx, svc, *root.Id, "")
	}
}

func listOUs(ctx context.Context, svc *organizations.Client, parentID string, prefix string) {
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parentID),
	}

	paginator := organizations.NewListOrganizationalUnitsForParentPaginator(svc, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to list OUs for parent %s: %v", parentID, err)
		}

		for i, ou := range page.OrganizationalUnits {
			var newPrefix string
			if i == len(page.OrganizationalUnits)-1 {
				fmt.Printf("%s└── %s (%s)\n", prefix, *ou.Name, *ou.Id)
				newPrefix = prefix + "    "
			} else {
				fmt.Printf("%s├── %s (%s)\n", prefix, *ou.Name, *ou.Id)
				newPrefix = prefix + "│   "
			}
			listOUs(ctx, svc, *ou.Id, newPrefix) // Recursive call to list child OUs
		}
	}
}

func DescribeOU(ctx context.Context, cfg aws.Config, Id string) {
	// Create an Organizations client from the configuration
	svc := organizations.NewFromConfig(cfg)

	// List the root(s) to get their ID(s)
	res, err := svc.DescribeOrganizationalUnit(ctx, &organizations.DescribeOrganizationalUnitInput{OrganizationalUnitId: &Id})
	if err != nil {
		log.Fatalf("failed to describe OU: %v", err)
	}

	fmt.Println("OU details:")
	fmt.Println("-----------")
	fmt.Printf("Name:\t%s\n", *res.OrganizationalUnit.Name)
	fmt.Printf("ID:\t%s\n", *res.OrganizationalUnit.Id)
	fmt.Printf("ARN:\t%s\n", *res.OrganizationalUnit.Arn)
}
