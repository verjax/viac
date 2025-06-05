package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"viac/parser"
	"viac/setup"
	"viac/visualizer"
)

var configPath string
var exportPath string
var verbose bool

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&exportPath, "export", "e", "", "Export Graphviz diagram to a .dot file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Oops. An error while executing InfraViz '%s'\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "Viac [config file]",
	Short: "Viac visualizes your infrastructure YAML files, Git repos, or directories",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath = args[0]

		cfg, err := setup.LoadConfig(configPath)
		if err != nil {
			log.Fatalf("Error loading config file: %v\n", err)
		}

		// If --verbose flag is used, override config's verbose field
		if verbose {
			cfg.Verbose = true
		}

		if cfg.Verbose {
			fmt.Printf("Loaded config: %+v\n", cfg)
		}

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()

		var allComposes []visualizer.Infrastructure

		// Process file_paths
		for _, path := range cfg.FilePaths {
			if cfg.Verbose {
				fmt.Println("Parsing file:", path)
			}
			compose, err := parser.ParseYAML(path)
			if err == nil {
				allComposes = append(allComposes, visualizer.Infrastructure{Name: path, Compose: compose})
			}
		}

		// Process git_repos
		for _, url := range cfg.GitRepos {
			repoName := strings.TrimSuffix(filepath.Base(url), ".git")
			repoPath := filepath.Join(".tmp", repoName)
			_ = os.MkdirAll(".tmp", 0755)
			_ = os.RemoveAll(repoPath)

			if cfg.Verbose {
				fmt.Println("Cloning repo:", url)
			}

			_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
				URL:      url,
				Progress: nil,
				Depth:    1,
			})
			if err == nil {
				_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
						compose, err := parser.ParseYAML(path)
						if err == nil {
							allComposes = append(allComposes, visualizer.Infrastructure{Name: path, Compose: compose})
						}
					}
					return nil
				})
			}

			_ = os.RemoveAll(repoPath)
		}

		// Process directories
		for _, dir := range cfg.Directories {
			_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
					compose, err := parser.ParseYAML(path)
					if err == nil {
						allComposes = append(allComposes, visualizer.Infrastructure{Name: path, Compose: compose})
					}
				}
				return nil
			})
		}

		// Auto-detect
		if cfg.AutoDetect {
			paths := []string{"/etc", "/srv", "/opt", "./"}
			for _, p := range paths {
				_ = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
						compose, err := parser.ParseYAML(path)
						if err == nil {
							allComposes = append(allComposes, visualizer.Infrastructure{Name: path, Compose: compose})
						}
					}
					return nil
				})
			}
		}

		s.Stop()

		// Output
		if exportPath != "" {
			visualizer.ExportGraphviz(allComposes, exportPath)
			fmt.Println("Exported Graphviz diagram to:", exportPath)
		} else {
			for _, nc := range allComposes {
				fmt.Printf("From: %s\n", nc.Name)
				visualizer.Visualize(nc.Compose)
			}
		}
	},
}
