// This Dagger configuration allows to create an EKS cluster from scratch and run tests on it then destroy it.

package main

import (
	"context"
	"dagger/cloud-native-ref/internal/dagger"
	"fmt"
	"strings"
)

type CloudNativeRef struct{}

// bootstrapContainer creates a container with the necessary tools to bootstrap the EKS cluster
func bootstrapContainer(env []string) (*dagger.Container, error) {
	// init a wolfi container with the necessary tools
	ctr := dag.Apko().Wolfi([]string{"aws-cli-v2", "bash", "bind-tools", "curl", "git", "jq", "opentofu", "vault"})

	// Add the environment variables to the container
	for _, e := range env {
		parts := strings.Split(e, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid environment variable format, must be in the form <key>:<value>: %s", e)
		}
		ctr = ctr.WithEnvVariable(parts[0], parts[1])
	}

	return ctr, nil
}

// Clean all Terraform cache files
func (m *CloudNativeRef) Clean(
	ctx context.Context,

	// source is the directory where the Terraform configuration is stored
	// +required
	source *dagger.Directory,
) (string, error) {

	ctr, err := bootstrapContainer([]string{})
	if err != nil {
		return "", err
	}

	cmd := []string{"find", ".", "(", "-type", "d", "-name", "*.terraform", "-or", "-name", "*.terraform.lock.hcl", ")", "-exec", "rm", "-vrf", "{}", "+"}
	return ctr.
		WithMountedDirectory("/cloud-native-ref", source).
		WithWorkdir("/cloud-native-ref").
		WithExec(cmd).Stdout(ctx)
}

// Plan display the terraform plan for all the modules
func (m *CloudNativeRef) Plan(
	ctx context.Context,

	// source is the directory where the Terraform configuration is stored
	// +required
	source *dagger.Directory,

	// The directory where the AWS authentication files will be stored
	// +optional
	authDir *dagger.Directory,

	// The AWS IAM Role ARN to assume
	// +optional
	assumeRoleArn string,

	// The AWS profile to use
	// +optional
	profile string,

	// The AWS secret access key
	// +optional
	secretAccessKey *dagger.Secret,

	// The AWS access key ID
	// +optional
	accessKeyID *dagger.Secret,

	// a list of environment variables, expected in (key:value) format
	// +optional
	env []string,

) (string, error) {

	ctr, err := bootstrapContainer(env)
	if err != nil {
		return "", err
	}

	// mount the source directory
	ctr = ctr.WithMountedDirectory("/cloud-native-ref", source)

	// Add the AWS credentials if provided as environment variables
	if accessKeyID != nil && secretAccessKey != nil {
		accessKeyIDValue, err := getSecretValue(ctx, accessKeyID)
		if err != nil {
			return "", err
		}
		secretAccessKeyValue, err := getSecretValue(ctx, secretAccessKey)
		if err != nil {
			return "", err
		}
		ctr = ctr.WithEnvVariable("AWS_ACCESS_KEY_ID", accessKeyIDValue).
			WithEnvVariable("AWS_SECRET_ACCESS_KEY", secretAccessKeyValue)
	}

	if authDir != nil {
		ctr = ctr.WithMountedDirectory("/root/.aws/", authDir)
	}

	createNetwork(ctx, ctr, false)

	return ctr.
		WithExec([]string{"echo", "Bootstrap the EKS cluster"}).
		Stdout(ctx)
}

