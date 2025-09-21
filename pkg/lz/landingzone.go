package lz

import (
	"aws-utils/pkg/util"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/fatih/color"
	ptable "github.com/jedib0t/go-pretty/v6/table"
)

func ListLandingZone(ctx context.Context, cfg aws.Config) {
	fmt.Println("Using region: " + color.HiCyanString(cfg.Region))
	// CT client
	ct := controltower.NewFromConfig(cfg)

	// Create the input for the ListLandingZones call
	input := &controltower.ListLandingZonesInput{}

	result, err := ct.ListLandingZones(ctx, input)

	if err != nil {
		log.Fatalf("failed to list landing zones, %v", err)
	}

	// Print the list of landing zones - SDK returns array!
	fmt.Println("-------------------------------")
	fmt.Println("List of deployed Landing Zones:")
	fmt.Println("-------------------------------")

	for _, landingZone := range result.LandingZones {
		fmt.Printf("ARN: %s\n", color.HiYellowString(aws.ToString(landingZone.Arn)))
		sel := &controltower.GetLandingZoneInput{LandingZoneIdentifier: landingZone.Arn}
		res, err := ct.GetLandingZone(ctx, sel)
		if err != nil {
			log.Fatalf("failed to get landing zone details, %v", err)
		}

		fmt.Printf("Version: %s\n", aws.ToString(res.LandingZone.Version))
		fmt.Printf("LatestAvailableVersion: %s\n", aws.ToString(res.LandingZone.LatestAvailableVersion))

		//fmt.Printf("%+v\n", res.LandingZone.Manifest)
		var kv map[string]interface{}
		if err := res.LandingZone.Manifest.UnmarshalSmithyDocument(&kv); err != nil {
			// handle error
			fmt.Println("error:", err)
		}
		// prettify JSON document
		b, err := json.MarshalIndent(kv, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Printf("Manifest:\n%s\n", b)
		fmt.Printf("Status: %s\n", res.LandingZone.Status)
		fmt.Printf("DriftStatus: %s\n", res.LandingZone.DriftStatus.Status)
	}
}

func FindAccounts(ctx context.Context, cfg aws.Config, Id string, filter string) {
	// Create an Organizations client
	svc := organizations.NewFromConfig(cfg)

	// Max results, 20 is hard limit
	var mr int32 = 20
	// Create a paginator to handle paginated results
	paginator := organizations.NewListAccountsPaginator(svc, &organizations.ListAccountsInput{MaxResults: &mr})

	if cfg.Region != "" {
		slog.Debug(fmt.Sprintf("Selected region: %s", cfg.Region))
	}

	fmt.Println("List of AWS Organizations Accounts:")
	for paginator.HasMorePages() {
		// Get the next page of results
		output, err := paginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get page, %v", err)
		}

		// Construct TableWriter
		tw := ptable.NewWriter()
		tw.AppendHeader(ptable.Row{"ID", "Email", "Name", "OU", "State"})

		// Process each account in the page
		for _, account := range output.Accounts {
			var ou string

			// List parents of the account
			parentsResp, err := svc.ListParents(ctx, &organizations.ListParentsInput{
				ChildId: aws.String(aws.ToString(account.Id)),
			})
			if err != nil {
				log.Fatalf("failed to list parents, %v", err)
			}

			for _, parent := range parentsResp.Parents {
				if parent.Type == types.ParentTypeOrganizationalUnit {
					// Describe the Organizational Unit
					ouResp, err := svc.DescribeOrganizationalUnit(ctx, &organizations.DescribeOrganizationalUnitInput{
						OrganizationalUnitId: parent.Id,
					})
					if err != nil {
						log.Fatalf("failed to describe organizational unit, %v", err)
					}

					// Print the OU name
					//fmt.Printf("Account is part of Organizational Unit: %s\n", *ouResp.OrganizationalUnit.Name)
					ou = *ouResp.OrganizationalUnit.Name

					// Find Root OU ID
					//rootId, err := getRootId(svc)
					//if err != nil {
					//	log.Fatalf("failed to get root id, %v", err)
					//}

					// Crawl OU path until get to Root
					//path, err := findAccountOUPath(svc, rootId, aws.ToString(account.Id), nil)
					//if err != nil {
					//	log.Fatalf("failed to find account OU path, %v", err)
					//}

					/*if path == nil {
						fmt.Printf("Account ID %s not found in the organization\n", aws.ToString(account.Id))
					} else {
						fmt.Printf("Account ID %s is in the following OU path: %s\n", aws.ToString(account.Id), formatPath(path))
					}*/
					//ou = formatPath(path)

				} else {
					ou = "Root"
				}
			}

			accept := false

			// Apply filters
			if filter != "" && strings.Contains(*account.Name, filter) {
				accept = true
			}
			if Id != "" && strings.HasPrefix(*account.Id, Id) {
				//if filter != "" || strings.Contains(*account.Name, filter) {
				//	tw.AppendRow(getAccountDetails(account, ou))
				//} else {
				accept = true
			}
			if Id == "" && filter == "" {
				accept = true
			}
			if accept {
				tw.AppendRow(getAccountDetails(account, ou))
			}
		}
		fmt.Println(tw.Render())
	}
}

