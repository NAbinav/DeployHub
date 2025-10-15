package cmd

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
)

func CloneRepo(url string, name string) error {
	_, err := git.PlainClone("/tmp/"+name, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	if err != nil {
		fmt.Println(err)
		log.Fatalf("ERR: %s", err)
	}
	return nil
}

func Pack(u)
