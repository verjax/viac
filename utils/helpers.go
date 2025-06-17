package utils

import (
	"regexp"
	"strings"
)

// InferIcon returns an emoji icon based on service image type
func InferIcon(image string) string {
	image = strings.ToLower(image)
	switch {
	case strings.Contains(image, "postgres"), strings.Contains(image, "postgresql"):
		return "🐘"
	case strings.Contains(image, "mysql"), strings.Contains(image, "mariadb"):
		return "🛢️"
	case strings.Contains(image, "mongodb"), strings.Contains(image, "mongo"):
		return "🍃"
	case strings.Contains(image, "redis"):
		return "🧠"
	case strings.Contains(image, "nginx"):
		return "🌐"
	case strings.Contains(image, "apache"), strings.Contains(image, "httpd"):
		return "🔥"
	case strings.Contains(image, "node"), strings.Contains(image, "nodejs"):
		return "💚"
	case strings.Contains(image, "python"), strings.Contains(image, "django"), strings.Contains(image, "flask"):
		return "🐍"
	case strings.Contains(image, "golang"), strings.Contains(image, "go"):
		return "🐹"
	case strings.Contains(image, "java"), strings.Contains(image, "openjdk"), strings.Contains(image, "spring"):
		return "☕"
	case strings.Contains(image, "php"):
		return "🐘"
	case strings.Contains(image, "ruby"), strings.Contains(image, "rails"):
		return "💎"
	case strings.Contains(image, "elasticsearch"), strings.Contains(image, "elastic"):
		return "🔍"
	case strings.Contains(image, "kibana"):
		return "📊"
	case strings.Contains(image, "logstash"):
		return "📋"
	case strings.Contains(image, "prometheus"):
		return "📈"
	case strings.Contains(image, "grafana"):
		return "📉"
	case strings.Contains(image, "ubuntu"), strings.Contains(image, "debian"), strings.Contains(image, "alpine"), strings.Contains(image, "centos"):
		return "🖥️"
	case strings.Contains(image, "traefik"):
		return "🚪"
	case strings.Contains(image, "consul"):
		return "🏛️"
	case strings.Contains(image, "vault"):
		return "🔐"
	case strings.Contains(image, "kafka"):
		return "📨"
	case strings.Contains(image, "rabbitmq"):
		return "🐰"
	case strings.Contains(image, "memcached"):
		return "💾"
	default:
		return "📦"
	}
}

// ExtractVersion parses version tag from image like "nginx:1.23"
func ExtractVersion(image string) string {
	if parts := strings.Split(image, ":"); len(parts) > 1 {
		version := parts[1]
		if version == "latest" || version == "stable" {
			return version
		}
		re := regexp.MustCompile(`(\d+\.?\d*\.?\d*)`)
		if matches := re.FindStringSubmatch(version); len(matches) > 1 {
			return "v" + matches[1]
		}
		return version
	}
	return "latest"
}