func getAccountDetails(account types.Account, ou string) ptable.Row {
	/*fmt.Printf("ID: %s, Email: %s, Name: %s, Status: %s\n",
	  aws.ToString(account.Id),
	  aws.ToString(account.Email),
	  aws.ToString(account.Name),
	  account.Status)*/

	return ptable.Row{aws.ToString(account.Id), aws.ToString(account.Email), aws.ToString(account.Name), ou, account.State}
}

func getRootId(ctx context.Context, client *organizations.Client) (string, error) {
	input := &organizations.ListRootsInput{}

	result, err := client.ListRoots(ctx, input)
	if err != nil {
		return "", err
	}

	if len(result.Roots) == 0 {
		return "", fmt.Errorf("no roots found")
	}

	return *result.Roots[0].Id, nil
}

func findAccountOUPath(ctx context.Context, client *organizations.Client, parentId string, accountId string, parentPath []string) ([]string, error) {
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: &parentId,
	}

	result, err := client.ListOrganizationalUnitsForParent(ctx, input)
	if err != nil {
		return nil, err
	}

	for _, ou := range result.OrganizationalUnits {
		currentPath := append(parentPath, *ou.Name)

		// Check if the account is in the current OU
		found, err := isAccountInOU(ctx, client, *ou.Id, accountId)
		if err != nil {
			return nil, err
		}

		if found {
			return currentPath, nil
		}

		// Recursively check sub-OUs
		path, err := findAccountOUPath(ctx, client, *ou.Id, accountId, currentPath)
		if err != nil {
			return nil, err
		}

		if path != nil {
			return path, nil
		}
	}

	return nil, nil
}

func isAccountInOU(ctx context.Context, client *organizations.Client, ouId string, accountId string) (bool, error) {
	input := &organizations.ListAccountsForParentInput{
		ParentId: &ouId,
	}

	result, err := client.ListAccountsForParent(ctx, input)
	if err != nil {
		return false, err
	}

	for _, account := range result.Accounts {
		if *account.Id == accountId {
			return true, nil
		}
	}

	return false, nil
}

func formatPath(path []string) string {
	//return fmt.Sprintf("%s", path)
	return strings.Join(path, "/")
}

