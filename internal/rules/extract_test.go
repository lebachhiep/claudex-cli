package rules

import "testing"

func TestTrackFileCounts(t *testing.T) {
	files := []string{
		"CLAUDE.md",
		".claude/skills/foo/SKILL.md",
		".claude/skills/bar/SKILL.md",
		".claude/skills/react-best-practices/rules/nested-1.md", // nested under skill, NOT a top-level rule
		".claude/skills/react-best-practices/rules/nested-2.md",
		".claude/agents/brainstormer.md",
		".claude/agents/debugger.md",
		".claude/agents/designer.md",
		".claude/rules/context-hygiene.md",
		".claude/rules/privacy-block-hook.md",
		".claude/rules/workflow-chaining.md",
		".claude/hooks/foo.cjs", // unrelated
	}

	stats := &FileStats{}
	for _, f := range files {
		trackFile(f, stats)
	}

	if stats.SkillCount != 2 {
		t.Errorf("SkillCount = %d, want 2", stats.SkillCount)
	}
	if stats.AgentCount != 3 {
		t.Errorf("AgentCount = %d, want 3 (top-level .md under .claude/agents/)", stats.AgentCount)
	}
	if stats.RuleCount != 3 {
		t.Errorf("RuleCount = %d, want 3 (top-level only, no nested skill rules)", stats.RuleCount)
	}
	if !stats.HasClaudeMD {
		t.Error("HasClaudeMD = false, want true")
	}
}
