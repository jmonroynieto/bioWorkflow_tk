package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/pydpll/errorutils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var (
	Version  = "0.0"
	CommitId = ""
)

func main() {

	app := &cli.Command{
		Name:     "saTherapist",
		Usage:    "provide useful information about alignments",
		Flags:    appFlags,
		Commands: appCmds,
		Version:  fmt.Sprintf("%s - %s", Version, CommitId),
	}

	err := app.Run(context.Background(), os.Args)
	errorutils.ExitOnFail(err)
}

var appFlags []cli.Flag = []cli.Flag{
	&cli.BoolFlag{
		Name:    "debug",
		Aliases: []string{"d"},
		Usage:   "activates debugging messages",
		Action: func(ctx context.Context, cmd *cli.Command, shouldDebug bool) error {
			if shouldDebug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil

		},
		Hidden: true,
	},
}

var appCmds []*cli.Command = []*cli.Command{
	{
		Name:   "parseflagstat",
		Usage:  "parse `samtools flagstat` text",
		Action: parseFlagstat,
		Flags:  flagstatFlags,
	},
	//	{
	//		Name:   "id",
	//		Usage:  "generate random id",
	//		Action: generateID,
	//		Flags:  idFlags,
	//	},
}

func parseFlagstat(ctx context.Context, cmd *cli.Command) error {
	var scanner *bufio.Scanner
	if cmd.Args().Len() >= 1 && cmd.Args().First() != "-" {
		file, err := os.OpenFile(cmd.Args().First(), os.O_RDONLY, 0)
		errorutils.ExitOnFail(err, errorutils.WithMsg("failed to open file: "+cmd.Args().First()))
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	flagstat, err := scanFlagstat(scanner)
	report(flagstat, jsonOutput)
	fmt.Printf("%s", flagstat.Output)
	errorutils.ExitOnFail(err)

	//	fmt.Printf("Parsed Flagstat: %+v\n", flagstat)
	return nil
}
