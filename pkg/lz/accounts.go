package lz

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/fatih/color"
)

func ListAccounts(ctx context.Context, cfg aws.Config) {
	// Create an Organizations client from the configuration
	svc := organizations.NewFromConfig(cfg)

	// Get info about the Organizations
	res, err := svc.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		log.Fatalf("failed to get Organizations details: %v", err)
	}
	fmt.Printf("Organization ID:\t%s\n", color.HiGreenString(*res.Organization.Id))
	fmt.Printf("Management Account:\t%s (%s)\n", color.HiYellowString(*res.Organization.MasterAccountId), *res.Organization.MasterAccountEmail)
	fmt.Println("---")

	// List the root(s) to get their ID(s)
	rootsOutput, err := svc.ListRoots(ctx, &organizations.ListRootsInput{})
	if err != nil {
		log.Fatalf("failed to list roots: %v", err)
	}

	for _, root := range rootsOutput.Roots {
		fmt.Printf("%s (%s)\n", *root.Name, *root.Id)

		// List accounts in the root
		listAccounts(ctx, svc, *root.Id, "│  ")

		// List OUs and their child accounts
		listOUsAndAccounts(ctx, svc, *root.Id, "")
	}
}

func listOUsAndAccounts(ctx context.Context, svc *organizations.Client, parentID string, prefix string) {
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

			// List accounts in the current OU
			listAccounts(ctx, svc, *ou.Id, newPrefix)

			// Recursive call to list child OUs
			listOUsAndAccounts(ctx, svc, *ou.Id, newPrefix)
		}
	}
}

func listAccounts(ctx context.Context, svc *organizations.Client, parentID string, prefix string) {
	// Create a new tabwriter.Writer instance.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	input := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(parentID),
	}

	paginator := organizations.NewListAccountsForParentPaginator(svc, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to list accounts for parent %s: %v", parentID, err)
		}

		for i, account := range page.Accounts {
			if i == len(page.Accounts)-1 {
				fmt.Fprintf(w, "%s└── %s (%s)\t |\t %s\t |\t %s\n", prefix, *account.Name, *account.Id, *account.Email, account.State)
			} else {
				fmt.Fprintf(w, "%s├── %s (%s)\t |\t %s\t |\t %s\n", prefix, *account.Name, *account.Id, *account.Email, account.State)
			}
		}
	}

	// Flush the Writer to ensure all data is written to the output.
	w.Flush()
}
