package fabric

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/internal/utils"
	"github.com/hyperledger/firefly-cli/pkg/types"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNewFabricProvider(t *testing.T) {
	Stack := &types.Stack{
		Name:                   "FabricUser",
		Members:                []*types.Organization{{OrgName: "Hyperledger-fabric"}},
		BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "fabric"),
		BlockchainConnector:    fftypes.FFEnumValue("BlockchainConnector", "fabonnect"),
		BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "fabric"),
	}
	ctx := log.WithLogger(context.Background(), &log.StdoutLogger{})
	fabricProvider := NewFabricProvider(ctx, Stack)
	assert.NotNil(t, fabricProvider)
	assert.NotNil(t, fabricProvider.ctx)
	assert.NotNil(t, fabricProvider.stack)
	assert.NotNil(t, fabricProvider.log)
}

func TestGetOrgConfig(t *testing.T) {
	testcases := []struct {
		Name      string
		Stack     *types.Stack
		OrgConfig *types.OrgConfig
		Member    *types.Organization
	}{
		{
			Name:  "TestFabric-1",
			Stack: &types.Stack{Name: "fabric_user-1"},
			Member: &types.Organization{
				OrgName:  "Org-1",
				NodeName: "fabric",
				Account: &ethereum.Account{
					Address:    "0x1234567890abcdef0123456789abcdef6789abcd",
					PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-1",
				Key:  "Org-1",
			},
		},
		{
			Name:  "TestFabric-2",
			Stack: &types.Stack{Name: "fabri_user-2"},
			Member: &types.Organization{
				OrgName:  "Org-2",
				NodeName: "besu",
				Account: &ethereum.Account{
					Address:    "0x1f2a000000000000000000000000000000000000",
					PrivateKey: "9876543210987654321098765432109876543210987654321098765432109876",
				},
			},
			OrgConfig: &types.OrgConfig{
				Name: "Org-2",
				Key:  "Org-2",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &FabricProvider{}
			orgConfig := p.GetOrgConfig(tc.Stack, tc.Member)
			assert.NotNil(t, orgConfig)
			assert.Equal(t, tc.OrgConfig, orgConfig)
		})

	}
}

func TestGetFabconnetServiceDefinitions(t *testing.T) {
	stack := &types.Stack{
		Name:       "fabric_user-1",
		RuntimeDir: "mock_runtime_dir",
		VersionManifest: &types.VersionManifest{
			Fabconnect: &types.ManifestEntry{
				Image: "fabric-apline",
				Local: true,
			},
		},
		RemoteFabricNetwork: true,
	}
	p := &FabricProvider{
		stack: stack,
	}
	// Create mock organizations
	members := []*types.Organization{
		{
			ID:                   "org1",
			ExposedConnectorPort: 4000,
		},
		{
			ID:                   "org2",
			ExposedConnectorPort: 4001,
		},
		{
			ID:                   "org3",
			ExposedConnectorPort: 4002,
		},
		{
			ID:                   "org4",
			ExposedConnectorPort: 4003,
		},
		{
			ID:                   "org5",
			ExposedConnectorPort: 4004,
		},
	}
	serviceDefinitions := p.getFabconnectServiceDefinitions(members)
	assert.NotNil(t, serviceDefinitions)
}

func TestParseAccount(t *testing.T) {
	input := map[string]interface{}{
		"name":    "user-1",
		"orgName": "hyperledger",
	}

	besuProvider := &FabricProvider{}
	result := besuProvider.ParseAccount(input)

	if _, ok := result.(*Account); !ok {
		t.Errorf("Expected result to be of type *ethereum.Account, but got %T", result)
	}
	expectedAccount := &Account{
		Name:    "user-1",
		OrgName: "hyperledger",
	}
	assert.Equal(t, expectedAccount, result, "Generated result unmatched")

}

func TestGetConnectorName(t *testing.T) {
	testString := "fabconnect"
	p := &FabricProvider{}
	connector := p.GetConnectorName()
	assert.NotNil(t, connector)
	assert.Equal(t, testString, connector)
}

func TestGetConnectorExternalURL(t *testing.T) {
	testCases := []struct {
		Name        string
		Org         *types.Organization
		ExpectedURL string
	}{
		{
			Name: "testcase-1",
			Org: &types.Organization{
				ID:       "user-1",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Nicko",
					OrgName: "hyperledger",
				},
				ExposedConnectorPort: 8900,
			},
			ExpectedURL: "http://127.0.0.1:8900",
		},
		{
			Name: "testcase-2",
			Org: &types.Organization{
				ID:       "user-2",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Richardson",
					OrgName: "hyperledger",
				},
				ExposedConnectorPort: 3000,
			},
			ExpectedURL: "http://127.0.0.1:3000",
		},
		{
			Name: "testcase-3",
			Org: &types.Organization{
				ID:       "user-3",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Philip",
					OrgName: "hyperledger",
				},
				ExposedConnectorPort: 4005,
			},
			ExpectedURL: "http://127.0.0.1:4005",
		},
	}
	for _, tc := range testCases {
		p := &FabricProvider{}
		ExternalURL := p.GetConnectorExternalURL(tc.Org)
		assert.Equal(t, tc.ExpectedURL, ExternalURL)
	}
}

