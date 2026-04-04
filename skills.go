package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a discovered Agent Skill.
type Skill struct {
	Name        string
	Description string
	Location    string // absolute path to SKILL.md
}

// skillScanDirs returns directories to scan in precedence order (project first).
func skillScanDirs() []string {
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	var dirs []string
	if cwd != "" {
		dirs = append(dirs,
			filepath.Join(cwd, ".au", "skills"),
			filepath.Join(cwd, ".agents", "skills"),
		)
	}
	if home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".au", "skills"),
			filepath.Join(home, ".agents", "skills"),
		)
	}
	return dirs
}

// discoverSkills scans standard directories for Agent Skills.
// Project-level skills take precedence over user-level skills with the same name.
func discoverSkills() []Skill {
	seen := make(map[string]bool)
	var skills []Skill
	for _, dir := range skillScanDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			skillFile := filepath.Join(dir, e.Name(), "SKILL.md")
			if _, err := os.Stat(skillFile); err != nil {
				continue
			}
			name, desc, err := parseSkillFrontmatter(skillFile)
			if err != nil || name == "" || desc == "" {
				continue
			}
			if seen[name] {
				continue
			}
			seen[name] = true
			abs, err := filepath.Abs(skillFile)
			if err != nil {
				abs = skillFile
			}
			skills = append(skills, Skill{
				Name:        name,
				Description: desc,
				Location:    abs,
			})
		}
	}
	return skills
}

// parseSkillFrontmatter extracts name and description from a SKILL.md YAML frontmatter.
func parseSkillFrontmatter(path string) (name, description string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return "", "", fmt.Errorf("no frontmatter")
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// Strip surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		switch key {
		case "name":
			name = val
		case "description":
			description = val
		}
	}
	return name, description, scanner.Err()
}

// loadSkillBody reads the markdown body of a SKILL.md (everything after the frontmatter).
func loadSkillBody(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Skip frontmatter
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		// No frontmatter — return whole file
		var sb strings.Builder
		sb.WriteString(scanner.Text() + "\n")
		for scanner.Scan() {
			sb.WriteString(scanner.Text() + "\n")
		}
		return strings.TrimSpace(sb.String()), scanner.Err()
	}
	inFront := true
	var sb strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if inFront {
			if strings.TrimSpace(line) == "---" {
				inFront = false
			}
			continue
		}
		sb.WriteString(line + "\n")
	}
	return strings.TrimSpace(sb.String()), scanner.Err()
}

// buildSkillCatalog returns the system prompt section for skill disclosure.
// Returns empty string when no skills are available.
func buildSkillCatalog(skills []Skill) string {
	if len(skills) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\nThe following skills provide specialized instructions for specific tasks. ")
	sb.WriteString("When a task matches a skill's description, use your read_file tool to load ")
	sb.WriteString("the SKILL.md at the listed location before proceeding. ")
	sb.WriteString("When a skill references relative paths, resolve them against the skill's ")
	sb.WriteString("directory (the parent of SKILL.md) and use absolute paths in tool calls.\n\n")
	sb.WriteString("<available_skills>\n")
	for _, s := range skills {
		sb.WriteString("  <skill>\n")
		fmt.Fprintf(&sb, "    <name>%s</name>\n", s.Name)
		fmt.Fprintf(&sb, "    <description>%s</description>\n", s.Description)
		fmt.Fprintf(&sb, "    <location>%s</location>\n", s.Location)
		sb.WriteString("  </skill>\n")
	}
	sb.WriteString("</available_skills>")
	return sb.String()
}

// findSkill returns a skill by name (case-insensitive).
func findSkill(skills []Skill, name string) *Skill {
	lower := strings.ToLower(name)
	for i := range skills {
		if strings.ToLower(skills[i].Name) == lower {
			return &skills[i]
		}
	}
	return nil
}

// loadAgentsMD discovers and returns the concatenated content of all AGENTS.md
// files relevant to the current working directory. It reads:
//  1. ~/.agents/AGENTS.md  (user-level, lowest precedence)
//  2. AGENTS.md files from the git root down to cwd (outer-to-inner order)
//
// Files that don't exist are silently skipped.
func loadAgentsMD() string {
	var parts []string

	// User-level
	if home, err := os.UserHomeDir(); err == nil {
		if data, err := os.ReadFile(filepath.Join(home, ".agents", "AGENTS.md")); err == nil {
			parts = append(parts, strings.TrimSpace(string(data)))
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return joinAgentsParts(parts)
	}

	// Walk from cwd up to find the git root (or stop at fs root).
	gitRoot := findGitRoot(cwd)

	// Collect directories from gitRoot down to cwd.
	dirs := ancestorDirs(gitRoot, cwd)

	for _, dir := range dirs {
		path := filepath.Join(dir, "AGENTS.md")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		parts = append(parts, strings.TrimSpace(string(data)))
	}

	return joinAgentsParts(parts)
}

// findGitRoot walks upward from dir looking for a .git entry.
// Returns dir itself if no .git is found.
func findGitRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

// ancestorDirs returns the chain of directories from root down to leaf,
// inclusive of both ends.
func ancestorDirs(root, leaf string) []string {
	root = filepath.Clean(root)
	leaf = filepath.Clean(leaf)

	// Build path from leaf upward until we hit root (or can't go further).
	var chain []string
	cur := leaf
	for {
		chain = append(chain, cur)
		if cur == root {
			break
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	// Reverse so order is root → leaf.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

func joinAgentsParts(parts []string) string {
	var nonempty []string
	for _, p := range parts {
		if p != "" {
			nonempty = append(nonempty, p)
		}
	}
	return strings.Join(nonempty, "\n\n")
}
