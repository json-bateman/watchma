package utils

import (
	"regexp"
	"strings"
)

type StoryLineType string

const (
	Dialogue       StoryLineType = "dialogue"
	Narration      StoryLineType = "narration"
	StageDirection StoryLineType = "stage_direction"
)

type StoryLine struct {
	Type    StoryLineType
	Speaker string // Only for dialogue
	Content string
	Index   int // For ordering
}

// ParseStory splits AI response into structured story lines
func ParseStory(aiResponse string) []StoryLine {
	var lines []StoryLine
	paragraphs := strings.Split(aiResponse, "\n\n")

	// Regex patterns
	// Matches: **Jordan Belfort (played by Leonardo DiCaprio):** or **Jordan Belfort:**
	dialoguePattern := regexp.MustCompile(`\*\*([^:]+?)(?:\s*\([^)]+\))?\s*:\*\*\s*\*"([^"]+)"\*`)
	// Matches: *text in italics*
	stagePattern := regexp.MustCompile(`^\*(.+?)\*$`)

	for i, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Check if it's dialogue
		if matches := dialoguePattern.FindStringSubmatch(p); matches != nil {
			lines = append(lines, StoryLine{
				Type:    Dialogue,
				Speaker: strings.TrimSpace(matches[1]),
				Content: matches[2],
				Index:   i,
			})
		} else if matches := stagePattern.FindStringSubmatch(p); matches != nil {
			// Stage direction
			lines = append(lines, StoryLine{
				Type:    StageDirection,
				Content: strings.TrimSpace(matches[1]),
				Index:   i,
			})
		} else {
			// Plain narration (strip ** markers if present)
			cleaned := strings.ReplaceAll(p, "**", "")
			cleaned = strings.ReplaceAll(cleaned, "*", "")
			lines = append(lines, StoryLine{
				Type:    Narration,
				Content: strings.TrimSpace(cleaned),
				Index:   i,
			})
		}
	}

	return lines
}
