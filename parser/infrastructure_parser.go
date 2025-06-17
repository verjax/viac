package parser

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
	"viac/models"
)

// DockerCompose represents the structure of docker-compose.yml
type DockerCompose struct {
	Version  string                   `yaml:"version"`
	Services map[string]DockerService `yaml:"services"`
	Networks map[string]interface{}   `yaml:"networks,omitempty"`
	Volumes  map[string]interface{}   `yaml:"volumes,omitempty"`
}

type DockerService struct {
	Image         string      `yaml:"image"`
	Build         interface{} `yaml:"build,omitempty"`
	Ports         []string    `yaml:"ports,omitempty"`
	Volumes       []string    `yaml:"volumes,omitempty"`
	Environment   interface{} `yaml:"environment,omitempty"` // Can be map[string]string or []string
	DependsOn     []string    `yaml:"depends_on,omitempty"`
	Networks      []string    `yaml:"networks,omitempty"`
	Restart       string      `yaml:"restart,omitempty"`
	Command       interface{} `yaml:"command,omitempty"`
	Links         []string    `yaml:"links,omitempty"`
	ExternalLinks []string    `yaml:"external_links,omitempty"`
}

// CloudInit represents the structure of cloud-init files
type CloudInit struct {
	CloudConfig bool                   `yaml:"#cloud-config,omitempty"`
	Packages    []string               `yaml:"packages,omitempty"`
	Runcmd      []string               `yaml:"runcmd,omitempty"`
	WriteFiles  []models.CloudInitFile `yaml:"write_files,omitempty"`
	Users       []interface{}          `yaml:"users,omitempty"`
	SSH         interface{}            `yaml:"ssh,omitempty"`
	Services    []string               `yaml:"services,omitempty"`
	Bootcmd     []string               `yaml:"bootcmd,omitempty"`
}

func ParseInfrastructureFile(filePath string) (*models.InfrastructureMap, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	infraMap := &models.InfrastructureMap{
		Components:    make(map[string]*models.InfrastructureComponent),
		Relationships: make([]models.Relationship, 0),
		Sources:       make(map[string]string),
	}

	filename := filepath.Base(filePath)

	// Try to determine file type based on name and content
	if isDockerCompose(filename, string(data)) {
		return parseDockerCompose(data, filePath, infraMap)
	} else if isCloudInit(filename, string(data)) {
		return parseCloudInit(data, filePath, infraMap)
	}

	return nil, fmt.Errorf("file %s is not a recognized Docker Compose or cloud-init file", filePath)
}

func isDockerCompose(filename, content string) bool {
	// Check filename patterns
	if strings.Contains(filename, "docker-compose") || strings.Contains(filename, "compose") {
		return true
	}

	// Check content for Docker Compose indicators
	return strings.Contains(content, "services:") &&
		(strings.Contains(content, "version:") || strings.Contains(content, "image:"))
}

func isCloudInit(filename, content string) bool {
	// Check filename patterns
	if strings.Contains(filename, "cloud-init") || strings.Contains(filename, "user-data") {
		return true
	}

	// Check content for cloud-init indicators
	return strings.Contains(content, "#cloud-config") ||
		(strings.Contains(content, "packages:") && strings.Contains(content, "runcmd:")) ||
		strings.Contains(content, "write_files:")
}

func parseDockerCompose(data []byte, filePath string, infraMap *models.InfrastructureMap) (*models.InfrastructureMap, error) {
	var compose DockerCompose
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse Docker Compose file: %v", err)
	}

	source := filePath
	infraMap.Sources[filePath] = "docker-compose"

	// Parse services
	for serviceName, service := range compose.Services {
		// Parse environment variables (handle both map and list formats)
		envMap := parseEnvironment(service.Environment)

		component := &models.InfrastructureComponent{
			Name:        serviceName,
			Type:        models.DockerService,
			Source:      source,
			Image:       service.Image,
			Ports:       service.Ports,
			Volumes:     service.Volumes,
			Environment: envMap,
			DependsOn:   service.DependsOn,
			Networks:    service.Networks,
		}

		infraMap.Components[serviceName] = component

		// Create explicit dependency relationships
		for _, dep := range service.DependsOn {
			infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
				From:        serviceName,
				To:          dep,
				Type:        models.DependsOn,
				Description: fmt.Sprintf("%s depends on %s", serviceName, dep),
			})
		}

		// Create relationships from links
		for _, link := range service.Links {
			// Handle "service:alias" format
			linkedService := strings.Split(link, ":")[0]
			infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
				From:        serviceName,
				To:          linkedService,
				Type:        models.ConnectsTo,
				Description: fmt.Sprintf("%s links to %s", serviceName, linkedService),
			})
		}
	}

	detectAdvancedRelationships(infraMap)

	return infraMap, nil
}

