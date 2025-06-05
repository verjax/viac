package utils

import (
	"strings"
	"viac/models"
)

// inferIcon returns an emoji icon based on service image type
func InferIcon(image string) string {
	image = strings.ToLower(image)
	switch {
	case strings.Contains(image, "postgres"), strings.Contains(image, "mysql"), strings.Contains(image, "mariadb"):
		return "🛢️"
	case strings.Contains(image, "nginx"), strings.Contains(image, "apache"):
		return "🌐"
	case strings.Contains(image, "redis"), strings.Contains(image, "memcached"):
		return "🧠"
	case strings.Contains(image, "ubuntu"), strings.Contains(image, "debian"), strings.Contains(image, "vm"):
		return "🖥️"
	case strings.Contains(image, "node"), strings.Contains(image, "golang"), strings.Contains(image, "java"):
		return "📦"
	default:
		return "📦"
	}
}

// extractVersion parses version tag from image like "nginx:1.23"
func ExtractVersion(image string) string {
	if parts := strings.Split(image, ":"); len(parts) > 1 {
		return parts[1]
	}
	return "latest"
}

// sanitizeID makes names safe for use in Graphviz IDs
func SanitizeID(name string) string {
	return strings.NewReplacer(".", "_", "-", "_", "/", "_").Replace(name)
}

// StringInSlice checks if string s exists in list
func StringInSlice(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

// MatchesPortInterest checks for overlapping exposed ports
func MatchesPortInterest(from models.NodeMeta, to models.NodeMeta) bool {
	for _, p1 := range from.Ports {
		for _, p2 := range to.Ports {
			if p1 == p2 {
				return true
			}
		}
	}
	return false
}
