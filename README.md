# VIAC - Visual Infrastructure as Code

### Author:
**Nyandoro Christopher**

**VIAC** is a specialized CLI tool for visualizing Infrastructure-as-Code configurations, focusing specifically on **Docker Compose** and **cloud-init** files found in Git repositories. It discovers, parses, and visualizes the relationships between infrastructure components to provide a comprehensive view of your infrastructure setup.

## 🎯 Purpose

VIAC helps you understand your infrastructure by:
- 🔍 **Discovering** Docker Compose and cloud-init files from Git repositories
- 🧩 **Parsing** infrastructure components and their relationships
- 📊 **Visualizing** dependencies, connections, and provisioning relationships
- 🖼️ **Exporting** professional Graphviz diagrams

## 🔧 Installation

### Prerequisites
- Go 1.23+ installed
- Git installed (for repository cloning)
- Graphviz installed (optional, for rendering .dot files to images)

### Build from Source

```bash
git clone https://git.mif.vu.lt/micac/2025/viac
cd viac
go mod tidy
go build -o viac
```

## 📝 Configuration

Create a `config.yaml` file to specify the Git or File path to analyze.:

```yaml
# config.yaml

# Git repository to clone and analyze for Docker Compose and cloud-init files
git_repo: https://github.com/verjax/viac.git

# Enable verbose output for detailed logging
file_paths:
  - docker-compose.yaml #AI generated example file for testing.

# Enable verbose output for detailed logging
verbose: true
```

## 🖥️ Usage

### Terminal Visualization
```bash
# Display infrastructure in terminal
./viac config.yaml

# With verbose output
./viac config.yaml --verbose

# Or in development terminal
go run main.go config.yaml --verbose
```

### Export Graphviz Diagram
```bash
# Export to .dot file
./viac config.yaml --export infrastructure.dot

# Convert to PNG (requires Graphviz)
`dot -Tpng infrastructure.dot -o output.png`

# Convert to PDF
dot -Tpdf infrastructure.dot -o output.pdf

# Convert to SVG
dot -Tsvg infrastructure.dot -o output.svg
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with Go and the excellent Cobra CLI framework
- Visualization powered by Graphviz
- Git operations using go-git library
