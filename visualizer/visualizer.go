package visualizer

import (
	"fmt"
	"os"
	"strings"

	"viac/models"
	"viac/utils"
)

type Infrastructure struct {
	Name    string
	Compose models.Node
}

// Visualize prints simple ASCII output to terminal
func Visualize(compose models.Node) {
	for serviceName, service := range compose.Services {
		fmt.Printf("📦 %s\n", serviceName)
		fmt.Printf("  └── image: %s\n", service.Image)
		if len(service.DependsOn) > 0 {
			fmt.Println("  └── depends_on:")
			for _, dep := range service.DependsOn {
				fmt.Printf("      └── %s\n", dep)
			}
		}
		fmt.Println()
	}
}

// ExportGraphviz outputs a rich .dot file with icons, versions, groupings, and relationship types
func ExportGraphviz(composes []Infrastructure, path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating export file:", err)
		return
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("Error closing export file:", err)
		}
	}(f)

	write := func(data string) bool {
		_, err := f.WriteString(data)
		return err == nil
	}

	if !write("digraph viac {\n") ||
		!write("  graph [fontsize=10];\n") ||
		!write("  node [shape=box, fontname=\"Arial\", style=\"filled\", fillcolor=\"white\"];\n") ||
		!write("  edge [fontname=\"Arial\", fontsize=9];\n\n") {
		return
	}

	allNodes := map[string]models.NodeMeta{}

	for _, nc := range composes {
		clusterID := utils.SanitizeID(nc.Name)
		if !write(fmt.Sprintf("  subgraph cluster_%s {\n", clusterID)) ||
			!write(fmt.Sprintf("    label = \"%s\";\n", nc.Name)) ||
			!write("    style = \"dashed\";\n\n") {
			return
		}

		for name, svc := range nc.Compose.Services {
			icon := utils.InferIcon(svc.Image)
			version := utils.ExtractVersion(svc.Image)

			portLine := ""
			if len(svc.Ports) > 0 {
				portLine = fmt.Sprintf("<TR><TD><FONT POINT-SIZE=\"9\">Ports: %s</FONT></TD></TR>", strings.Join(svc.Ports, ", "))
			}

			label := fmt.Sprintf(
				"<<TABLE BORDER=\"0\" CELLBORDER=\"0\">"+
					"<TR><TD>%s %s</TD></TR>"+
					"<TR><TD><FONT POINT-SIZE=\"10\">%s</FONT></TD></TR>"+
					"%s</TABLE>>",
				icon, name, version, portLine,
			)

			if !write(fmt.Sprintf("    \"%s\" [label=%s tooltip=\"From: %s\"];\n", name, label, nc.Name)) {
				return
			}

			for _, dep := range svc.DependsOn {
				if !write(fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"depends_on\", color=\"blue\"];\n", name, dep)) {
					return
				}
			}

			allNodes[name] = models.NodeMeta{
				Name:    name,
				Ports:   svc.Ports,
				Depends: svc.DependsOn,
				File:    nc.Name,
				Image:   svc.Image,
			}
		}

		if !write("  }\n\n") {
			return
		}
	}

	// Detect connects_to relationships based on matching ports
	for fromName, from := range allNodes {
		for toName, to := range allNodes {
			if fromName == toName || utils.StringInSlice(toName, from.Depends) {
				continue
			}
			if utils.MatchesPortInterest(from, to) {
				if !write(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"connects_to\", style=dashed, color=gray];\n", fromName, toName)) {
					return
				}
			}
		}
	}

	_ = write("}\n")
}
