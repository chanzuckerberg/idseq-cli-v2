package czid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var inputExp = regexp.MustCompile(`\.(fasta|fa|fastq|fq)(\.gz)?$`)

func IsInput(path string) bool {
	return inputExp.MatchString(path)
}

var sampleNameExp = regexp.MustCompile(`(_L\d\d\d)?(_R[12]|_R[12]_001)?\.(fasta|fa|fastq|fq)(\.gz)?$`)

func ToSampleName(path string) string {
	return sampleNameExp.ReplaceAllString(filepath.Base(path), "")
}

var r1Exp = regexp.MustCompile(`_R1(_001)?\.(fasta|fa|fastq|fq)(\.gz)?$`)

func IsR1(path string) bool {
	return r1Exp.MatchString(path)
}

var r2Exp = regexp.MustCompile(`_R2(_001)?\.(fasta|fa|fastq|fq)(\.gz)?$`)

func IsR2(path string) bool {
	return r2Exp.MatchString(path)
}

func extractLaneNumber(path string) (int, error) {
	match := sampleNameExp.FindString(path)
	if len(match) < 5 {
		log.Fatalf("path has no lane number %s", path)
	}

	n, err := strconv.Atoi(match[2:5])
	if err != nil {
		return n, fmt.Errorf("path has no lane number %s", path)
	}
	return n, nil
}

func sortByLaneNumber(arr *[]string) {
	sort.Slice(*arr, func(i, j int) bool {
		iN, _ := extractLaneNumber((*arr)[i])
		jN, _ := extractLaneNumber((*arr)[i])
		return iN < jN
	})
}

type SampleFiles struct {
	R1     []string
	R2     []string
	Single []string
}

func SamplesFromDir(directory string, verbose bool) (map[string]SampleFiles, error) {
	pairs := make(map[string]SampleFiles)
	if dir, err := os.Stat(directory); err != nil {
		return pairs, err
	} else if !dir.IsDir() {
		return pairs, fmt.Errorf("path %s must be a directory", directory)
	}

	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if match := IsInput(path); match {
			sampleName := ToSampleName(path)
			sampleFiles := pairs[sampleName]

			if IsR1(path) {
				if len(sampleFiles.Single) != 0 {
					return fmt.Errorf("found R1 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}

				if verbose {
					fmt.Printf("detected R1 sample file for sample: %s at path %s\n", sampleName, path)
				}

				sampleFiles.R1 = append(sampleFiles.R1, path)
			} else if IsR2(path) {
				if len(sampleFiles.Single) != 0 {
					return fmt.Errorf("found R2 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}

				if verbose {
					fmt.Printf("detected R2 sample file for sample: %s at path %s\n", sampleName, path)
				}

				sampleFiles.R2 = append(sampleFiles.R2, path)
			} else {
				if len(sampleFiles.R1) != 0 {
					return fmt.Errorf("found R1 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.R1)
				}
				if len(sampleFiles.R2) != 0 {
					return fmt.Errorf("found R2 file and single end file for sample '%s': %s, %s", sampleName, path, sampleFiles.R2)
				}
				if len(sampleFiles.Single) != 0 {
					return fmt.Errorf("found multiple single end files for sample '%s': %s, %s", sampleName, path, sampleFiles.Single)
				}

				if verbose {
					fmt.Printf("detected single sample file for sample: %s at path %s\n", sampleName, path)
				}

				sampleFiles.Single = append(sampleFiles.Single, path)
			}
			pairs[sampleName] = sampleFiles
		}
		return err
	})
	for sampleName, pair := range pairs {
		if verbose {
			fmt.Printf("detected sample: %s\n", sampleName)
		}
		if len(pair.R1) != len(pair.R2) {
			return pairs, fmt.Errorf("missmatch in R1 and R2 file count for sample name '%s' %d != %d", sampleName, len(pair.R1), len(pair.R2))
		}

		if len(pair.R1) > 1 {
			sortByLaneNumber(&pair.R1)
		}

		if len(pair.R2) > 1 {
			sortByLaneNumber(&pair.R2)
		}

		if len(pair.Single) > 1 {
			sortByLaneNumber(&pair.Single)
		}

		if len(pair.R1) > 0 && len(pair.R2) > 0 {
			laneNumbers := make([]int, len(pair.R1))
			for i, path := range pair.R1 {
				laneNumbers[i], _ = extractLaneNumber(path)
			}

			for i, path := range pair.R2 {
				if n, _ := extractLaneNumber(path); laneNumbers[i] != n {
					return pairs, fmt.Errorf("missmatched lane numbers")
				}
			}
		}
	}
	return pairs, err
}
