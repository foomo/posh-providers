package gcloud

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/foomo/posh/pkg/util/files"
)

var gcloudAccountFileNameRegex = regexp.MustCompile(`(\w+)@(\w+)-(\w+)\.json`)

type GCloudAccount struct {
	Role        string
	Environment string
	Cluster     string
	Path        string
}

func (a GCloudAccount) Label() string {
	return fmt.Sprintf("")
}

func ParseAccounts(ctx context.Context, cnfDir string) ([]GCloudAccount, error) {
	accountFiles, err := files.Find(ctx, cnfDir, "*.json")
	if err != nil {
		return nil, err
	}

	var accounts []GCloudAccount
	for _, f := range accountFiles {
		matchString := gcloudAccountFileNameRegex.FindAllStringSubmatch(filepath.Base(f), 1)
		if len(matchString) == 0 {
			continue
		}
		match := matchString[0]
		acc := GCloudAccount{
			Role:        match[1],
			Environment: match[2],
			Cluster:     match[3],
			Path:        f,
		}
		accounts = append(accounts, acc)
	}

	return accounts, err
}
