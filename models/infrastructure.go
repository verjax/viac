package models

// InfrastructureComponent represents either a Docker Compose service or cloud-init configuration
type InfrastructureComponent struct {
	Name        string            `yaml:"name"`
	Type        ComponentType     `yaml:"type"`
	Source      string            `yaml:"source"`
	Image       string            `yaml:"image,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`

	// Cloud-init specific fields
	Packages    []string        `yaml:"packages,omitempty"`
	Services    []string        `yaml:"services,omitempty"`
	WriteFiles  []CloudInitFile `yaml:"write_files,omitempty"`
	RunCommands []string        `yaml:"runcmd,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

type ComponentType string

const (
	DockerService   ComponentType = "docker_service"
	CloudInitConfig ComponentType = "cloud_init"
)

type CloudInitFile struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Permissions string `yaml:"permissions,omitempty"`
}

// Relationship represents connections between infrastructure components
type Relationship struct {
	From        string
	To          string
	Type        RelationType
	Description string
}

type RelationType string

const (
	DependsOn   RelationType = "depends_on"
	ConnectsTo  RelationType = "connects_to"
	Configures  RelationType = "configures"
	Provisions  RelationType = "provisions"
	NetworkLink RelationType = "network_link"
	VolumeMount RelationType = "volume_mount"
)

// InfrastructureMap holds all discovered components and their relationships
type InfrastructureMap struct {
	Components    map[string]*InfrastructureComponent
	Relationships []Relationship
	Sources       map[string]string // filename -> source type
}
