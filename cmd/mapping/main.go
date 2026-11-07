package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"community-mapping/pkg/mapping"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	masterDir := filepath.Join(workDir, "master")
	generatedDir := filepath.Join(workDir, "generated")

	switch os.Args[1] {
	case "build":
		start := time.Now()
		fmt.Println("Compiling master datasets into generated/...")
		if err := mapping.Build(mapping.BuildOptions{
			MasterDir:    masterDir,
			GeneratedDir: generatedDir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Build error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Build successful in %v\n", time.Since(start))

	case "validate":
		start := time.Now()
		fmt.Println("Validating dataset integrity...")
		errs := mapping.Validate(mapping.ValidateOptions{
			MasterDir:    masterDir,
			GeneratedDir: generatedDir,
		})
		if len(errs) > 0 {
			fmt.Fprintf(os.Stderr, "Validation failed with %d error(s):\n", len(errs))
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %s\n", e)
			}
			os.Exit(1)
		}
		fmt.Printf("Validation passed with zero errors in %v\n", time.Since(start))

	case "clean":
		cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)
		apply := cleanCmd.Bool("apply", false, "Write cleaned entries back to master/ files")
		cleanCmd.Parse(os.Args[2:])

		fmt.Println("Cleaning dataset (normalizing titles and deduplicating synonyms)...")
		stats, err := mapping.Clean(mapping.CleanOptions{
			MasterDir: masterDir,
			Apply:     *apply,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Clean error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Clean finished: %d total entries scanned, %d titles trimmed, %d duplicate/empty synonyms removed.\n",
			stats.TotalEntries, stats.TitlesTrimmed, stats.SynonymsRemoved)
		if *apply {
			fmt.Printf("Changes saved across %d files. Rebuilding distributions...\n", stats.FilesModified)
			_ = mapping.Build(mapping.BuildOptions{
				MasterDir:    masterDir,
				GeneratedDir: generatedDir,
			})
		} else {
			fmt.Println("Dry-run complete. Run with --apply to write changes.")
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("aniSync Community Mapping CLI")
	fmt.Println()
	fmt.Println("Usage: go run ./cmd/mapping <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build            Compile master/*.yaml into generated/ distributions (JSON & SQLite)")
	fmt.Println("  validate         Validate dataset integrity, schema constraints, and ID uniqueness")
	fmt.Println("  clean [--apply]  Normalize titles and deduplicate synonyms in master/*.yaml")
}
