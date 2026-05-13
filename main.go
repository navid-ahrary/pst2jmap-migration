package main

import (
	"flag"
	"fmt"
	"os"

	pstreader "github.com/navid/pst2jmap/internal/pst"
)

var version = "dev"

func main() {

	var (
		pstFile  string
		jmapURL  string
		username string
		password string
	)

	flag.StringVar(&pstFile, "pst", "", "Src. Path to PST file")
	flag.StringVar(&jmapURL, "url", "", "Dest. JMAP mail server (Stalwart) endpoint")
	flag.StringVar(&username, "user", "", "Dest. Username")
	flag.StringVar(&password, "password", "", "Dest. Password")

	showVersion := flag.Bool("version", false, "Show version")

	// Custom help output
	flag.Usage = func() {

		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "pst2jmap - Import Outlook PST files into Stalwart via JMAP\n")
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  pst2jmap [options]\n")
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Fprintf(os.Stderr, "REQUIRED OPTIONS\n")
		fmt.Fprintf(os.Stderr, "  --pst         Src. Path to PST file\n")
		fmt.Fprintf(os.Stderr, "  --url         Dest. JMAP mail server (Stalwart) endpoint\n")
		fmt.Fprintf(os.Stderr, "  --user        Dest. Username\n")
		fmt.Fprintf(os.Stderr, "  --password    Dest. Password\n")
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Fprintf(os.Stderr, "OPTIONAL OPTIONS\n")
		fmt.Fprintf(os.Stderr, "  --version     Show version\n")
		fmt.Fprintf(os.Stderr, "  --help        Show help\n")
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Fprintf(os.Stderr, "EXAMPLES\n")
		fmt.Fprintf(os.Stderr, "  pst2jmap \\\n")
		fmt.Fprintf(os.Stderr, "    --pst ./backup.pst \\\n")
		fmt.Fprintf(os.Stderr, "    --url https://mail.example.com/jmap \\\n")
		fmt.Fprintf(os.Stderr, "    --user admin@example.com \\\n")
		fmt.Fprintf(os.Stderr, "    --password secret\n")
		fmt.Fprintf(os.Stderr, "\n")

		fmt.Fprintf(os.Stderr, "NOTES\n")
		fmt.Fprintf(os.Stderr, "  • Folder structure is preserved\n")
		fmt.Fprintf(os.Stderr, "  • Messages are uploaded using JMAP\n")
		fmt.Fprintf(os.Stderr, "  • PST files are processed incrementally\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if pstFile == "" {
		fmt.Fprintln(os.Stderr, "ERROR: missing required option --pst")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	if jmapURL == "" {
		fmt.Fprintln(os.Stderr, "ERROR: missing required option --url")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	if username == "" {
		fmt.Fprintln(os.Stderr, "ERROR: missing required option --user")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	if password == "" {
		fmt.Fprintln(os.Stderr, "ERROR: missing required option --password")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	reader, err := pstreader.Open(pstFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to open PST file: %v\n", err)
		os.Exit(1)
	}

	defer reader.Close()

	if err := reader.WalkMailboxFolders(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to process PST file: %v\n", err)
		os.Exit(1)
	}
}