func TestGetConnectorURL(t *testing.T) {
	testCases := []struct {
		Name        string
		Org         *types.Organization
		ExpectedURL string
	}{
		{
			Name: "testcase-1",
			Org: &types.Organization{
				ID:       "user-1",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Nicko",
					OrgName: "hyperledger",
				},
			},
			ExpectedURL: "http://fabconnect_user-1:3000",
		},
		{
			Name: "testcase-2",
			Org: &types.Organization{
				ID:       "user-2",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Richardson",
					OrgName: "hyperledger",
				},
			},
			ExpectedURL: "http://fabconnect_user-2:3000",
		},
		{
			Name: "testcase-3",
			Org: &types.Organization{
				ID:       "user-3",
				NodeName: "fabric",
				Account: &Account{
					Name:    "Philip",
					OrgName: "hyperledger",
				},
			},
			ExpectedURL: "http://fabconnect_user-3:3000",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			p := &FabricProvider{}
			URL := p.GetConnectorURL(tc.Org)
			assert.Equal(t, tc.ExpectedURL, URL)
		})
	}

}

func TestGetContracts(t *testing.T) {
	FilePath := t.TempDir()
	testContractFile := filepath.Join(FilePath, "/test_contracts.json")
	// Sample contract JSON content for testing
	const testContractJSON = `{
			"contracts": {
				"Contract1": {
					"name": "fabric_1",
					"abi": "fabric_abi_1",
					"bin": "sample_bin_1"
				},
				"Contract2": {
					"name": "fabric_2",
					"abi": "fabric_abi_2",
					"bin": "fabric_bin_2"
				}
			}
		}`
	p := &FabricProvider{}

	err := os.WriteFile(testContractFile, []byte(testContractJSON), 0755)
	if err != nil {
		t.Log("unable to write file:", err)
	}
	contracts, err := p.GetContracts(testContractFile, nil)
	if err != nil {
		t.Log("unable to get contract", err)
	}
	assert.NotNil(t, contracts)
}

func TestCreateAccount(t *testing.T) {
	testAccounts := []struct {
		Name  string
		Stack *types.Stack
		Args  []string
	}{
		{
			Name: "TestAccount-1",
			Args: []string{},
			Stack: &types.Stack{
				Name:                   "user-1",
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "fabric"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockChainConnector", "fabric"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "fabric"),
				Members: []*types.Organization{
					{
						ID:                   "org1",
						OrgName:              "hyperledger",
						ExposedConnectorPort: 4000,
					},
				},
			},
		},
		{
			Name: "TestAccount-2",
			Args: []string{},
			Stack: &types.Stack{
				Name:                   "user-2",
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "fabric"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockChainConnector", "fabric"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "fabric"),
				Members: []*types.Organization{
					{
						ID:                   "org1",
						OrgName:              "solana",
						ExposedConnectorPort: 4001,
					},
				},
			},
		},

		{
			Name: "TestAccount-3",
			Args: []string{},
			Stack: &types.Stack{
				Name:                   "user-3",
				BlockchainProvider:     fftypes.FFEnumValue("BlockchainProvider", "fabric"),
				BlockchainConnector:    fftypes.FFEnumValue("BlockChainConnector", "fabric"),
				BlockchainNodeProvider: fftypes.FFEnumValue("BlockchainNodeProvider", "fabric"),
				Members: []*types.Organization{
					{
						ID:                   "org1",
						OrgName:              "ethereum",
						ExposedConnectorPort: 4002,
					},
				},
			},
		},
	}
	for _, tc := range testAccounts {
		p := &FabricProvider{
			stack: tc.Stack,
		}
		Account, err := p.CreateAccount(tc.Args)
		if err != nil {
			t.Log("unable to create account", err)
		}
		assert.NotNil(t, Account)
	}
}

func TestRegisterIdentity(t *testing.T) {
	t.Run("register", func(t *testing.T) {
		utils.StartMockServer(t)

		Member := &types.Organization{
			ID:                   "fabric_user-1",
			NodeName:             "fabric",
			Account:              &Account{Name: "Nicko", OrgName: "hyperledger"},
			ExposedConnectorPort: 3000,
		}
		createIdentityURL := fmt.Sprintf("http://127.0.0.1:%v/identities", Member.ExposedConnectorPort)
		enrollIdentityURL := fmt.Sprintf("http://127.0.0.1:%v/identities/Nicko/enroll", Member.ExposedConnectorPort)

		IdentityName := "Nicko"
		createdApiResponse := `
		{
			"Name": "Nicko",
			"Secret": "9876543210987654321098765432109876543210987654321098765432109876"
		}`
		enrolledApiResponse := `
		{
			"Name": "Nicko",
			"OrgName": "hyperledger"
		}`

		httpmock.RegisterResponder("POST", createIdentityURL,
			httpmock.NewStringResponder(200, createdApiResponse))

		httpmock.RegisterResponder("POST", enrollIdentityURL,
			httpmock.NewStringResponder(200, enrolledApiResponse))

		p := &FabricProvider{}

		account, err := p.registerIdentity(Member, IdentityName)
		if err != nil {
			t.Log("cannot register identity:", err)
		}
		assert.NotNil(t, account)
		assert.NotNil(t, account.Name)
		assert.NotNil(t, account.OrgName)
	})

}

