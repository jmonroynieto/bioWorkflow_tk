package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pydpll/errorutils"
	"github.com/urfave/cli/v3"
)

var jsonOutput bool
var fastqReadCount int64

var flagstatFlags []cli.Flag = []cli.Flag{
	&cli.BoolFlag{
		Name:        "json",
		Aliases:     []string{"j"},
		Usage:       "output json blob including original flagstat text",
		Value:       false,
		Destination: &jsonOutput,
	},
	&cli.IntFlag{
		Name:        "fastq read count",
		Aliases:     []string{"e"},
		Usage:       "number of reads in fastq file that are expected to show",
		Required:    false,
		Destination: &fastqReadCount,
	},
}

//missing due to flagstat limitations:
//- mapped reads for each read pair, can be obtained by filtering -F 0x904 and comparing Read1 and Read2 fields to the unfiltered version. That sums up to the total mapped reads that this subcommand does report.

type Flagstat struct {
	Input                 string  `json:"input"`
	Output                string  `json:"output"`
	FastqReadCount        int64   `json:"fastq_read_count"`
	Total                 [2]int  `json:"total"`
	Primary               [2]int  `json:"primary"`
	Secondary             [2]int  `json:"secondary"`
	Supplementary         [2]int  `json:"supplementary"`
	Duplicates            [2]int  `json:"duplicates"`
	PrimaryDuplicates     [2]int  `json:"primary_duplicates"`
	Mapped                [2]int  `json:"mapped"`
	MappedPercent         float64 `json:"mapped_percent"`
	PrimaryMapped         [2]int  `json:"primary_mapped"`
	PrimaryMappedPercent  float64 `json:"primary_mapped_percent"`
	PairedInSeq           [2]int  `json:"paired_in_seq"`
	Read1                 [2]int  `json:"read1"`
	Read2                 [2]int  `json:"read2"`
	ProperlyPaired        [2]int  `json:"properly_paired"`
	ProperlyPairedPercent float64 `json:"properly_paired_percent"`
	pppmp                 float64 `json:"pppmp"` //QC-passing ProperlyPairedPrimaryMappedPercent
	WithMateMapped        [2]int  `json:"with_mate_mapped"`
	Singletons            [2]int  `json:"singletons"`
	SingletonsPercent     float64 `json:"singletons_percent"`
	MateDiffChr           [2]int  `json:"mate_diff_chr"`
	MateDiffChrMapQ5      [2]int  `json:"mate_diff_chr_mapq5"`
}

func scanFlagstat(scanner *bufio.Scanner) (*Flagstat, error) {
	var flagstat Flagstat

	// Regex patterns for parsing each line.
	patterns := map[string]*regexp.Regexp{
		"total":            regexp.MustCompile(`^(\d+) \+ (\d+) in total`),
		"primary":          regexp.MustCompile(`^(\d+) \+ (\d+) primary$`),
		"secondary":        regexp.MustCompile(`^(\d+) \+ (\d+) secondary`),
		"supplementary":    regexp.MustCompile(`^(\d+) \+ (\d+) supplementary`),
		"duplicates":       regexp.MustCompile(`^(\d+) \+ (\d+) duplicates`),
		"primaryDup":       regexp.MustCompile(`^(\d+) \+ (\d+) primary duplicates`),
		"mapped":           regexp.MustCompile(`^(\d+) \+ (\d+) mapped \((\d+\.\d+)%`),
		"primaryMapped":    regexp.MustCompile(`^(\d+) \+ (\d+) primary mapped \((\d+\.\d+)%`),
		"pairedInSeq":      regexp.MustCompile(`^(\d+) \+ (\d+) paired in sequencing`),
		"read1":            regexp.MustCompile(`^(\d+) \+ (\d+) read1`),
		"read2":            regexp.MustCompile(`^(\d+) \+ (\d+) read2`),
		"properlyPaired":   regexp.MustCompile(`^(\d+) \+ (\d+) properly paired \((\d+\.\d+)%`),
		"withMateMapped":   regexp.MustCompile(`^(\d+) \+ (\d+) with itself and mate mapped`),
		"singletons":       regexp.MustCompile(`^(\d+) \+ (\d+) singletons \((\d+\.\d+)%`),
		"mateDiffChr":      regexp.MustCompile(`^(\d+) \+ (\d+) with mate mapped to a different chr$`),
		"mateDiffChrMapQ5": regexp.MustCompile(`^(\d+) \+ (\d+) with mate mapped to a different chr \(mapQ>=5\)`),
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		flagstat.Input += line + "\n"
		var sentinel error

	switchTry2:
		switch {
		case patterns["total"].MatchString(line):
			flagstat.Total, _, sentinel = parseFlagstatLine(line, patterns["total"])
		case patterns["primary"].MatchString(line):
			flagstat.Primary, _, sentinel = parseFlagstatLine(line, patterns["primary"])
		case patterns["secondary"].MatchString(line):
			flagstat.Secondary, _, sentinel = parseFlagstatLine(line, patterns["secondary"])
		case patterns["supplementary"].MatchString(line):
			flagstat.Supplementary, _, sentinel = parseFlagstatLine(line, patterns["supplementary"])
		case patterns["duplicates"].MatchString(line):
			flagstat.Duplicates, _, sentinel = parseFlagstatLine(line, patterns["duplicates"])
		case patterns["primaryDup"].MatchString(line):
			flagstat.PrimaryDuplicates, _, sentinel = parseFlagstatLine(line, patterns["primaryDup"])
		case patterns["mapped"].MatchString(line):
			flagstat.Mapped, flagstat.MappedPercent, sentinel = parseFlagstatLine(line, patterns["mapped"])
		case patterns["primaryMapped"].MatchString(line):
			flagstat.PrimaryMapped, flagstat.PrimaryMappedPercent, sentinel = parseFlagstatLine(line, patterns["primaryMapped"])
		case patterns["pairedInSeq"].MatchString(line):
			flagstat.PairedInSeq, _, sentinel = parseFlagstatLine(line, patterns["pairedInSeq"])
		case patterns["read1"].MatchString(line):
			flagstat.Read1, _, sentinel = parseFlagstatLine(line, patterns["read1"])
		case patterns["read2"].MatchString(line):
			flagstat.Read2, _, sentinel = parseFlagstatLine(line, patterns["read2"])
		case patterns["properlyPaired"].MatchString(line):
			flagstat.ProperlyPaired, flagstat.ProperlyPairedPercent, sentinel = parseFlagstatLine(line, patterns["properlyPaired"])
		case patterns["withMateMapped"].MatchString(line):
			flagstat.WithMateMapped, _, sentinel = parseFlagstatLine(line, patterns["withMateMapped"])
		case patterns["singletons"].MatchString(line):
			flagstat.Singletons, flagstat.SingletonsPercent, sentinel = parseFlagstatLine(line, patterns["singletons"])
		case patterns["mateDiffChr"].MatchString(line):
			flagstat.MateDiffChr, _, sentinel = parseFlagstatLine(line, patterns["mateDiffChr"])
		case patterns["mateDiffChrMapQ5"].MatchString(line):
			flagstat.MateDiffChrMapQ5, _, sentinel = parseFlagstatLine(line, patterns["mateDiffChrMapQ5"])
		default:
			errorutils.WarnOnFail(fmt.Errorf("unrecognized flagstat line: %q", line))
			goto switchTry2
		}
		errorutils.ExitOnFail(sentinel)
		err := scanner.Err()
		errorutils.ExitOnFail(err, errorutils.WithMsg("scanner error on input: "+line))
		flagstat.pppmp = float64(flagstat.ProperlyPaired[0]) / float64(flagstat.PrimaryMapped[0]) * 100
	}
	return &flagstat, nil
}

