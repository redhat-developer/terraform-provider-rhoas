package build

import (
	"runtime/debug"
)

// Define public variables here which you wish to be configurable at build time
var (
	// Version is dynamically set by the toolchain or overridden by the Makefile.
	Version = "dev"

	// Language used, can be overridden by Makefile or CI
	Language = "en"

	// RepositoryOwner is the remote GitHub organization for the releases
	RepositoryOwner = "redhat-developer"

	// RepositoryName is the remote GitHub repository for the releases
	RepositoryName = "app-services-cli"

	// TermsReviewEventCode is the event code used when checking the terms review
	TermsReviewEventCode = "onlineService"

	// TermsReviewSiteCode is the site code used when checking the terms review
	TermsReviewSiteCode = "ocm"
)

// Auth Build variables
var (
	ProductionAPIURL            = "https://api.openshift.com"
	StagingAPIURL               = "https://api.stage.openshift.com"
	DefaultClientID             = "rhoas-cli-prod"
	DefaultOfflineTokenClientID = "cloud-services"
	ProductionAuthURL           = "https://sso.redhat.com/auth/realms/redhat-external"
	ProductionMasAuthURL        = "https://identity.api.openshift.com/auth/realms/rhoas"
	StagingMasAuthURL           = "https://identity.api.stage.openshift.com/auth/realms/rhoas"
)

func init() {
	if isDevBuild() {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}

// isDevBuild returns true if the current build is "dev" (dev build)
func isDevBuild() bool {
	return Version == "dev"
}