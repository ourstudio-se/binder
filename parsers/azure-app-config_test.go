package parsers

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAzureConfigParser_Parse_MultiPageAndKeyVaultRef(t *testing.T) {
	firstKey := "firstKey"
	firstValue := "firstValue"

	keyVaultRefSettingKey := "kvSecretKey"
	secretUriValue := "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets/mySecretName/0123456789abcdef\"}" //nolint
	secretContentType := keyVaultRef

	thirdKey := "thirdKey"
	thirdValue := "thirdValue"

	mockPager := &MockMultiPagePager{
		pages: [][]azappconfig.Setting{
			{
				{Key: &firstKey, Value: &firstValue},
				{Key: &keyVaultRefSettingKey, Value: &secretUriValue, ContentType: &secretContentType},
			},
			{
				{Key: &thirdKey, Value: &thirdValue},
			},
		},
	}

	parser := AzureConfigParser{
		settingsPager: mockPager,
		secretClientFactory: func(url string) (KeyVaultClient, error) {
			return MockSecretClient{secret: "resolvedSecretValue"}, nil
		},
	}

	configValues, err := parser.Parse()
	require.NoError(t, err)

	assert.Equal(t, 3, len(configValues))
	assert.Equal(t, firstValue, configValues[firstKey])
	assert.Equal(t, "resolvedSecretValue", configValues[keyVaultRefSettingKey])
	assert.Equal(t, thirdValue, configValues[thirdKey])
}

type MockMultiPagePager struct {
	current int
	pages   [][]azappconfig.Setting
}

func (m *MockMultiPagePager) More() bool {
	return m.current < len(m.pages)
}

func (m *MockMultiPagePager) NextPage(_ context.Context) (azappconfig.ListSettingsPageResponse, error) {
	if !m.More() {
		return azappconfig.ListSettingsPageResponse{}, fmt.Errorf("no more pages")
	}
	page := m.pages[m.current]
	m.current++
	return azappconfig.ListSettingsPageResponse{Settings: page}, nil
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

func TestAzureConfigParser_resolveKeyVaultSecret(t *testing.T) {
	secretValue := "secretValue"
	type fields struct {
		secretClientFactory func(url string) (KeyVaultClient, error)
	}
	type args struct {
		ctx               context.Context
		keyVaultReference string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Test resolveKeyVaultSecret with valid keyVaultReference",
			fields: fields{
				secretClientFactory: func(url string) (KeyVaultClient, error) {
					return MockSecretClient{
						secret: secretValue,
					}, nil
				},
			},
			args: args{
				ctx:               context.Background(),
				keyVaultReference: "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets/mySecretName/0123456789abcdef\"}",
			},
			want:    secretValue,
			wantErr: assert.NoError,
		},
		{
			name: "Test resolveKeyVaultSecret with invalid keyVaultReference",
			fields: fields{
				secretClientFactory: func(url string) (KeyVaultClient, error) {
					return MockSecretClient{
						secret: secretValue,
					}, nil
				},
			},

			args: args{
				ctx:               context.Background(),
				keyVaultReference: "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets\"}",
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "Test resolveKeyVaultSecret creating key vault client fails",
			fields: fields{
				secretClientFactory: func(url string) (KeyVaultClient, error) {
					return nil, fmt.Errorf("failed to create key vault client")
				},
			},
			args: args{
				ctx:               context.Background(),
				keyVaultReference: "{\"uri\":\"https://mykeyvault.vault.azure.net/secrets/mySecretName/0123456789abcdef\"}",
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &AzureConfigParser{
				secretClientFactory: tt.fields.secretClientFactory,
			}
			got, err := p.resolveKeyVaultSecret(tt.args.ctx, tt.args.keyVaultReference)
			if !tt.wantErr(t, err, fmt.Sprintf("resolveKeyVaultSecret(%v, %v)", tt.args.ctx, tt.args.keyVaultReference)) {
				return
			}
			assert.Equalf(t, tt.want, got, "resolveKeyVaultSecret(%v, %v)", tt.args.ctx, tt.args.keyVaultReference)
		})
	}
}

func Test_getSecretRequestObject(t *testing.T) {
	type args struct {
		kVRef string
	}
	tests := []struct {
		name    string
		args    args
		want    keyVaultRequestObject
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Test parseKeyVaultURI with valid kVRef",
			args: args{
				kVRef: "https://vault-name.vault.azure.net/secrets/secret-name/secret-version",
			},
			want: keyVaultRequestObject{
				vaultURL:      "https://vault-name.vault.azure.net",
				secretName:    "secret-name",
				secretVersion: "secret-version",
			},
			wantErr: assert.NoError,
		},
		{
			name: "Test parseKeyVaultURI with invalid kVRef",
			args: args{
				kVRef: "https://vault-name.vault.azure.net/secrets",
			},
			want:    keyVaultRequestObject{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyVaultURI(tt.args.kVRef)
			if !tt.wantErr(t, err, fmt.Sprintf("parseKeyVaultURI(%v)", tt.args.kVRef)) {
				return
			}
			assert.Equalf(t, tt.want, got, "parseKeyVaultURI(%v)", tt.args.kVRef)
		})
	}
}
