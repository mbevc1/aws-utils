package lz

import (
	"aws-utils/pkg/util"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/fatih/color"
	ptable "github.com/jedib0t/go-pretty/v6/table"
)

func ListProvisionedProducts(ctx context.Context, cfg aws.Config) {
	// Create an Amazon Service Catalog client
	svc := servicecatalog.NewFromConfig(cfg)

	input := &servicecatalog.ScanProvisionedProductsInput{
		AccessLevelFilter: &types.AccessLevelFilter{
			Key:   types.AccessLevelFilterKeyAccount,
			Value: aws.String("self"),
		},
	}

	result, err := svc.ScanProvisionedProducts(ctx, input)
	if err != nil {
		log.Fatalf("failed to list provisioned products, %v", err)
	}

	if len(result.ProvisionedProducts) == 0 {
		fmt.Println("No provisioned products found.")
		return
	}

	// Construct TableWriter
	tw := ptable.NewWriter()
	tw.AppendHeader(ptable.Row{"ID", "Name", "Type", "Status", "Params"})

	fmt.Println("Provisioned Products:")
	for _, product := range result.ProvisionedProducts {
		//fmt.Printf("ID: %s, Name: %s, Status: %s\n", aws.ToString(product.Id), aws.ToString(product.Name), product.Status)
		// UNDER_CHANGE, PLAN_IN_PROGRESS Transitive states. Operations performed might not have valid results.
		if product.Status != types.ProvisionedProductStatusUnderChange && product.Status != types.ProvisionedProductStatusPlanInProgress {
			tw.AppendRow(ptable.Row{aws.ToString(product.Id), aws.ToString(product.Name), aws.ToString(product.Type), product.Status, listProvisionedProductOutputs(ctx, svc, aws.ToString(product.Id))})
		} else {
			tw.AppendRow(ptable.Row{aws.ToString(product.Id), aws.ToString(product.Name), aws.ToString(product.Type), product.Status, ""})
		}
		// Get and print the product outputs
		//listProvisionedProductOutputs(svc, aws.ToString(product.Id))
	}

	fmt.Println(tw.Render())
}

func listProvisionedProductOutputs(ctx context.Context, svc *servicecatalog.Client, provisionedProductId string) (res string) {
	input := &servicecatalog.GetProvisionedProductOutputsInput{
		ProvisionedProductId: aws.String(provisionedProductId),
	}

	result, err := svc.GetProvisionedProductOutputs(ctx, input)
	if err != nil {
		log.Fatalf("failed to get provisioned product outputs, %v", err)
	}

	if len(result.Outputs) == 0 {
		//fmt.Printf("  No outputs found for product ID: %s\n", provisionedProductId)
		return res
	}

	//fmt.Printf("  Outputs for product ID: %s\n", provisionedProductId)
	for _, output := range result.Outputs {
		//fmt.Printf("   * %s: %s (%s)\n", aws.ToString(output.OutputKey), aws.ToString(output.OutputValue), aws.ToString(output.Description))i
		if aws.ToString(output.OutputKey) != "SSOUserPortal" {
			res += fmt.Sprintf("%s: %s\n", aws.ToString(output.OutputKey), aws.ToString(output.OutputValue))
		}
	}

	return strings.TrimSpace(res)
}

func VendAccountProduct(ctx context.Context, cfg aws.Config, params util.AccountParams) {
	// Create an Amazon Service Catalog & Org clients
	svc := servicecatalog.NewFromConfig(cfg)

	fmt.Printf("Params: %+v\n", params)
	// Check for confirmation
	color.HiYellow("**WARNING**")
	c := util.AskForConfirmation("Are you sure to vend new account?")
	if !c {
		return
	}

	// SSOemail not provided, assume mgmt account email
	if params.SSOemail == "" {
		// Load the Shared AWS Configuration (~/.aws/config)
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		// Create an Organizations client
		svc := organizations.NewFromConfig(cfg)

		// Describe the organization
		orgOutput, err := svc.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
		if err != nil {
			log.Fatalf("failed to describe organization, %v", err)
		}

		// Get the management account ID
		managementAccountId := orgOutput.Organization.MasterAccountId

		// Describe the management account
		accountOutput, err := svc.DescribeAccount(ctx, &organizations.DescribeAccountInput{
			AccountId: managementAccountId,
		})
		if err != nil {
			log.Fatalf("failed to describe account, %v", err)
		}

		params.SSOemail = *accountOutput.Account.Email
		// Print the management account ID and email
		//fmt.Printf("Management Account ID: %s\n", *managementAccountId)
		//fmt.Printf("Management Account Email: %s\n", *accountOutput.Account.Email)
		color.HiCyan("**INFO**")
		fmt.Printf("SSOemail not provided, using Management Account Email: %s\n", params.SSOemail)
	}

	// Find the product ID by name
	productName := "AWS Control Tower Account Factory"
	productID, err := findProductIDByName(ctx, svc, productName)
	if err != nil {
		log.Fatalf("failed to find product ID: %v", err)
	}
	fmt.Printf("Found product ID: %s\n", productID)

	// Find the active provisioning artifact ID for the product
	provisioningArtifactID, err := findActiveProvisioningArtifactID(ctx, svc, productID)
	if err != nil {
		log.Fatalf("failed to find provisioning artifact ID: %v", err)
	}
	fmt.Printf("Found active provisioning artifact ID: %s\n", provisioningArtifactID)

	// Define the parameters for the provisioning request
	provisionedProductName := params.Name // matching account name
	parameters := []types.ProvisioningParameter{
		{
			Key:   aws.String("AccountEmail"),
			Value: aws.String(params.Email),
		},
		{
			Key:   aws.String("AccountName"),
			Value: aws.String(params.Name),
		},
		{
			Key:   aws.String("ManagedOrganizationalUnit"),
			Value: aws.String(params.OU),
		},
		{
			Key:   aws.String("SSOUserEmail"),
			Value: aws.String(params.SSOemail),
		},
		{
			Key:   aws.String("SSOUserFirstName"),
			Value: aws.String(params.SSOfirstname),
		},
		{
			Key:   aws.String("SSOUserLastName"),
			Value: aws.String(params.SSOlastname),
		},
	}

	// Create the provisioning request
	input := &servicecatalog.ProvisionProductInput{
		ProductId:              aws.String(productID),
		ProvisioningArtifactId: aws.String(provisioningArtifactID),
		ProvisionedProductName: aws.String(provisionedProductName),
		ProvisioningParameters: parameters,
	}

	// Provision the product
	result, err := svc.ProvisionProduct(ctx, input)
	if err != nil {
		log.Fatalf("failed to provision product: %v", err)
	}

	// Print the result
	fmt.Printf("Provisioned product ID: %s\n", *result.RecordDetail.ProvisionedProductId)
	fmt.Printf("Record %s (%s): %s @ %s\n", *result.RecordDetail.RecordId, *result.RecordDetail.RecordType, result.RecordDetail.Status, result.RecordDetail.UpdatedTime)
}

