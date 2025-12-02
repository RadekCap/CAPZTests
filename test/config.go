package test

import (
	"fmt"
	"os"
	"sync"
)

var (
	defaultRepoDir     string
	defaultRepoDirOnce sync.Once
)

// getDefaultRepoDir returns a unique temporary directory path for the repository.
// The path is generated once per test run to ensure consistency across all test phases.
func getDefaultRepoDir() string {
	defaultRepoDirOnce.Do(func() {
		if dir := os.Getenv("ARO_REPO_DIR"); dir != "" {
			defaultRepoDir = dir
			return
		}

		// Create a unique temporary directory path using mktemp pattern
		tmpDir, err := os.MkdirTemp("", "cluster-api-installer-aro-*")
		if err != nil {
			// Fallback to process-based unique name if mktemp fails
			defaultRepoDir = fmt.Sprintf("/tmp/cluster-api-installer-aro-%d", os.Getpid())
			return
		}

		// Remove the directory since git clone will create it
		os.RemoveAll(tmpDir)
		defaultRepoDir = tmpDir
	})

	return defaultRepoDir
}

// TestConfig holds configuration for ARO-CAPZ tests
type TestConfig struct {
	// Repository configuration
	RepoURL    string
	RepoBranch string
	RepoDir    string

	// Cluster configuration
	KindClusterName   string
	ClusterName       string
	ResourceGroup     string
	OpenShiftVersion  string
	Region            string
	AzureSubscription string
	Environment       string
	User              string

	// Paths
	ClusterctlBinPath string
	ScriptsPath       string
	GenScriptPath     string
}

// NewTestConfig creates a new test configuration with defaults
func NewTestConfig() *TestConfig {
	return &TestConfig{
		// Repository defaults
		RepoURL:    GetEnvOrDefault("ARO_REPO_URL", "https://github.com/RadekCap/cluster-api-installer.git"),
		RepoBranch: GetEnvOrDefault("ARO_REPO_BRANCH", "ARO-ASO"),
		RepoDir:    getDefaultRepoDir(),

		// Cluster defaults
		KindClusterName:   GetEnvOrDefault("KIND_CLUSTER_NAME", "capz-stage"),
		ClusterName:       GetEnvOrDefault("CLUSTER_NAME", "test-cluster"),
		ResourceGroup:     GetEnvOrDefault("RESOURCE_GROUP", "test-rg"),
		OpenShiftVersion:  GetEnvOrDefault("OPENSHIFT_VERSION", "4.18"),
		Region:            GetEnvOrDefault("REGION", "uksouth"),
		AzureSubscription: os.Getenv("AZURE_SUBSCRIPTION_NAME"),
		Environment:       GetEnvOrDefault("ENV", "stage"),
		User:              GetEnvOrDefault("USER", os.Getenv("USER")),

		// Paths
		ClusterctlBinPath: GetEnvOrDefault("CLUSTERCTL_BIN", "./bin/clusterctl"),
		ScriptsPath:       GetEnvOrDefault("SCRIPTS_PATH", "./scripts"),
		GenScriptPath:     GetEnvOrDefault("GEN_SCRIPT_PATH", "./doc/aro-hcp-scripts/aro-hcp-gen.sh"),
	}
}

// GetOutputDirName returns the output directory name for generated infrastructure files
func (c *TestConfig) GetOutputDirName() string {
	return fmt.Sprintf("%s-%s", c.ClusterName, c.Environment)
}
