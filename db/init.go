package db

import (
	"context"
	"os"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/d1"
	"github.com/cloudflare/cloudflare-go/v3/option"
)

var Client *cloudflare.Client

func Init_Cloudflare() error {
	Client = cloudflare.NewClient(option.WithAPIKey(os.Getenv("CLOUDFLARE_API_TOKEN")))
	_, err := Client.D1.Database.List(context.TODO(), d1.DatabaseListParams{
		AccountID: cloudflare.F(os.Getenv("CLOUDFLARE_ACCOUNT_ID")),
	})
	return err
}
