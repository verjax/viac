package visualizer

import (
	"fmt"
	"os"
	"strings"
	"viac/models"
	"viac/utils"
)

// VisualizeInfrastructure prints ASCII visualization to terminal
func VisualizeInfrastructure(infraMap *models.InfrastructureMap) {
	fmt.Println("🏗️  Infrastructure Overview")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Group components by type
	dockerServices := make([]*models.InfrastructureComponent, 0)
	cloudInitConfigs := make([]*models.InfrastructureComponent, 0)

	for _, comp := range infraMap.Components {
		switch comp.Type {
		case models.DockerService:
			dockerServices = append(dockerServices, comp)
		case models.CloudInitConfig:
			cloudInitConfigs = append(cloudInitConfigs, comp)
		}
	}

	// Display Docker Services
	if len(dockerServices) > 0 {
		fmt.Println("\n📦 Docker Services:")
		for _, comp := range dockerServices {
			icon := utils.InferIcon(comp.Image)
			version := utils.ExtractVersion(comp.Image)

			fmt.Printf("  %s %s (%s)\n", icon, comp.Name, version)
			fmt.Printf("    └── Image: %s\n", comp.Image)

			if len(comp.Ports) > 0 {
				fmt.Printf("    └── Ports: %s\n", strings.Join(comp.Ports, ", "))
			}

			if len(comp.Environment) > 0 {
				fmt.Printf("    └── Environment: %d variables\n", len(comp.Environment))
			}

			if len(comp.Networks) > 0 {
				fmt.Printf("    └── Networks: %s\n", strings.Join(comp.Networks, ", "))
			}

			fmt.Printf("    └── Source: %s\n", comp.Source)
			fmt.Println()
		}
	}

	// Display Cloud-Init Configurations
	if len(cloudInitConfigs) > 0 {
		fmt.Println("\n☁️  Cloud-Init Configurations:")
		for _, comp := range cloudInitConfigs {
			fmt.Printf("  🖥️  %s\n", comp.Name)

			if len(comp.Packages) > 0 {
				fmt.Printf("    └── Packages: %s\n", strings.Join(comp.Packages, ", "))
			}

			if len(comp.Services) > 0 {
				fmt.Printf("    └── Services: %s\n", strings.Join(comp.Services, ", "))
			}

			if len(comp.WriteFiles) > 0 {
				fmt.Printf("    └── Files: %d configuration files\n", len(comp.WriteFiles))
			}

			if len(comp.RunCommands) > 0 {
				fmt.Printf("    └── Commands: %d run commands\n", len(comp.RunCommands))
			}

			fmt.Printf("    └── Source: %s\n", comp.Source)
			fmt.Println()
		}
	}

	// Display Relationships
	if len(infraMap.Relationships) > 0 {
		fmt.Println("\n🔗 Infrastructure Relationships:")

		// Group relationships by type
		relsByType := make(map[models.RelationType][]models.Relationship)
		for _, rel := range infraMap.Relationships {
			relsByType[rel.Type] = append(relsByType[rel.Type], rel)
		}

		for relType, rels := range relsByType {
			fmt.Printf("  %s %s:\n", getRelationshipIcon(relType), relType)
			for _, rel := range rels {
				fmt.Printf("    └── %s → %s\n", rel.From, rel.To)
			}
			fmt.Println()
		}
	}

	// Summary
	fmt.Println("📊 Summary:")
	fmt.Printf("  └── Docker Services: %d\n", len(dockerServices))
	fmt.Printf("  └── Cloud-Init Configs: %d\n", len(cloudInitConfigs))
	fmt.Printf("  └── Relationships: %d\n", len(infraMap.Relationships))
	fmt.Printf("  └── Source Files: %d\n", len(infraMap.Sources))
}

func getRelationshipIcon(relType models.RelationType) string {
	switch relType {
	case models.DependsOn:
		return "🔗"
	case models.ConnectsTo:
		return "🌐"
	case models.Configures:
		return "⚙️"
	case models.Provisions:
		return "🚀"
	case models.NetworkLink:
		return "🔌"
	case models.VolumeMount:
		return "💾"
	default:
		return "📎"
	}
}

// ExportInfrastructureGraphviz creates an interconnected network visualization like Go dependency graphs
func ExportInfrastructureGraphviz(infraMap *models.InfrastructureMap, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer f.Close()

	// Write Graphviz header in Go dependency style
	fmt.Fprintf(f, "digraph infrastructure {\n")
	fmt.Fprintf(f, " fontname=\"Helvetica,Arial,sans-serif\"\n")
	fmt.Fprintf(f, " node [fontname=\"Helvetica,Arial,sans-serif\"]\n")
	fmt.Fprintf(f, " edge [fontname=\"Helvetica,Arial,sans-serif\"]\n")

	// Create node mapping (component name -> node ID)
	nodeMap := make(map[string]string)
	nodeCounter := 0

	// First pass: create all nodes with numbered IDs
	for _, comp := range infraMap.Components {
		nodeId := fmt.Sprintf("n%d", nodeCounter)
		nodeMap[comp.Name] = nodeId

		var label string
		var tooltip string

		if comp.Type == models.DockerService {
			icon := utils.InferIcon(comp.Image)
			label = fmt.Sprintf("%s %s", icon, comp.Name)

			tooltip = fmt.Sprintf("Docker Service: %s", comp.Image)
			if len(comp.Ports) > 0 {
				tooltip += fmt.Sprintf(" | Ports: %s", strings.Join(comp.Ports, ", "))
			}

		} else if comp.Type == models.CloudInitConfig {
			label = fmt.Sprintf("🖥️ %s", comp.Name)
			tooltip = "Cloud-Init Configuration"
			if len(comp.Packages) > 0 {
				tooltip += fmt.Sprintf(" | Packages: %d", len(comp.Packages))
			}
		}

		// Write node definition
		_, err := fmt.Fprintf(f, " %s [label=\"%s\", tooltip=\"%s\"];\n", nodeId, label, tooltip)
		if err != nil {
			return err
		}
		nodeCounter++
	}

	processedEdges := make(map[string]bool) // Track to avoid duplicate edges

	for _, rel := range infraMap.Relationships {
		fromNodeId, fromExists := nodeMap[rel.From]
		toNodeId, toExists := nodeMap[rel.To]

		if !fromExists || !toExists {
			continue // Skip if nodes don't exist
		}

		edgeKey := fmt.Sprintf("%s->%s", fromNodeId, toNodeId)
		if processedEdges[edgeKey] {
			continue
		}
		processedEdges[edgeKey] = true

		_, err2 := fmt.Fprintf(f, " %s -> %s;\n", fromNodeId, toNodeId)
		if err2 != nil {
			return err2
		}
	}

	fmt.Fprintf(f, "}\n")
	return nil
}
