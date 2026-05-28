package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/navid/pst2jmap-migration/internal/jmap"
	"github.com/navid/pst2jmap-migration/internal/pst"
)

const JMAP_URL string = "https://postmaster.collab24.net/jmap"

var version = "dev"

func main() {

	startedAt := time.Now()

	var (
		pstFile  string
		username string
		password string
	)

	flag.StringVar(
		&pstFile,
		"pst",
		"",
		"Src. Path to PST file",
	)

	flag.StringVar(
		&username,
		"user",
		"",
		"Dest. Username",
	)

	flag.StringVar(
		&password,
		"password",
		"",
		"Dest. Password",
	)

	showVersion :=
		flag.Bool(
			"version",
			false,
			"Show version",
		)

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	validate(
		pstFile,
		username,
		password,
	)

	ctx := context.Background()

	fmt.Println("Extracting PST...")

	reader, err := pst.NewReader(
		ctx,
		pstFile,
	)

	if err != nil {
		exit(
			"failed to extract PST",
			err,
		)
	}

	defer reader.Close()

	client := jmap.NewClient(
		JMAP_URL,
		username,
		password,
	)

	fmt.Println("Connecting to JMAP...")

	err = client.Connect()

	if err != nil {
		exit(
			"failed to connect jmap",
			err,
		)
	}

	fmt.Println(
		"Authenticated as:",
		client.Session.Username,
	)

	fmt.Println()

	mailboxes, err := client.GetMailboxIDs()

	if err != nil {
		exit("failed to get mailboxes", err)
	}

	fmt.Println("Detected mailboxes:")

	for role, id := range mailboxes {

		fmt.Printf(
			"  %-10s %s\n",
			role,
			id,
		)
	}

	fmt.Println()

	total, err :=
		pst.CountMessages(
			reader.OutputDir,
		)

	if err != nil {
		exit(
			"failed to count messages",
			err,
		)
	}

	stats :=
		pst.NewStats(
			total,
			startedAt,
		)

	fmt.Printf(
		"Found %d emails\n\n",
		total,
	)

	fmt.Println("Scanning messages...")

	folderCounts := map[string]int{}

	imported := false

	err = reader.Walk(
		func(path string) error {

			if imported {
				return filepath.SkipAll
			}

			msg, err :=
				pst.ParseMessage(path)

			if err != nil {

				fmt.Println(
					"PARSE FAILED:",
					err,
				)

				return nil
			}

			mailboxID :=
				jmap.ResolveMailboxID(
					msg.Folder,
					mailboxes,
				)

			fmt.Println()
			fmt.Println(
				"Testing EML:",
				path,
			)

			fmt.Println(
				"Folder:",
				msg.Folder,
			)

			fmt.Println(
				"Subject:",
				msg.Subject,
			)

			fmt.Println(
				"Mailbox:",
				mailboxID,
			)

			blobID, err :=
				client.UploadEML(
					path,
				)

			if err != nil {

				fmt.Println(
					"UPLOAD FAILED:",
					err,
				)

				return nil
			}

			fmt.Println(
				"Uploaded blob:",
				blobID,
			)

			err =
				client.ImportEmail(
					blobID,
					mailboxID,
				)

			if err != nil {

				fmt.Println(
					"IMPORT FAILED:",
					err,
				)

				return nil
			}

			fmt.Println(
				"Email imported successfully",
			)

			folderCounts[msg.Folder]++

			stats.IncrementImported()

			imported = true

			return filepath.SkipAll
		},
	)

	if err != nil {
		exit(
			"migration failed",
			err,
		)
	}

	stats.Finish()

	fmt.Println()
	fmt.Println("Folder summary:")

	for folder, count := range folderCounts {

		fmt.Printf(
			"  %-15s %d\n",
			folder,
			count,
		)
	}

	fmt.Println()
	fmt.Println(stats)
}

func validate(
	pstFile string,
	user string,
	pass string,
) {

	required := map[string]string{
		"--pst":      pstFile,
		"--user":     user,
		"--password": pass,
	}

	for k, v := range required {

		if v == "" {

			fmt.Fprintf(
				os.Stderr,
				"ERROR: missing required option %s\n",
				k,
			)

			os.Exit(1)
		}
	}
}

func exit(
	msg string,
	err error,
) {

	fmt.Fprintf(
		os.Stderr,
		"ERROR: %s: %v\n",
		msg,
		err,
	)

	os.Exit(1)
}
