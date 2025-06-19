package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// BuildInfo contains all build information
type BuildInfo struct {
	Version    string `json:"version"`
	GitCommit  string `json:"git_commit"`
	GitBranch  string `json:"git_branch"`
	GitTag     string `json:"git_tag"`
	BuildDate  string `json:"build_date"`
	GoVersion  string `json:"go_version"`
	Platform   string `json:"platform"`
	Compiler   string `json:"compiler"`
	IsModified bool   `json:"is_modified"`
	ModulePath string `json:"module_path"`
	ModuleSum  string `json:"module_sum"`
}

// GetBuildInfo extracts comprehensive build information from runtime debug info
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version:   "unknown",
		GitCommit: "unknown",
		GitBranch: "unknown",
		GitTag:    "unknown",
		BuildDate: "unknown",
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Compiler:  runtime.Compiler,
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	// Extract module information
	info.ModulePath = buildInfo.Main.Path
	info.ModuleSum = buildInfo.Main.Sum

	// If we have a version from go.mod/module system, use it
	if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		info.Version = buildInfo.Main.Version
	}

	// Extract VCS (Version Control System) information
	var hasVCSInfo bool
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs":
			// Usually "git"
			continue
		case "vcs.revision":
			info.GitCommit = setting.Value
			hasVCSInfo = true
			// Create short commit hash
			if len(setting.Value) > 7 {
				info.GitCommit = setting.Value[:7]
			}
		case "vcs.time":
			info.BuildDate = setting.Value
			hasVCSInfo = true
		case "vcs.modified":
			info.IsModified = setting.Value == "true"
			hasVCSInfo = true
		case "-tags":
			// Build tags could be useful
			continue
		case "CGO_ENABLED":
			// Could be useful for debugging
			continue
		}
	}

	// If we have VCS info but no version, create a development version
	if hasVCSInfo && (info.Version == "unknown" || info.Version == "(devel)") {
		if info.GitCommit != "unknown" {
			info.Version = fmt.Sprintf("dev-%s", info.GitCommit)
			if info.IsModified {
				info.Version += "-dirty"
			}
		}
	}

	// Try to parse git tag information if we have a proper version
	if strings.HasPrefix(info.Version, "v") && len(strings.Split(info.Version, ".")) >= 2 {
		info.GitTag = info.Version
	}

	// Format build date if it's in Go's time format
	if info.BuildDate != "unknown" {
		if parsedTime, err := time.Parse(time.RFC3339, info.BuildDate); err == nil {
			info.BuildDate = parsedTime.Format("2006-01-02T15:04:05Z")
		}
	}

	return info
}

// GetVersion returns a simple version string
func GetVersion() string {
	buildInfo := GetBuildInfo()
	return buildInfo.Version
}

// GetVersionWithCommit returns version with commit info
func GetVersionWithCommit() string {
	buildInfo := GetBuildInfo()

	if buildInfo.GitCommit != "unknown" {
		version := buildInfo.Version
		if buildInfo.IsModified {
			return fmt.Sprintf("%s (%s-dirty)", version, buildInfo.GitCommit)
		}
		return fmt.Sprintf("%s (%s)", version, buildInfo.GitCommit)
	}

	return buildInfo.Version
}

// GetShortVersion returns a concise version string
func GetShortVersion() string {
	buildInfo := GetBuildInfo()

	// For released versions, just return the version
	if strings.HasPrefix(buildInfo.Version, "v") && !strings.Contains(buildInfo.Version, "dev") {
		return buildInfo.Version
	}

	// For development versions, include commit
	if buildInfo.GitCommit != "unknown" {
		if buildInfo.IsModified {
			return fmt.Sprintf("%s-dirty", buildInfo.GitCommit)
		}
		return buildInfo.GitCommit
	}

	return buildInfo.Version
}

// GetDetailedVersion returns comprehensive version information
func GetDetailedVersion() string {
	info := GetBuildInfo()

	var parts []string

	parts = append(parts, fmt.Sprintf("Version: %s", info.Version))

	if info.GitCommit != "unknown" {
		commit := info.GitCommit
		if info.IsModified {
			commit += " (modified)"
		}
		parts = append(parts, fmt.Sprintf("Commit: %s", commit))
	}

	if info.BuildDate != "unknown" {
		parts = append(parts, fmt.Sprintf("Built: %s", info.BuildDate))
	}

	parts = append(parts, fmt.Sprintf("Go: %s", info.GoVersion))
	parts = append(parts, fmt.Sprintf("Platform: %s", info.Platform))

	if info.ModulePath != "" {
		parts = append(parts, fmt.Sprintf("Module: %s", info.ModulePath))
	}

	return strings.Join(parts, "\n")
}

// GetJSONVersion returns version info in JSON format
func GetJSONVersion() string {
	info := GetBuildInfo()

	return fmt.Sprintf(`{
  "version": "%s",
  "git_commit": "%s",
  "git_branch": "%s", 
  "git_tag": "%s",
  "build_date": "%s",
  "go_version": "%s",
  "platform": "%s",
  "compiler": "%s",
  "is_modified": %t,
  "module_path": "%s"
}`,
		info.Version,
		info.GitCommit,
		info.GitBranch,
		info.GitTag,
		info.BuildDate,
		info.GoVersion,
		info.Platform,
		info.Compiler,
		info.IsModified,
		info.ModulePath)
}

// IsRelease checks if this is a release build
func IsRelease() bool {
	info := GetBuildInfo()
	return strings.HasPrefix(info.Version, "v") &&
		!strings.Contains(info.Version, "dev") &&
		!info.IsModified
}

// IsDevelopment checks if this is a development build
func IsDevelopment() bool {
	return !IsRelease()
}

// GetBuildSummary returns a one-line build summary
func GetBuildSummary() string {
	info := GetBuildInfo()

	summary := info.Version
	if info.GitCommit != "unknown" && len(info.GitCommit) > 0 {
		summary += fmt.Sprintf(" (%s)", info.GitCommit)
	}
	if info.IsModified {
		summary += " [modified]"
	}
	if info.BuildDate != "unknown" {
		if parsedTime, err := time.Parse(time.RFC3339, info.BuildDate); err == nil {
			summary += fmt.Sprintf(" built %s", parsedTime.Format("2006-01-02"))
		}
	}

	return summary
}
