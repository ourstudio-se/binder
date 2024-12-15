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

// Parse returns the key/value pairs as a map[string]interface{}.
// TODO: Parse with context as parameter.
func (p *AzureConfigParser) Parse() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	for p.settingsPager.More() {
		snapshotPage, err := p.settingsPager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get next page of pager %v", err)
		}

		for _, setting := range snapshotPage.Settings {
			value := ""
			if setting.ContentType != nil && strings.EqualFold(*setting.ContentType, keyVaultRef) {
				secret, err := getSecret(context.Background(), p.secretClientFactory, *setting.Value)
				if err != nil {
					return nil, fmt.Errorf("failed to get secret value: %v", err)
				}

				value = secret
			} else {
				value = *setting.Value
			}

			settings[*setting.Key] = value
		}
	}

	return settings, nil
}

type KeyVaultSecretRequestObject struct {
	vaultURL      string
	secretName    string
	secretVersion string
}

func getSecret(ctx context.Context, secretClientFactory func(url string) (KeyVaultClient, error), keyVaultReference string) (string, error) {
	var kvRef struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal([]byte(keyVaultReference), &kvRef); err != nil {
		return "", fmt.Errorf("failed to parse config value %s value that is Key Vault reference: %v", keyVaultReference, err)
	}

	secretRequestObject, err := getSecretRequestObject(kvRef.URI)
	if err != nil {
		return "", fmt.Errorf("failed to get secret request object: %v", err)
	}

	keyVaultSecretClient, err := secretClientFactory(secretRequestObject.vaultURL)
	if err != nil {
		return "", fmt.Errorf("failed to create key vault client: %v", err)
	}

	secret, err := getSecretValue(ctx, keyVaultSecretClient, secretRequestObject.secretName, secretRequestObject.secretVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get secret value: %v", err)
	}

	return secret, nil
}

func getSecretRequestObject(kVRef string) (KeyVaultSecretRequestObject, error) {
	// https://vault-name.vault.azure.net/secrets/secret-name/secret-version
	parts := strings.Split(strings.TrimPrefix(kVRef, "https://"), "/")

	if len(parts) < 3 {
		return KeyVaultSecretRequestObject{}, fmt.Errorf("invalid Key Vault reference: %s", kVRef)
	}

	kv := KeyVaultSecretRequestObject{
		vaultURL:   "https://" + parts[0],
		secretName: parts[2],
	}
	if len(parts) > secretLengthWithVersion {
		kv.secretVersion = parts[3]
	}

	return kv, nil
}

func getSecretValue(ctx context.Context, client KeyVaultClient, secretName, secretVersion string) (string, error) {
	secretResp, err := client.GetSecret(ctx, secretName, secretVersion, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get secret from Key Vault: %v", err)
	}

	if secretResp.Value != nil {
		return *secretResp.Value, nil
	}

	return "", fmt.Errorf("secret with name %s has no value", secretName)
}