func ResetLandingZone(ctx context.Context, cfg aws.Config) {
	var lzArn string
	// CT client
	ct := controltower.NewFromConfig(cfg)

	res, e := ct.ListLandingZones(ctx, &controltower.ListLandingZonesInput{})

	if e != nil {
		log.Fatalf("failed to list landing zones, %v", e)
	}

	// Check for confirmation
	color.HiYellow("**WARNING**")
	c := util.AskForConfirmation(fmt.Sprintf("Are you sure you want to reset your LZ?"))

	if !c {
		return
	}

	if len(res.LandingZones) > 0 {
		lzArn = aws.ToString(res.LandingZones[0].Arn)
		fmt.Printf("Found Landing Zone with ARN: %s\n", lzArn)
	} else {
		fmt.Println("Landing Zone not found. Nothing to do!")
		return
	}

	// Create the input for the ListLandingZones call
	input := &controltower.ResetLandingZoneInput{LandingZoneIdentifier: &lzArn}

	result, err := ct.ResetLandingZone(ctx, input)

	if err != nil {
		log.Fatalf("failed to list landing zones, %v", err)
	}

	// Print the list of landing zones - SDK returns array!
	fmt.Printf("Landing Zone reset triggered: %s\n", *result.OperationIdentifier)
}

func UpdateLandingZone(ctx context.Context, cfg aws.Config) {
	// CT client
	ct := controltower.NewFromConfig(cfg)

	// Create the input for the ListLandingZones call
	input := &controltower.ListLandingZonesInput{}

	result, err := ct.ListLandingZones(ctx, input)

	if err != nil {
		log.Fatalf("failed to list landing zones, %v", err)
	}

	// Print the list of landing zones - SDK returns array!
	fmt.Println("-------------------------------")
	for _, landingZone := range result.LandingZones {
		fmt.Printf("ARN: %s\n", aws.ToString(landingZone.Arn))
		sel := &controltower.GetLandingZoneInput{LandingZoneIdentifier: landingZone.Arn}
		res, _ := ct.GetLandingZone(ctx, sel)
		fmt.Printf("Current version: %s\n", aws.ToString(res.LandingZone.Version))

		// Check for confirmation
		color.HiYellow("**WARNING**")
		c := util.AskForConfirmation(fmt.Sprintf("Are you sure to update your LZ to LatestAvailableVersion (%s) ?", color.HiGreenString(aws.ToString(res.LandingZone.LatestAvailableVersion))))
		if !c {
			return
		}

		//fmt.Printf("%+v\n", res.LandingZone.Manifest)
		var kv map[string]interface{}
		if err := res.LandingZone.Manifest.UnmarshalSmithyDocument(&kv); err != nil {
			// handle error
			fmt.Println("error:", err)
		}
		// prettify JSON document
		//b, err := json.MarshalIndent(kv, "", "  ")
		//if err != nil {
		//	fmt.Println("error:", err)
		//}
		//fmt.Printf("Manifest:\n%s\n", b)
		//fmt.Printf("Status: %s\n", res.LandingZone.Status)
		//fmt.Println("-------------------------------")
		ures, err := ct.UpdateLandingZone(ctx, &controltower.UpdateLandingZoneInput{
			LandingZoneIdentifier: landingZone.Arn,
			Manifest:              res.LandingZone.Manifest,
			Version:               res.LandingZone.LatestAvailableVersion},
		)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Println("Update started: " + *ures.OperationIdentifier)
	}
}

func ListOpsLandingZone(ctx context.Context, cfg aws.Config) {
	// CT client
	ct := controltower.NewFromConfig(cfg)

	result, err := ct.ListLandingZoneOperations(ctx, &controltower.ListLandingZoneOperationsInput{})

	if err != nil {
		log.Fatalf("failed to list landing zones, %v", err)
	}

	for _, oo := range result.LandingZoneOperations {
		r, _ := ct.GetLandingZoneOperation(ctx, &controltower.GetLandingZoneOperationInput{OperationIdentifier: oo.OperationIdentifier})
		fmt.Println("Operation: ", *oo.OperationIdentifier)
		fmt.Println("OperationType: ", oo.OperationType)
		fmt.Println("Status: ", oo.Status)
		fmt.Println("StatusMessage: ", r.OperationDetails.StatusMessage)
		fmt.Println("StartTime: ", r.OperationDetails.StartTime)
		fmt.Println("EndTime: ", r.OperationDetails.EndTime)
		fmt.Println("----------")
	}
}
