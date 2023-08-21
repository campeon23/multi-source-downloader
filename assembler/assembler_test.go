package assembler

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/campeon23/multi-source-downloader/hasher"
	"github.com/campeon23/multi-source-downloader/logger"
	"github.com/campeon23/multi-source-downloader/manifest"
	"github.com/stretchr/testify/assert"
)

// Mock logger for our tests
type MockLogger struct {
	*logger.Logger  // Embedding
}

func (l *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	// Your mock implementation. For instance, just print them.
	fmt.Println("Mocked Infow:", msg, keysAndValues)
}
func (l *MockLogger) Debugw(msg string, keysAndValues ...interface{}) {}
func (l *MockLogger) Fatalw(msg string, keysAndValues ...interface{}) {}

func TestAssembleFileFromParts(t *testing.T) {
	partsDir := "test_data_tmp"
	prefixParts := "part_"

	l := logger.InitLogger(true)
	a := NewAssembler(3, partsDir, false, prefixParts, l)
	h := hasher.NewHasher(partsDir, prefixParts, l)

	currentDir, err := os.Getwd()
	assert.NoErrorf(t, err, "Failed to get current dir: %v", err)

	a.PartsDir = path.Join(currentDir, partsDir)

	// Mock data for test
	m := manifest.DownloadManifest{
		DownloadedParts: []manifest.DownloadedPart{
			{FileHash: "hash1", PartNumber: 1, Timestamp: 0, PartFile: path.Join(a.PartsDir, prefixParts) + strconv.Itoa(1)},
			{FileHash: "hash2", PartNumber: 2, Timestamp: 0, PartFile: path.Join(a.PartsDir, prefixParts) + strconv.Itoa(2)},
			{FileHash: "hash3", PartNumber: 3, Timestamp: 0, PartFile: path.Join(a.PartsDir, prefixParts) + strconv.Itoa(3)},
		},
	}

	l.Debugf("Manifest: %v", m)

	err = os.Mkdir(partsDir, 0755)
	assert.NoErrorf(t, err, "Failed to create test directory: %v", err)
	defer os.RemoveAll(partsDir) // Cleanup

	// Create mock parts
	for i := range m.DownloadedParts {
		content := "content" + strconv.Itoa(i+1)
		err = os.WriteFile(path.Join(a.PartsDir, prefixParts) + strconv.Itoa(i+1), []byte(content), 0644)
		if err != nil {
			assert.NoErrorf(t, err, "Failed to create mock part file: %v", err)
		}
		m.DownloadedParts[i].FileHash, err = h.CalculateSHA256(path.Join(a.PartsDir, prefixParts) + strconv.Itoa(i+1))
		assert.NoErrorf(t, err, "Failed to calculate hash: %v", err)
	}

	outFile, _ := os.CreateTemp(a.PartsDir, "out_")
	defer os.Remove(outFile.Name())

	err = a.AssembleFileFromParts(m, outFile, 0, 0, hasher.Hasher{})
	// l.Infow(err.Error())
	assert.NoErrorf(t, err, "Failed to assemble file from parts: %v", err)
	resultContent, _ := os.ReadFile(outFile.Name())
	expectedContent := "content1content2content3"
	assert.Equal(t, expectedContent, string(resultContent))
}

func TestPrepareAssemblyEnvironment(t *testing.T) {
	partsDir := "./testdata"
	prefixParts := "part_"
	l := logger.InitLogger(true)
	a := NewAssembler(3, "./testdata", false, "part_", l)

	// Mock manifest content
	manifestContent := []byte(`{"partsDir": "./testdata", "prefixParts": "part_"}`)

	_, outFile, _, err := a.PrepareAssemblyEnviroment("./testdata/testfile", manifestContent)
	defer os.Remove(outFile.Name())

	assert.NoErrorf(t, err, "Failed to prepare assembly environment: %v", err)
	assert.Equal(t, "./testdata", partsDir)
	assert.Equal(t, "part_", prefixParts)
}

// Additional tests can be written to simulate errors and check edge cases.
