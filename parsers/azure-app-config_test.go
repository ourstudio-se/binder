package parsers

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/assert"
	"testing"
)

var numberOfFetches = 0

func (m MockAppConfigPager) More() bool {
	if numberOfFetches < m.availablePages {
		numberOfFetches++
		return true
	}

	return false
}

func (m MockAppConfigPager) NextPage(_ context.Context) (azappconfig.ListSettingsPageResponse, error) {
	return azappconfig.ListSettingsPageResponse{
		Settings:  m.settings,
		SyncToken: "",
	}, nil
}

type MockAppConfigPager struct {
	currentPage    int
	availablePages int
	settings       []azappconfig.Setting
}

type MockSecretClient struct {
	secret string
}

func (m MockSecretClient) GetSecret(_ context.Context, _ string, _ string, _ *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{
			Value: &m.secret,
		},
	}, nil
}

func Test_Azure_App_Config_Parse(t *testing.T) {
	firstKey := "firstKey"
	firstValue := "firstValue"
	secondKey := "secondKey"
	secondValue := "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets/mySecretName/0123456789abcdef\"}"
	secretContentType := keyVaultRef
	mock := MockAppConfigPager{
		availablePages: 1,
		settings: []azappconfig.Setting{
			{
				Key:   &firstKey,
				Value: &firstValue,
			},
			{
				Key:         &secondKey,
				Value:       &secondValue,
				ContentType: &secretContentType,
			},
		},
	}
	mockParser := AzureConfigParser{
		settingsPager: mock,
		secretClientFactory: func(url string) (KeyVaultClient, error) {
			return MockSecretClient{
				secret: "secretValue",
			}, nil
		},
	}

	numberOfFetches = 0
	configValues, err := mockParser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(configValues))
	assert.Equal(t, firstValue, configValues[firstKey])
	assert.Equal(t, "secretValue", configValues[secondKey])
}

func Test_getSecretRequestObject(t *testing.T) {
	type args struct {
		kVRef string
	}
	tests := []struct {
		name    string
		args    args
		want    KeyVaultSecretRequestObject
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Test getSecretRequestObject with valid kVRef",
			args: args{
				kVRef: "https://vault-name.vault.azure.net/secrets/secret-name/secret-version",
			},
			want: KeyVaultSecretRequestObject{
				vaultURL:      "https://vault-name.vault.azure.net",
				secretName:    "secret-name",
				secretVersion: "secret-version",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSecretRequestObject(tt.args.kVRef)
			if !tt.wantErr(t, err, fmt.Sprintf("getSecretRequestObject(%v)", tt.args.kVRef)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getSecretRequestObject(%v)", tt.args.kVRef)
		})
	}
}

func Test_getSecret(t *testing.T) {
	secretValue := "secretValue"

	type args struct {
		secretClientFactory func(url string) (KeyVaultClient, error)
		keyVaultReference   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Test getSecret with valid keyVaultReference",
			args: args{
				secretClientFactory: func(url string) (KeyVaultClient, error) {
					return MockSecretClient{
						secret: secretValue,
					}, nil
				},
				keyVaultReference: "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets/mySecretName/0123456789abcdef\"}",
			},
			want:    secretValue,
			wantErr: assert.NoError,
		},
		{
			name: "Test getSecret with invalid keyVaultReference",
			args: args{
				secretClientFactory: func(url string) (KeyVaultClient, error) {
					return MockSecretClient{
						secret: secretValue,
					}, nil
				},
				keyVaultReference: "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets\"}",
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "Test getSecret when creating key vault client fails",
			args: args{
				secretClientFactory: func(url string) (KeyVaultClient, error) {
					return nil, fmt.Errorf("failed to create key vault client")
				},
				keyVaultReference: "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets/mySecretName/0123456789abcdef\"}",
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSecret(context.Background(), tt.args.secretClientFactory, tt.args.keyVaultReference)
			if !tt.wantErr(t, err, fmt.Sprintf("getSecret %s", tt.name)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getSecret(%v, %v)", tt.args.secretClientFactory, tt.args.keyVaultReference)
		})
	}
}
