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
	"viac/models"
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
		_, err := fmt.Fprintf(os.Stderr, "Error executing VIAC: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "viac [config file]",
	Short: "VIAC visualizes Docker Compose and cloud-init infrastructure configurations",
	Long:  `VIAC (Visual Infrastructure as Code) parses Docker Compose and cloud-init files from Git repositories and local files to visualize infrastructure relationships and dependencies.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath = args[0]

		cfg, err := setup.LoadConfig(configPath)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}

		if verbose {
			cfg.Verbose = true
		}

		if cfg.Verbose {
			fmt.Printf("Loaded config: %+v\n", cfg)
		}

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Scanning infrastructure files..."
		s.Start()

		// Initialize infrastructure map
		combinedInfraMap := &models.InfrastructureMap{
			Components:    make(map[string]*models.InfrastructureComponent),
			Relationships: make([]models.Relationship, 0),
			Sources:       make(map[string]string),
		}

		if len(cfg.FilePaths) > 0 {
			if cfg.Verbose {
				fmt.Printf("Processing %d local files...\n", len(cfg.FilePaths))
			}

			for _, filePath := range cfg.FilePaths {
				if cfg.Verbose {
					fmt.Printf("Processing file: %s\n", filePath)
				}
				processInfrastructureFile(filePath, combinedInfraMap, cfg.Verbose)
			}
		}

		// Process Git repository if specified
		if cfg.GitRepo != "" {
			if cfg.Verbose {
				fmt.Printf("Processing Git repository: %s\n", cfg.GitRepo)
			}

			err = processGitRepo(cfg.GitRepo, combinedInfraMap, cfg.Verbose)
			if err != nil {
				s.Stop()
				log.Fatalf("Error processing Git repository: %v", err)
			}
		}

		s.Stop()

		if len(combinedInfraMap.Components) == 0 {
			fmt.Println("No Docker Compose or cloud-init files found.")
			return
		}

		detectCrossFileRelationships(combinedInfraMap)

		if exportPath != "" {
			if err := visualizer.ExportInfrastructureGraphviz(combinedInfraMap, exportPath); err != nil {
				log.Fatalf("Error exporting Graphviz: %v", err)
			}
			fmt.Printf("✅ Exported infrastructure diagram to: %s\n", exportPath)
			fmt.Println("   Convert to PNG with: dot -Tpng", exportPath, "-o output.png")
		} else {
			visualizer.VisualizeInfrastructure(combinedInfraMap)
		}
	},
}

func processInfrastructureFile(filePath string, combinedMap *models.InfrastructureMap, verbose bool) {
	if verbose {
		fmt.Printf("Checking file: %s\n", filePath)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("File not found: %s\n", filePath)
		}
		return
	}

	infraMap, err := parser.ParseInfrastructureFile(filePath)
	if err != nil {
		if verbose {
			fmt.Printf("Skipping %s: %v\n", filePath, err)
		}
		return
	}

	if verbose {
		fmt.Printf("✅ Found infrastructure file: %s\n", filePath)
	}

	mergeInfrastructureMaps(combinedMap, infraMap)
}

func processGitRepo(url string, combinedMap *models.InfrastructureMap, verbose bool) error {
	repoName := strings.TrimSuffix(filepath.Base(url), ".git")
	repoPath := filepath.Join(".tmp", repoName)

	if err := os.MkdirAll(".tmp", 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	if _, err := os.Stat(repoPath); err == nil {
		if verbose {
			fmt.Printf("Repository already exists at %s, using existing clone\n", repoPath)
		}
	} else {
		if verbose {
			fmt.Printf("Cloning repository %s to %s\n", url, repoPath)
		}

		_, err := git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:      url,
			Progress: nil,
			Depth:    1,
		})

		if err != nil {
			return fmt.Errorf("failed to clone repository %s: %v", url, err)
		}
	}

	err := scanRepositoryForInfrastructure(repoPath, combinedMap, verbose)
	if err != nil {
		return fmt.Errorf("failed to scan repository: %v", err)
	}

	if verbose {
		fmt.Printf("Keeping repository at %s for future use\n", repoPath)
	}

	return nil
}

func scanRepositoryForInfrastructure(dir string, combinedMap *models.InfrastructureMap, verbose bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Error accessing %s: %v\n", path, err)
			}
			return nil // Continue walking despite errors
		}

		if info.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		if strings.Contains(path, "/.") {
			return nil
		}

		if verbose {
			fmt.Printf("Checking file: %s\n", path)
		}

		infraMap, err := parser.ParseInfrastructureFile(path)
		if err != nil {
			if verbose {
				fmt.Printf("Skipping %s: %v\n", path, err)
			}
			return nil
		}

		if verbose {
			fmt.Printf("✅ Found infrastructure file: %s\n", path)
		}

		mergeInfrastructureMaps(combinedMap, infraMap)
		return nil
	})
}

func mergeInfrastructureMaps(target *models.InfrastructureMap, source *models.InfrastructureMap) {
	// Merge components
	for name, comp := range source.Components {
		if _, exists := target.Components[name]; !exists {
			target.Components[name] = comp
		} else {
			// Handle naming conflicts by adding source suffix
			newName := fmt.Sprintf("%s_%s", name, filepath.Base(comp.Source))
			target.Components[newName] = comp
			comp.Name = newName
		}
	}

	target.Relationships = append(target.Relationships, source.Relationships...)

	// Merge sources
	for source, sourceType := range source.Sources {
		target.Sources[source] = sourceType
	}
}

func detectCrossFileRelationships(infraMap *models.InfrastructureMap) {
	// Detect relationships between cloud-init and docker services across files
	cloudInitComponents := make([]*models.InfrastructureComponent, 0)
	dockerComponents := make([]*models.InfrastructureComponent, 0)

	for _, comp := range infraMap.Components {
		switch comp.Type {
		case models.CloudInitConfig:
			cloudInitComponents = append(cloudInitComponents, comp)
		case models.DockerService:
			dockerComponents = append(dockerComponents, comp)
		}
	}

	// Find cloud-init configs that provision Docker for services in other files
	for _, cloudInit := range cloudInitComponents {
		for _, pkg := range cloudInit.Packages {
			if strings.Contains(strings.ToLower(pkg), "docker") {
				for _, dockerComp := range dockerComponents {
					if dockerComp.Source != cloudInit.Source {
						// Cross-file provisioning relationship
						infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
							From:        cloudInit.Name,
							To:          dockerComp.Name,
							Type:        models.Provisions,
							Description: fmt.Sprintf("%s provisions Docker environment for %s", cloudInit.Name, dockerComp.Name),
						})
					}
				}
			}
		}

		for _, file := range cloudInit.WriteFiles {
			if strings.Contains(file.Path, "docker-compose") || strings.Contains(file.Content, "docker") {
				for _, dockerComp := range dockerComponents {
					infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
						From:        cloudInit.Name,
						To:          dockerComp.Name,
						Type:        models.Configures,
						Description: fmt.Sprintf("%s writes Docker configuration for %s", cloudInit.Name, dockerComp.Name),
					})
				}
			}
		}
	}
}
