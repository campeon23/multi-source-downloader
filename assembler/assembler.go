// Inside assembler/assembler.go

package assembler

import (
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/campeon23/multi-source-downloader/hasher"
	"github.com/campeon23/multi-source-downloader/logger"
	"github.com/campeon23/multi-source-downloader/manifest"
)

type Assembler struct {
	PartsDir string
	Log	*logger.Logger
}

func NewAssembler(partsDir string, log *logger.Logger) *Assembler {
	return &Assembler{
		PartsDir: partsDir,
		Log: log,
	}
}

func (a *Assembler) AssembleFileFromParts(manifest manifest.DownloadManifest, outFile *os.File, numParts int, rangeSize int, size int, keepParts bool, hasher hasher.Hasher) {
    // Search for all output-* files in the current directory 
	//	to proceed to assemble the final file
	files, err := filepath.Glob(a.PartsDir + "output-*")
	if err != nil {
		a.Log.Fatal("Error: ", err)
	}

	sort.Slice(files, func(i, j int) bool {
		hashI, err := hasher.CalculateSHA256(files[i])
		if err != nil {
			a.Log.Fatal("Calculating hash: ", "error", err.Error())
		}
		hashJ, err := hasher.CalculateSHA256(files[j])
		if err != nil {
			a.Log.Fatal("Calculating hash: ", "error", err.Error())
		}

		// Get the part numbers from the .file_parts_manifest.json file
		numI, numJ := -1, -1
		for _, part := range manifest.DownloadedParts {
			if part.FileHash == hashI {
				numI = part.PartNumber
			}
			if part.FileHash == hashJ {
				numJ = part.PartNumber
			}
		}

		// Compare the part numbers to determine the sorting order
		return numI < numJ
	})


	// Iterate through `files` and read and combine them in the sorted order
	for i, file := range files {
		a.Log.Debugw(
			"Downloaded part", 
			"part file",	i+1,
			"file", 		file,
		) // Print the part being assembled. Debug output
		partFile, err := os.Open(file)
		if err != nil {
			a.Log.Fatal("Error: ", err)
		}

		copied, err := io.Copy(outFile, partFile)
		if err != nil {
			a.Log.Fatal("Error: ", err)
		}

		if i != numParts-1 && copied != int64(rangeSize) {
			a.Log.Fatal("Error: File part not completely copied")
		} else if i == numParts-1 && copied != int64(size)-int64(rangeSize)*int64(numParts-1) {
			a.Log.Fatal("Error: Last file part not completely copied")
		}

		partFile.Close()
		if !keepParts { // If keepParts is false, remove the part file
			// Remove manifest file and leave only the encrypted one
			err = os.Remove(file)
			if err != nil {
				a.Log.Fatal("Removing part file: ", "error", err.Error())
			}
		}
	}

	a.Log.Infow("File downloaded and assembled")
}