func parseFlagstatLine(line string, pattern *regexp.Regexp) ([2]int, float64, error) {
	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return [2]int{}, 0, fmt.Errorf("line does not match expected format: %q", line)
	}

	values := [2]int{}
	var err error
	values[0], err = strconv.Atoi(matches[1])
	if err != nil {
		return [2]int{}, 0, fmt.Errorf("failed to parse first integer value: %v", err)
	}

	values[1], err = strconv.Atoi(matches[2])
	if err != nil {
		return [2]int{}, 0, fmt.Errorf("failed to parse second integer value: %v", err)
	}

	var percent float64
	if len(matches) > 3 && matches[3] != "" {
		percent, err = strconv.ParseFloat(matches[3], 64)
		if err != nil {
			return [2]int{}, 0, fmt.Errorf("failed to parse percentage: %v", err)
		}
	}

	return values, percent, nil
}

// report generation
func report(flagstat *Flagstat, jsonOutput bool) {
	if !jsonOutput {
		flagstat.generateReport()
		return
	}
	//marshall to json
	flagstatJson, err := json.Marshal(flagstat)
	errorutils.ExitOnFail(err, errorutils.WithMsg("failed to marshal struct to json"))
	flagstat.Output = string(flagstatJson)
	return
}

func (f *Flagstat) generateReport() {
	result := "Report\n"
	if f.FastqReadCount > 0 {
		tp := f.Primary[0] + f.Primary[1]
		if tp != int(f.FastqReadCount) {
			result += fmt.Sprintf("Read count mismatch between user provided fastq read count and flagstat: %d != %d. ", f.FastqReadCount, tp)
		} else {
			result += "all reads accounted for. "
		}
	}
	result += fmt.Sprintf("QC fail fraction: %.2f%%. Only passing used for counts in this report.\n", float64(f.Total[1])/float64(f.Total[0])*100)
	result += fmt.Sprintf("%.2f%% reads (%d) are primarily mapped of which %.2f%% (%d) are aligned and spaced as expected;. %d additional secondary alignments have been recorded.\n", f.PrimaryMappedPercent, f.PrimaryMapped[0], f.pppmp, f.ProperlyPaired[0], f.Secondary[0]+f.Secondary[1])
	ur := f.PairedInSeq[0] + f.PairedInSeq[1] - f.PrimaryMapped[0] - f.PrimaryMapped[1] //unmapped reads
	result += fmt.Sprintf("%d unmapped include %d singletons\n", ur, f.Singletons[0])
	prc_HQsnv := float64(f.MateDiffChrMapQ5[0]) / float64(f.MateDiffChr[0]) * 100
	result += fmt.Sprintf("Structural variation evidence in %d reads where mates mapped to a different chr of which %.2f%% (%d) are high quality mappings. %d Supplementary mappings could also indicate SV\n", f.MateDiffChr[0], prc_HQsnv, f.MateDiffChrMapQ5[0], f.Supplementary[0])
	f.Output = result
}