func TestValidateChaincodePackage(t *testing.T) {
	p := &FabricProvider{}

	// Helper function to create a valid tar.gz file
	createValidTarGz := func(t *testing.T, filename string) {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		tarWriter := tar.NewWriter(gzWriter)

		// Add a file to the tar archive
		content := []byte("test chaincode content")
		header := &tar.Header{
			Name: "metadata.json",
			Mode: 0644,
			Size: int64(len(content)),
		}
		err := tarWriter.WriteHeader(header)
		assert.NoError(t, err)
		_, err = tarWriter.Write(content)
		assert.NoError(t, err)

		err = tarWriter.Close()
		assert.NoError(t, err)
		err = gzWriter.Close()
		assert.NoError(t, err)

		err = os.WriteFile(filename, buf.Bytes(), 0644)
		assert.NoError(t, err)
	}

	t.Run("valid tar.gz file", func(t *testing.T) {
		tmpDir := t.TempDir()
		validTarGzFile := filepath.Join(tmpDir, "valid.tar.gz")
		createValidTarGz(t, validTarGzFile)

		err := p.validateChaincodePackage(validTarGzFile)
		assert.NoError(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		err := p.validateChaincodePackage("/nonexistent/path/chaincode.tar.gz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chaincode package file not found")
	})

	t.Run("path is a directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := p.validateChaincodePackage(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chaincode package path is a directory")
	})

	t.Run("invalid - plain text file", func(t *testing.T) {
		tmpDir := t.TempDir()
		plainTextFile := filepath.Join(tmpDir, "plain.txt")
		err := os.WriteFile(plainTextFile, []byte("this is not a gzip file"), 0644)
		assert.NoError(t, err)

		err = p.validateChaincodePackage(plainTextFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chaincode package file format")
		assert.Contains(t, err.Error(), "does not appear to be a valid gzip file")
	})

	t.Run("invalid - zip file", func(t *testing.T) {
		tmpDir := t.TempDir()
		// ZIP files start with PK (0x50, 0x4B)
		zipLikeFile := filepath.Join(tmpDir, "fake.zip")
		err := os.WriteFile(zipLikeFile, []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00}, 0644)
		assert.NoError(t, err)

		err = p.validateChaincodePackage(zipLikeFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chaincode package file format")
	})

	t.Run("invalid - random binary", func(t *testing.T) {
		tmpDir := t.TempDir()
		randomFile := filepath.Join(tmpDir, "random.bin")
		err := os.WriteFile(randomFile, []byte{0xDE, 0xAD, 0xBE, 0xEF}, 0644)
		assert.NoError(t, err)

		err = p.validateChaincodePackage(randomFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chaincode package file format")
	})

	t.Run("invalid - gzip but not tar (plain gzipped content)", func(t *testing.T) {
		tmpDir := t.TempDir()
		gzipOnlyFile := filepath.Join(tmpDir, "gzip_only.gz")

		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err := gzWriter.Write([]byte("this is gzipped but not a tar archive"))
		assert.NoError(t, err)
		err = gzWriter.Close()
		assert.NoError(t, err)

		err = os.WriteFile(gzipOnlyFile, buf.Bytes(), 0644)
		assert.NoError(t, err)

		err = p.validateChaincodePackage(gzipOnlyFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not contain a valid tar archive")
	})

	t.Run("invalid - empty tar.gz", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyTarGzFile := filepath.Join(tmpDir, "empty.tar.gz")

		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		tarWriter := tar.NewWriter(gzWriter)
		// Close without adding any files
		err := tarWriter.Close()
		assert.NoError(t, err)
		err = gzWriter.Close()
		assert.NoError(t, err)

		err = os.WriteFile(emptyTarGzFile, buf.Bytes(), 0644)
		assert.NoError(t, err)

		err = p.validateChaincodePackage(emptyTarGzFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tar.gz archive is empty")
	})

	t.Run("invalid - empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyFile := filepath.Join(tmpDir, "empty.tar.gz")
		err := os.WriteFile(emptyFile, []byte{}, 0644)
		assert.NoError(t, err)

		err = p.validateChaincodePackage(emptyFile)
		assert.Error(t, err)
	})
}
