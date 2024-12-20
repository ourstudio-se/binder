package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"strings"
)

const (
	keyVaultRef             = "application/vnd.microsoft.appconfig.keyvaultref+json;charset=utf-8"
	secretLengthWithVersion = 3
)

type AppConfigPager interface {
	More() bool
	NextPage(ctx context.Context) (azappconfig.ListSettingsPageResponse, error)
}

type KeyVaultClient interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

// AzureConfigParser is a configuration parser
// which reads configuration-values from an Azure AppConfig and the backing KeyVault if config values is
// key vault reference.
type AzureConfigParser struct {
	settingsPager       AppConfigPager
	secretClientFactory func(url string) (KeyVaultClient, error)
}

// NewAzureConfigParser returns a new AzureConfigParser.
// A AzureConfigParser reads an Azure AppConfig store and returns a map
// with key/value pairs.
func NewAzureConfigParser(appConfig string, additionallyAllowedTenants []string) (*AzureConfigParser, error) {
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		AdditionallyAllowedTenants: additionallyAllowedTenants,
	})
	if err != nil {
		return nil, err
	}

	client, err := azappconfig.NewClient(appConfig, cred, nil)
	if err != nil {
		return nil, err
	}

	pager := client.NewListSettingsPager(azappconfig.SettingSelector{}, nil)

	keyVaultClientFactoryFunc := func(url string) (KeyVaultClient, error) {
		kvClient, err := azsecrets.NewClient(url, cred, nil)
		if err != nil {
			return nil, err
		}
		return kvClient, nil
	}

	return &AzureConfigParser{pager, keyVaultClientFactoryFunc}, nil
}

// Parse retrieves all settings from AppConfig and resolves any KeyVault references.
// TODO: Parse with context as parameter.
func (p *AzureConfigParser) Parse() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	for p.settingsPager.More() {
		settingsPage, err := p.settingsPager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get next page of pager %w", err)
		}

		for _, setting := range settingsPage.Settings {
			if setting.Key == nil || setting.Value == nil {
				continue
			}

			var finalValue string
			if setting.ContentType != nil && strings.EqualFold(*setting.ContentType, keyVaultRef) {
				secretValue, err := p.resolveKeyVaultSecret(context.Background(), *setting.Value)
				if err != nil {
					return nil, fmt.Errorf("failed to get secret value: %w", err)
				}

				finalValue = secretValue
			} else {
				finalValue = *setting.Value
			}

			settings[*setting.Key] = finalValue
		}
	}

	return settings, nil
}

func (p *AzureConfigParser) resolveKeyVaultSecret(ctx context.Context, keyVaultReference string) (string, error) {
	var kvRef struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal([]byte(keyVaultReference), &kvRef); err != nil {
		return "", fmt.Errorf("failed to parse config value %s value that is Key Vault reference: %w", keyVaultReference, err)
	}

	reqObj, err := parseKeyVaultURI(kvRef.URI)
	if err != nil {
		return "", fmt.Errorf("failed to get secret request object: %w", err)
	}

	keyVaultClient, err := p.secretClientFactory(reqObj.vaultURL)
	if err != nil {
		return "", fmt.Errorf("failed to create KeyVault client for %s: %w", reqObj.vaultURL, err)
	}

	secret, err := fetchSecretValue(ctx, keyVaultClient, reqObj.secretName, reqObj.secretVersion, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %w", err)
	}

	return secret, nil
}

type keyVaultRequestObject struct {
	vaultURL      string
	secretName    string
	secretVersion string
}

// parseKeyVaultURI extracts vault URL, secret name and optionally secret version from a KeyVault URI.
// Example URI: https://myvault.vault.azure.net/secrets/mysecret/myversion
func parseKeyVaultURI(kVRef string) (keyVaultRequestObject, error) {
	parts := strings.Split(strings.TrimPrefix(kVRef, "https://"), "/")

	if len(parts) < 3 {
		return keyVaultRequestObject{}, fmt.Errorf("invalid Key Vault reference: %s", kVRef)
	}

	kv := keyVaultRequestObject{
		vaultURL:   "https://" + parts[0],
		secretName: parts[2],
	}
	if len(parts) > secretLengthWithVersion {
		kv.secretVersion = parts[3]
	}

	return kv, nil
}

// fetchSecretValue retrieves the secret value from KeyVault.
func fetchSecretValue(ctx context.Context, client KeyVaultClient, secretName, secretVersion string, options *azsecrets.GetSecretOptions) (string, error) {
	resp, err := client.GetSecret(ctx, secretName, secretVersion, options)
	if err != nil {
		return "", fmt.Errorf("failed to get secret from Key Vault: %w", err)
	}

	if resp.Value != nil {
		return *resp.Value, nil
	}

	return "", fmt.Errorf("secret with name %s has no value", secretName)
}
