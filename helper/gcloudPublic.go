package helper

import (
	"context"
	"fmt"

	iampb "cloud.google.com/go/iam/apiv1/iampb"
	run "cloud.google.com/go/run/apiv2"
)

func MakeCloudRunPublic(
	ctx context.Context,
	projectID, region, serviceName string,
) error {
	fmt.Println("=== MakeCloudRunPublic START ===")
	fmt.Printf("ProjectID: %s, Region: %s, ServiceName: %s\n", projectID, region, serviceName)

	fmt.Println("Creating Cloud Run client...")
	client, err := run.NewServicesClient(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to create client: %v\n", err)
		return err
	}
	defer client.Close()
	fmt.Println("✓ Client created successfully")

	resource := fmt.Sprintf(
		"projects/%s/locations/%s/services/%s",
		projectID, region, serviceName,
	)
	fmt.Printf("Resource path: %s\n", resource)

	// Get current IAM policy
	fmt.Println("Fetching current IAM policy...")
	policy, err := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: resource,
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to get IAM policy: %v\n", err)
		return err
	}
	fmt.Printf("✓ Policy retrieved - Version: %d, Bindings: %d\n", policy.Version, len(policy.Bindings))

	const (
		role   = "roles/run.invoker"
		member = "allUsers"
	)

	// Ensure policy version
	policy.Version = 3

	// Check if binding already exists
	found := false
	for i, b := range policy.Bindings {
		if b.Role == role {
			fmt.Printf("Found existing binding at index %d for role: %s\n", i, role)
			fmt.Printf("Current members: %v\n", b.Members)

			// Check if member already exists
			for _, m := range b.Members {
				if m == member {
					fmt.Printf("✓ Member '%s' already exists - service is already public\n", member)
					return nil
				}
			}

			// Add member to existing binding
			fmt.Printf("Adding '%s' to existing binding\n", member)
			b.Members = append(b.Members, member)
			found = true
			break
		}
	}

	// Add new binding if role doesn't exist
	if !found {
		fmt.Printf("No existing binding found - creating new binding for role: %s\n", role)
		policy.Bindings = append(policy.Bindings, &iampb.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	// Set the updated policy
	fmt.Println("Setting updated IAM policy...")
	fmt.Printf("Total bindings to set: %d\n", len(policy.Bindings))
	_, err = client.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: resource,
		Policy:   policy,
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to set IAM policy: %v\n", err)
		return err
	}

	fmt.Println("✓ Successfully made Cloud Run service public!")
	fmt.Println("=== MakeCloudRunPublic END ===")
	return nil
}
