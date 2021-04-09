package idseq

import (
	"log"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/idseq-cli-v2/pkg/upload"
)

func UploadSamplesFlow(
	sampleFiles map[string]SampleFiles,
	stringMetadata map[string]string,
	projectName string,
	metadataCSVPath string,
	workflow string,
	technology string,
	wetlabProtocol string,
) error {
	metadata := NewMetadata(stringMetadata)
	projectID, err := GetProjectID(projectName)
	if err != nil {
		log.Fatal(err)
	}

	samplesMetadata := SamplesMetadata{}
	if metadataCSVPath != "" {
		samplesMetadata, err = CSVMetadata(metadataCSVPath)
		if err != nil {
			log.Fatal(err)
		}
		for sampleName := range samplesMetadata {
			if _, hasSampleName := sampleFiles[sampleName]; !hasSampleName {
				delete(samplesMetadata, sampleName)
			}
		}
	}
	for sampleName := range sampleFiles {
		if _, hasMetadata := samplesMetadata[sampleName]; !hasMetadata {
			samplesMetadata[sampleName] = Metadata{}
		}
	}
	for sampleName, m := range samplesMetadata {
		samplesMetadata[sampleName] = m.Fuse(metadata)
	}
	err = GeoSearchSuggestions(&samplesMetadata)
	if err != nil {
		log.Fatal(err)
	}
	err = ValidateSamplesMetadata(projectID, samplesMetadata)
	if err != nil {
		if err.Error() == "metadata validation failed" {
			os.Exit(1)
		}
		log.Fatal(err)
	}

	credentials, samples, err := CreateSamples(
		projectID,
		sampleFiles,
		samplesMetadata,
		workflow,
		technology,
		wetlabProtocol,
	)
	if err != nil {
		log.Fatal(err)
	}

	u := upload.NewUploader(credentials)
	for _, sample := range samples {
		sF := sampleFiles[sample.Name]
		for _, inputFile := range sample.InputFiles {
			filename := ""
			if filepath.Base(sF.R1) == filepath.Base(inputFile.S3Path) {
				filename = sF.R1
			} else if filepath.Base(sF.R2) == filepath.Base(inputFile.S3Path) {
				filename = sF.R2
			} else {
				filename = sF.Single
			}
			err := u.UploadFile(filename, inputFile.S3Path, inputFile.MultipartUploadId)
			if err != nil {
				log.Fatal(err)
			}
		}
		err := MarkSampleUploaded(sample.ID, sample.Name)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}