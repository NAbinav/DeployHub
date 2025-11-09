package helper

import (
	"context"
	"fmt"
	"os"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/option"
)

func DeleteDeploy(ctx context.Context, serviceName string) error {
	runClient, err := run.NewServicesClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	logStream := make(chan string, 100) // Buffered channel to avoid blocking
	defer close(logStream)
	region := "asia-south1"
	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	name := fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, region, serviceName)

	op, err := runClient.DeleteService(ctx, &runpb.DeleteServiceRequest{
		Name: name,
	})
	if err != nil {
		logStream <- fmt.Sprintf("DeleteService failed: %v", err)
		return err
	}

	logStream <- fmt.Sprintf("Deleting service %s...", serviceName)
	_, err = op.Wait(ctx)
	if err != nil {
		logStream <- fmt.Sprintf("Delete operation failed: %v", err)
		return err
	}

	logStream <- fmt.Sprintf("Service %s deleted successfully", serviceName)
	return nil
}
