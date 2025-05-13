package mcp

type ServerConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Version     string            `json:"version"`
	Options     map[string]string `json:"options"`
}

type FileSystemOptions struct {
	AllowedDirs []string `json:"allowed_dirs"`
	ReadOnly    bool     `json:"read_only"`
	Transport   string   `json:"transport,omitempty"`
	Port        int      `json:"port,omitempty"`
}
