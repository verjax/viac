# VIAC
### Author: 
**Nyandoro Christopher**


**Viac** is a lightweight and powerful CLI tool to visualize infrastructure described using Infrastructure-as-Code (IaC) YAML configurations such as Docker Compose and cloud-init files. It supports input from files, Git repositories, local directories, or automatic detection from common server paths.


Viac can output:
- 📦 **Terminal ASCII diagrams** of services and dependencies
- 📈 **Graphviz `.dot` files** for graphical visualization (with optional PNG/PDF rendering)
---

## 🚀 Features

- ✅ **Multi-source support** via a single config file:
    - Parse local `file_paths`
    - Clone and process `git_repos`
    - Scan `directories` recursively
    - Auto-detect YAMLs from common paths
- ✅ **YAML and JSON config formats**
- ✅ **Graphviz export** with `--export`
- ✅ **Verbose logging** with `verbose: true`
- ✅ **Clean CLI** powered by Cobra
- ✅ **Pure Go** implementation, no external binaries needed

---

## 🛠 Installation

### Prerequisites
- Go 1.18 or higher installed
- Git installed (for `go-git` cloning)

### Clone & Build

```bash
git clone https://github.com/yourusername/viac.git
cd viac
go build -o Viac
```

```yaml
# config.yaml

file_paths:
  - docker-compose.yml
  - cloud-init.yml

git_repos:
  - https://github.com/example/infrastructure.git
  - https://github.com/example/services-config.git

directories:
  - ./viac-configs
  - /srv/app-configs

auto_detect: true
verbose: true
```

### Run normally (ASCII output in terminal)
```bash

./Viac config.yaml

```
### Run and export a Graphviz .dot file
```bash

./Viac config.yaml --e output.dot

```
### Convert .dot to PNG using Graphviz (optional)
```bash

dot -Tpng output.dot -o output.png

```
      