// findProductIDByName searches for a Service Catalog product by name and returns its product ID.
func findProductIDByName(ctx context.Context, svc *servicecatalog.Client, productName string) (string, error) {
	input := &servicecatalog.SearchProductsInput{
		Filters: map[string][]string{
			"FullTextSearch": {productName},
		},
	}

	output, err := svc.SearchProducts(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to search products: %w", err)
	}

	for _, productViewSummary := range output.ProductViewSummaries {
		if *productViewSummary.Name == productName {
			return *productViewSummary.ProductId, nil
		}
	}

	return "", fmt.Errorf("product not found: %s", productName)
}

// findActiveProvisioningArtifactID retrieves the active provisioning artifact ID for a given product ID.
func findActiveProvisioningArtifactID(ctx context.Context, svc *servicecatalog.Client, productID string) (string, error) {
	input := &servicecatalog.ListProvisioningArtifactsInput{
		ProductId: aws.String(productID),
	}

	output, err := svc.ListProvisioningArtifacts(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to list provisioning artifacts: %w", err)
	}

	for _, artifactDetail := range output.ProvisioningArtifactDetails {
		if artifactDetail.Active != nil && *artifactDetail.Active {
			return *artifactDetail.Id, nil
		}
	}

	return "", fmt.Errorf("active provisioning artifact not found for product: %s", productID)
}

func TerminateProvisionedProduct(ctx context.Context, cfg aws.Config, params util.AccountParams) {
	fmt.Printf("Params: %+v\n", params)
	// Check for confirmation
	color.HiRed("**DESTRUCTIVE**")
	c := util.AskForConfirmation("Are you sure to terminate?")
	if !c {
		return
	}
	// Close the account as well
	color.HiRed("**DESTRUCTIVE**")
	ca := util.AskForConfirmation("Are you sure to close the acccount as well?")

	// Create an Amazon Service Catalog & Org clients
	svc := servicecatalog.NewFromConfig(cfg)
	orgClient := organizations.NewFromConfig(cfg)

	// Find the provisioned product ID
	//provisionedProductId := "pp-" // Replace with the actual provisioned product ID
	provisionedProductName := params.Name // Replace with the actual provisioned product ID
	terminateProductInput := &servicecatalog.TerminateProvisionedProductInput{
		//ProvisionedProductId: aws.String(provisionedProductId),
		ProvisionedProductName: aws.String(provisionedProductName),
		TerminateToken:         aws.String("unique-terminate-token"), // Ensure this token is unique for each request
	}

	// Terminate the provisioned product
	terminateProductOutput, err := svc.TerminateProvisionedProduct(ctx, terminateProductInput)
	if err != nil {
		log.Fatalf("failed to terminate product, %v", err)
	}

	fmt.Printf("Termination started: %v\n", terminateProductOutput.RecordDetail.Status)

	// Poll for termination completion
	recordId := terminateProductOutput.RecordDetail.RecordId
	for {
		describeRecordInput := &servicecatalog.DescribeRecordInput{
			Id: recordId,
		}
		describeRecordOutput, err := svc.DescribeRecord(ctx, describeRecordInput)
		if err != nil {
			log.Fatalf("failed to describe record, %v", err)
		}

		status := describeRecordOutput.RecordDetail.Status
		if status == types.RecordStatusSucceeded {
			fmt.Println("Product terminated successfully")
			break
		} else if status == types.RecordStatusFailed {
			log.Fatalf("Product termination failed: %+v\n", describeRecordOutput.RecordDetail.RecordErrors)
		} else {
			fmt.Println("Product termination in progress...")
			time.Sleep(10 * time.Second)
		}
	}

	// Close the AWS account
	if ca {
		closeAccountInput := &organizations.CloseAccountInput{
			AccountId: aws.String(params.ID), // Replace with the actual AWS account ID
		}

		_, err = orgClient.CloseAccount(ctx, closeAccountInput)
		if err != nil {
			log.Fatalf("failed to close account, %v", err)
		}

		fmt.Println("Account closure initiated successfully")
	}
}
