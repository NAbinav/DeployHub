package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/option"
)

func DeleteDeploy(ctx context.Context, serviceName string) error {
	region := "asia-south1"
	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	cred := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	runClient, err := run.NewServicesClient(ctx, option.WithCredentialsFile(cred))
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run client: %w", err)
	}
	defer runClient.Close()

	name := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, region, serviceName)
	log.Printf("Deleting Cloud Run service: %s", name)

	op, err := runClient.DeleteService(ctx, &runpb.DeleteServiceRequest{Name: name})
	if err != nil {
		return fmt.Errorf("failed to delete Cloud Run service: %w", err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed waiting for service deletion: %w", err)
	}
	log.Println("Cloud Run service deleted successfully")

	// 2. Delete image from shared repo
	svc, err := artifactregistry.NewService(ctx, option.WithCredentialsFile(cred))
	if err != nil {
		return fmt.Errorf("failed to create Artifact Registry client: %w", err)
	}

	repo := fmt.Sprintf("projects/%s/locations/%s/repositories/deploy-hub", projectID, region)
	log.Printf("Listing packages in repository: %s", repo)

	packages, err := svc.Projects.Locations.Repositories.Packages.List(repo).Do()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	imageFound := false
	for _, pkg := range packages.Packages {
		// Package name format: projects/.../repositories/.../packages/SERVICE-NAME-image
		if strings.Contains(pkg.Name, serviceName+"-image") {
			log.Printf("Deleting package: %s", pkg.Name)
			imageFound = true

			// Delete the entire package (which includes all versions)
			op, err := svc.Projects.Locations.Repositories.Packages.Delete(pkg.Name).Do()
			if err != nil {
				log.Printf("Failed to delete package %s: %v", pkg.Name, err)
				continue
			}
			log.Printf("Package deletion initiated: %+v", op)
		}
	}

	if !imageFound {
		log.Printf("No packages found matching '%s-image'", serviceName)
	}
	log.Println("Cleanup completed")
	return nil
}
