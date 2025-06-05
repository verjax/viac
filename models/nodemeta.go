package models

type NodeMeta struct {
	Name    string
	Ports   []string
	Depends []string
	File    string
	Image   string
}