// Bootstrap the EKS cluster
func (m *CloudNativeRef) Bootstrap(
	ctx context.Context,

	// source is the directory where the Terraform configuration is stored
	// +required
	source *dagger.Directory,

	// The directory where the AWS authentication files will be stored
	// +optional
	authDir *dagger.Directory,

	// The AWS IAM Role ARN to assume
	// +optional
	assumeRoleArn string,

	// The AWS profile to use
	// +optional
	profile string,

	// The AWS secret access key
	// +optional
	secretAccessKey *dagger.Secret,

	// The AWS access key ID
	// +optional
	accessKeyID *dagger.Secret,

	// AWS region to use
	// +optional
	// +default="eu-west-3"
	region string,

	// apply if set to true, the terraform apply will be executed
	// +optional
	apply bool,

	// tsKey is the Tailscale key to use
	// +optional
	tsKey *dagger.Secret,

	// tsTailnet is the Tailscale tailnet to use
	// +optional
	// +default="smainklh@gmail.com"
	tsTailnet string,

	// tsHostname is the Tailscale hostname to use
	// +optional
	// +default="cloud-native-ref"
	tsHostname string,

	// privateDomainName is the private domain name to use
	// +optional
	// +default="priv.cloud.ogenki.io"
	privateDomainName string,

	// vaultAddr is the Vault address to use
	// +optional
	vaultAddr string,

	// vaultSkipVerify is the Vault skip verify to use
	// +optional
	// +default="true"
	vaultSkipVerify string,

	// Tooling applications to enable
	// +optional
	toolingApps []string,

	// Security applications to enable
	// +optional
	// +default=["../base/cert-manager", "../base/external-secrets", "../base/kyverno"]
	securityApps []string,

	// Observability applications to enable
	// +optional
	// +default=["kube-prometheus-stack"]
	observabilityApps []string,

	// branch is the new branch to use for flux configuration
	branch string,

	// a list of environment variables, expected in (key:value) format
	// +optional
	env []string,

) (string, error) {

	ctr, err := bootstrapContainer(env)
	if err != nil {
		return "", err
	}

	// mount the source directory
	ctr = ctr.WithMountedDirectory("/cloud-native-ref", source)

	// // Configure AWS authentication
	// if accessKeyID != nil && secretAccessKey != nil {
	// 	ctr = ctr.WithSecretVariable("AWS_ACCESS_KEY_ID", accessKeyID).
	// 		WithSecretVariable("AWS_SECRET_ACCESS_KEY", secretAccessKey)
	// }
	// if authDir != nil {
	// 	ctr = ctr.WithMountedDirectory("/root/.aws/", authDir)
	// }

	// // Create the network components
	// _, err = createNetwork(ctx, ctr, true)
	// if err != nil {
	// 	return "", err
	// }
	// sess, err := createAWSSession(ctx, region, accessKeyID, secretAccessKey)
	// if err != nil {
	// 	return "", err
	// }
	// tailscaleSvc, err := tailscaleService(ctx, tsKey, tsTailnet, tsHostname)
	// if err != nil {
	// 	return "", err
	// }

	// // Deploy and configure Vault
	// if vaultAddr == "" && privateDomainName != "" {
	// 	vaultAddr = fmt.Sprintf("https://vault.%s:8200", privateDomainName)
	// }
	// vaultOutput, err := createVault(ctx, ctr, true)
	// if err != nil {
	// 	return "", err
	// }
	// vaultRootToken, err := initVault(vaultOutput, sess)
	// if err != nil {
	// 	return "", err
	// }
	// vaultRootTokenSecret := dag.SetSecret("vaultRootToken", vaultRootToken)
	// ctr = ctr.
	// 	WithEnvVariable("VAULT_ADDR", vaultAddr).
	// 	WithEnvVariable("VAULT_SKIP_VERIFY", vaultSkipVerify).
	// 	WithSecretVariable("VAULT_TOKEN", vaultRootTokenSecret).
	// 	WithServiceBinding("tailscale", tailscaleSvc).
	// 	WithEnvVariable("ALL_PROXY", "socks5h://tailscale:1055").
	// 	WithEnvVariable("HTTP_PROXY", "http://tailscale:1055").
	// 	WithEnvVariable("http_proxy", "http://tailscale:1055")
	// err = configureVaultPKI(ctx, ctr, sess, fmt.Sprintf("certificates/%s/root-ca", privateDomainName))
	// if err != nil {
	// 	return "", err
	// }
	// _, err = configureVault(ctx, ctr, true)
	// if err != nil {
	// 	return "", err
	// }

	// Update kustomizations
	dir, err := createKustomization(ctx, ctr, source, branch, "security/mycluster-0", securityApps)
	if err != nil {
		return "", err
	}

	_, err = dir.Directory("security/mycluster-0").Export(ctx, "/tmp/security/mycluster-0")
	if err != nil {
		return "", err
	}

	return ctr.
		Terminal().
		WithExec([]string{"echo", "Bootstrap the EKS cluster"}).
		Stdout(ctx)
}

func (m *CloudNativeRef) UpdateKustomization(
	ctx context.Context,

	// source is the directory where the Terraform configuration is stored
	// +required
	source *dagger.Directory,

	// branch is the new branch to use for flux configuration
	// +required
	branch string,

	// kustPath is the path to the kustomize directory
	// +required
	path string,

	// resources is the list of resources to use
	// +required
	resources []string,

) (*dagger.Directory, error) {

	ctr, err := bootstrapContainer([]string{})
	if err != nil {
		return nil, err
	}

	// mount the source directory
	ctr = ctr.WithMountedDirectory("/cloud-native-ref", source)

	// Update kustomizations
	dir, err := createKustomization(ctx, ctr, source, branch, path, resources)
	if err != nil {
		return nil, err
	}

	return dir, nil
}
