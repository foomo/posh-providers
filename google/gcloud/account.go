package gcloud

import (
	"regexp"
)

var gcloudAccountFileNameRegex = regexp.MustCompile(`(\w+)@(\w+)-(\w+)\.json`)

type Account struct {
	Role        string
	Environment string
	Cluster     string
	Path        string
}