func parseCloudInit(data []byte, filePath string, infraMap *models.InfrastructureMap) (*models.InfrastructureMap, error) {
	var cloudInit CloudInit
	if err := yaml.Unmarshal(data, &cloudInit); err != nil {
		return nil, fmt.Errorf("failed to parse cloud-init file: %v", err)
	}

	source := filePath
	infraMap.Sources[filePath] = "cloud-init"

	// Create a component for the cloud-init configuration
	componentName := fmt.Sprintf("cloud-init-%s", strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)))

	component := &models.InfrastructureComponent{
		Name:        componentName,
		Type:        models.CloudInitConfig,
		Source:      source,
		Packages:    cloudInit.Packages,
		Services:    cloudInit.Services,
		WriteFiles:  cloudInit.WriteFiles,
		RunCommands: cloudInit.Runcmd,
	}

	infraMap.Components[componentName] = component

	// Detect relationships with Docker services if they exist
	detectCloudInitRelationships(infraMap, component)

	return infraMap, nil
}

func detectAdvancedRelationships(infraMap *models.InfrastructureMap) {
	detectExplicitPortConnections(infraMap)
	detectSharedVolumes(infraMap)
	detectSharedNetworks(infraMap)
	detectEnvironmentReferences(infraMap)
}

func detectExplicitPortConnections(infraMap *models.InfrastructureMap) {
	portToServices := make(map[string][]string)

	for name, comp := range infraMap.Components {
		if comp.Type != models.DockerService {
			continue
		}

		for _, portMapping := range comp.Ports {
			// Only consider explicit port mappings like "8080:80"
			if strings.Contains(portMapping, ":") {
				externalPort := strings.Split(portMapping, ":")[0]
				// Only track common web ports to avoid noise
				if isCommonWebPort(externalPort) {
					portToServices[externalPort] = append(portToServices[externalPort], name)
				}
			}
		}
	}

	// Only create connections for ports with exactly 2 services (likely load balancer -> service)
	for port, services := range portToServices {
		if len(services) == 2 {
			// Create a single directional connection from first to second
			infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
				From:        services[0],
				To:          services[1],
				Type:        models.ConnectsTo,
				Description: fmt.Sprintf("%s forwards port %s to %s", services[0], port, services[1]),
			})
		}
	}
}

func detectSharedVolumes(infraMap *models.InfrastructureMap) {
	volumeUsers := make(map[string][]string)

	for name, comp := range infraMap.Components {
		if comp.Type != models.DockerService {
			continue
		}

		for _, volume := range comp.Volumes {
			volumeName := strings.Split(volume, ":")[0]
			if !strings.HasPrefix(volumeName, "/") && !strings.HasPrefix(volumeName, ".") && volumeName != "" {
				volumeUsers[volumeName] = append(volumeUsers[volumeName], name)
			}
		}
	}

	for volumeName, users := range volumeUsers {
		if len(users) == 2 {
			infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
				From:        users[0],
				To:          users[1],
				Type:        models.VolumeMount,
				Description: fmt.Sprintf("%s shares volume %s with %s", users[0], volumeName, users[1]),
			})
		}
	}
}

func detectSharedNetworks(infraMap *models.InfrastructureMap) {
	networkUsers := make(map[string][]string)

	for name, comp := range infraMap.Components {
		if comp.Type != models.DockerService {
			continue
		}

		for _, network := range comp.Networks {
			// Skip default network
			if network != "default" && network != "" {
				networkUsers[network] = append(networkUsers[network], name)
			}
		}
	}

	for networkName, users := range networkUsers {
		if len(users) >= 2 && len(users) <= 4 {
			// Create a star topology: first service connects to others
			for i := 1; i < len(users); i++ {
				infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
					From:        users[0],
					To:          users[i],
					Type:        models.NetworkLink,
					Description: fmt.Sprintf("%s connects to %s via network %s", users[0], users[i], networkName),
				})
			}
		}
	}
}

