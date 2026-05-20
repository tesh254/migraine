package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tesh254/migraine/internal/skill"
	"github.com/tesh254/migraine/internal/ui"
)

var skillCmd = &cobra.Command{
	Use:     "skill",
	Aliases: []string{"sk"},
	Short:   "Manage workflow skills (pre-built workflow templates)",
	Long: `Manage workflow skills — pre-built workflow templates for common tasks.

Skills are ready-to-use workflow definitions that can be installed globally
or per-project. They cover common patterns like deployments, CI pipelines,
and Git workflows.

Installation scopes:
  --global   Install to ~/.migraine/skills/ (available everywhere)
  --project  Install to ./.migraine/skills/ (available in this project only)

For non-interactive / agent use, specify a skill name and scope directly:
  migraine skill add docker-deploy --global
  migraine skill add docker-deploy --project

Agent integration:
  migraine skill init --agent opencode
  migraine skill init --agent claude
  migraine skill init --agent codex`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and installed skills",
	Long: `List all available built-in skills and any currently installed skills.

Shows skill name, category, description, and tags for each available skill.
Installed skills are marked with their scope (global or project).`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.SectionHeader("Available Skills")
		skills := skill.List()
		if len(skills) == 0 {
			fmt.Println("  No skills available.")
			return
		}

		categories := make(map[string][]skill.Skill)
		for _, s := range skills {
			categories[s.Category] = append(categories[s.Category], s)
		}

		for cat, catSkills := range categories {
			fmt.Printf("\n  %s:\n", strings.Title(cat))
			for _, s := range catSkills {
				tags := ""
				if len(s.Tags) > 0 {
					tags = fmt.Sprintf(" [%s]", strings.Join(s.Tags, ", "))
				}
				fmt.Printf("    %-20s %s%s\n", s.Name, s.Description, tags)
			}
		}

		installed, err := skill.ListInstalled()
		if err == nil && len(installed) > 0 {
			ui.SectionHeader("Installed")
			for _, i := range installed {
				fmt.Printf("  • %s\n", i)
			}
		}

		fmt.Println("\n  Use 'migraine skill add <name> --global|--project' to install a skill.")
	},
}

var skillAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Install a skill (interactive if no flags)",
	Long: `Install a workflow skill.

If --global or --project is specified, installs non-interactively.
If no scope flag is given, prompts for the scope interactively.

For agents and scripts:
  migraine skill add docker-deploy --global
  migraine skill add docker-deploy --project

For interactive use:
  migraine skill add docker-deploy`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		globalFlag, _ := cmd.Flags().GetBool("global")
		projectFlag, _ := cmd.Flags().GetBool("project")

		scope := ""
		if globalFlag {
			scope = "global"
		} else if projectFlag {
			scope = "project"
		} else {
			fmt.Printf("Install '%s' globally or per-project? [global/project]: ", name)
			input := strings.TrimSpace(strings.ToLower(readLine()))
			switch input {
			case "global", "g":
				scope = "global"
			case "project", "p", "local":
				scope = "project"
			default:
				fmt.Println("  Defaulting to project scope.")
				scope = "project"
			}
		}

		return skill.Install(name, scope)
	},
}

var skillRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove an installed skill",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return skill.Remove(args[0])
	},
}

var skillShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show skill details and .mg format preview",
	Long: `Show detailed information about a skill, including its full workflow
definition and a preview of the .mg format output.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, ok := skill.Find(args[0])
		if !ok {
			return fmt.Errorf("skill '%s' not found. Use 'migraine skill list' to see available skills", args[0])
		}

		mgFlag, _ := cmd.Flags().GetBool("mg")

		ui.SectionHeader(fmt.Sprintf("Skill: %s", s.Name))
		fmt.Printf("  Description: %s\n", s.Description)
		fmt.Printf("  Category:    %s\n", s.Category)
		if len(s.Tags) > 0 {
			fmt.Printf("  Tags:        %s\n", strings.Join(s.Tags, ", "))
		}
		if len(s.Variables) > 0 {
			fmt.Printf("  Variables:   %s\n", strings.Join(mapKeys(s.Variables), ", "))
		}

		fmt.Printf("\n  Pre-checks: %d\n", len(s.Workflow.PreChecks))
		fmt.Printf("  Steps:       %d\n", len(s.Workflow.Steps))
		fmt.Printf("  Actions:     %d\n", len(s.Workflow.Actions))

		if mgFlag {
			fmt.Printf("\n--- .mg format ---\n%s\n--- end ---\n", skill.RenderMG(s))
		}

		return nil
	},
}

var skillSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available skills by name, category, or tag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		results := skill.Search(args[0])
		if len(results) == 0 {
			fmt.Printf("No skills matching '%s' found.\n", args[0])
			return
		}

		ui.SectionHeader(fmt.Sprintf("Search: %s", args[0]))
		for _, s := range results {
			tags := ""
			if len(s.Tags) > 0 {
				tags = fmt.Sprintf(" [%s]", strings.Join(s.Tags, ", "))
			}
			fmt.Printf("  %-20s %s%s\n", s.Name, s.Description, tags)
		}
	},
}

var skillInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up agent integration for migraine skills",
	Long: `Configure an AI coding agent to recognize migraine skills and workflows.

Generates the appropriate config file for the agent so it knows about
available skills and how to invoke them.

Supported agents: opencode, claude, codex, cursor, windsurf, cline

For non-interactive / agent use, specify --agent directly:
  migraine skill init --agent opencode

For interactive use:
  migraine skill init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		agentFlag, _ := cmd.Flags().GetString("agent")

		if agentFlag != "" {
			agentConfig, ok := skill.FindAgent(agentFlag)
			if !ok {
				return fmt.Errorf("unsupported agent '%s'. Supported: %s", agentFlag, skill.AgentList())
			}

			allSkills := skill.List()
			installed, _ := skill.ListInstalled()

			var skillsToInclude []skill.Skill
			if len(installed) > 0 {
				skillsToInclude = allSkills
			} else {
				skillsToInclude = allSkills
			}

			return skill.SetupAgent(agentConfig.Name, skillsToInclude)
		}

		fmt.Println("Which agent would you like to configure?")
		fmt.Println()
		agents := skill.SupportedAgents()
		for i, a := range agents {
			fmt.Printf("  %d. %s (%s)\n", i+1, a.DisplayName, a.ConfigFile)
		}
		fmt.Println()

		fmt.Print("Select agent [number or name]: ")
		input := strings.TrimSpace(strings.ToLower(readLine()))

		var selected *skill.AgentConfig
		for i, a := range agents {
			if input == fmt.Sprintf("%d", i+1) || input == a.Name {
				selected = &agents[i]
				break
			}
		}

		if selected == nil {
			return fmt.Errorf("invalid selection. Supported agents: %s", skill.AgentList())
		}

		return skill.SetupAgent(selected.Name, skill.List())
	},
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	skillAddCmd.Flags().Bool("global", false, "Install skill globally (~/.migraine/skills/)")
	skillAddCmd.Flags().Bool("project", false, "Install skill per-project (./.migraine/skills/)")

	skillShowCmd.Flags().Bool("mg", false, "Show skill in .mg format")

	skillInitCmd.Flags().String("agent", "", "Agent to configure (opencode, claude, codex, cursor, windsurf, cline)")

	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillAddCmd)
	skillCmd.AddCommand(skillRemoveCmd)
	skillCmd.AddCommand(skillShowCmd)
	skillCmd.AddCommand(skillSearchCmd)
	skillCmd.AddCommand(skillInitCmd)

	rootCmd.AddCommand(skillCmd)
}