func detectEnvironmentReferences(infraMap *models.InfrastructureMap) {
	for name1, comp1 := range infraMap.Components {
		if comp1.Type != models.DockerService {
			continue
		}

		for envKey, envValue := range comp1.Environment {
			// Only look for URL-style references
			if strings.Contains(envValue, "://") {
				parts := strings.Split(envValue, "://")
				if len(parts) > 1 {
					hostPart := strings.Split(parts[1], "/")[0]
					hostPart = strings.Split(hostPart, ":")[0]

					// Check if this hostname matches a service name exactly
					if comp2, exists := infraMap.Components[hostPart]; exists && comp2.Type == models.DockerService && name1 != hostPart {
						infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
							From:        name1,
							To:          hostPart,
							Type:        models.ConnectsTo,
							Description: fmt.Sprintf("%s connects to %s via %s", name1, hostPart, envKey),
						})
					}
				}
			}
		}
	}
}

func isCommonWebPort(port string) bool {
	commonPorts := map[string]bool{
		"80":   true,
		"443":  true,
		"8080": true,
		"3000": true,
		"8000": true,
		"5000": true,
	}
	return commonPorts[port]
}

func detectCloudInitRelationships(infraMap *models.InfrastructureMap, cloudInitComp *models.InfrastructureComponent) {
	// Check if cloud-init installs Docker or configures services that relate to Docker services
	for _, pkg := range cloudInitComp.Packages {
		if strings.Contains(pkg, "docker") {
			// This cloud-init provisions Docker, so it provisions all Docker services
			for name, comp := range infraMap.Components {
				if comp.Type == models.DockerService {
					infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
						From:        cloudInitComp.Name,
						To:          name,
						Type:        models.Provisions,
						Description: fmt.Sprintf("%s provisions Docker for %s", cloudInitComp.Name, name),
					})
				}
			}
		}
	}

	// Check if cloud-init configures services that match Docker service names
	for _, service := range cloudInitComp.Services {
		if comp, exists := infraMap.Components[service]; exists && comp.Type == models.DockerService {
			infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
				From:        cloudInitComp.Name,
				To:          service,
				Type:        models.Configures,
				Description: fmt.Sprintf("%s configures service %s", cloudInitComp.Name, service),
			})
		}
	}

	// Check run commands for Docker/service references
	for _, cmd := range cloudInitComp.RunCommands {
		cmdLower := strings.ToLower(cmd)

		if strings.Contains(cmdLower, "docker") {
			for name, comp := range infraMap.Components {
				if comp.Type == models.DockerService && strings.Contains(cmdLower, strings.ToLower(name)) {
					infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
						From:        cloudInitComp.Name,
						To:          name,
						Type:        models.Configures,
						Description: fmt.Sprintf("%s runs commands affecting %s", cloudInitComp.Name, name),
					})
				}
			}
		}
	}

	for _, file := range cloudInitComp.WriteFiles {
		if strings.Contains(file.Path, "docker") || strings.Contains(file.Content, "docker") {
			for name, comp := range infraMap.Components {
				if comp.Type == models.DockerService {
					infraMap.Relationships = append(infraMap.Relationships, models.Relationship{
						From:        cloudInitComp.Name,
						To:          name,
						Type:        models.Configures,
						Description: fmt.Sprintf("%s writes Docker configuration for %s", cloudInitComp.Name, name),
					})
				}
			}
		}
	}
}

// parseEnvironment handles both map and list formats for environment variables
func parseEnvironment(env interface{}) map[string]string {
	envMap := make(map[string]string)

	if env == nil {
		return envMap
	}

	switch v := env.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if strValue, ok := value.(string); ok {
				envMap[key] = strValue
			} else {
				envMap[key] = fmt.Sprintf("%v", value)
			}
		}
	case map[interface{}]interface{}:
		// Handle YAML map format
		for key, value := range v {
			if keyStr, ok := key.(string); ok {
				if strValue, ok := value.(string); ok {
					envMap[keyStr] = strValue
				} else {
					envMap[keyStr] = fmt.Sprintf("%v", value)
				}
			}
		}
	case []interface{}:
		// Handle list format: environment: [ "KEY=value", "KEY2=value2" ]
		for _, item := range v {
			if strItem, ok := item.(string); ok {
				if parts := strings.SplitN(strItem, "=", 2); len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				} else {
					envMap[strItem] = ""
				}
			}
		}
	case []string:
		for _, item := range v {
			if parts := strings.SplitN(item, "=", 2); len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			} else {
				envMap[item] = ""
			}
		}
	default:
	}

	return envMap
}
