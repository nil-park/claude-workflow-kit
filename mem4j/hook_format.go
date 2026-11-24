package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// claudeSessionSoftLimit leaves margin below Claude Code's 10,000-character
// additionalContext truncation limit.
const claudeSessionSoftLimit = 9000

type claudeHookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

type claudeHookOutput struct {
	HookSpecificOutput claudeHookSpecificOutput `json:"hookSpecificOutput"`
}

type agySystemMessageInner struct {
	SystemMessage string `json:"systemMessage"`
}

type agyInjectStep struct {
	SystemMessage    *agySystemMessageInner `json:"systemMessage,omitempty"`
	EphemeralMessage string                 `json:"ephemeralMessage,omitempty"`
}

type agyInjectOutput struct {
	InjectSteps []agyInjectStep `json:"injectSteps"`
}

func formatSessionContext(activeScopes []string, statePath string, insts []Instruction, allScopes []Scope, maxRunes int) string {
	if len(insts) == 0 && len(allScopes) == 0 {
		return ""
	}
	full := buildSessionContext(activeScopes, statePath, insts, allScopes, false)
	if maxRunes <= 0 || len(insts) == 0 || utf8.RuneCountInString(full) <= maxRunes {
		return full
	}
	return buildSessionContext(activeScopes, statePath, nil, allScopes, true)
}

func buildSessionContext(activeScopes []string, statePath string, insts []Instruction, allScopes []Scope, omittedInstructions bool) string {
	var b strings.Builder
	b.WriteString("다음은 사용자 공유 기억 저장소(mem4j)에서 현재 세션의 활성 스코프 기준으로 조회한 정보이다.\n")
	fmt.Fprintf(&b, "활성 스코프: %s\n", strings.Join(activeScopes, ", "))
	if statePath != "" {
		fmt.Fprintf(&b, "세션 상태 파일: %s\n", filepath.ToSlash(statePath))
	}
	if omittedInstructions {
		if statePath != "" {
			fmt.Fprintf(&b, "\n적용 중인 상시 지침이 컨텍스트 길이 제한을 넘어 생략되었다. `mem4j instruction recall -state %q`로 능동 조회한다.\n", filepath.ToSlash(statePath))
		} else {
			b.WriteString("\n적용 중인 상시 지침이 컨텍스트 길이 제한을 넘어 생략되었다. `mem4j instruction recall`로 능동 조회한다.\n")
		}
	} else if len(insts) > 0 {
		b.WriteString("\n### 적용 중인 상시 지침\n\n")
		writeMerged(&b, insts)
	}
	if len(allScopes) > 0 {
		b.WriteString("\n### 등록된 전체 스코프 목록\n\n")
		writeScopes(&b, allScopes)
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatPromptContext(activeScopes []string, reminders []Reminder, statements []Statement, events []Event) string {
	if len(reminders) == 0 && len(statements) == 0 && len(events) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("다음은 사용자 공유 기억 저장소(mem4j)에서 현재 프롬프트와 관련해 조회한 기억이다.\n")
	if len(activeScopes) > 0 {
		fmt.Fprintf(&b, "활성 스코프: %s\n", strings.Join(activeScopes, ", "))
	}
	if len(reminders) > 0 {
		b.WriteString("\n### 조건부 리마인더\n\n")
		writeReminders(&b, reminders)
	}
	if len(statements) > 0 {
		b.WriteString("\n### 관련 사실\n\n")
		writeStatements(&b, statements, true)
	}
	if len(events) > 0 {
		b.WriteString("\n### 관련 이력\n\n")
		writeEvents(&b, events, true)
	}
	return strings.TrimRight(b.String(), "\n")
}

func recordPromptMatches(entry *hookLogEntry, reminders []Reminder, statements []Statement, events []Event) {
	for _, r := range reminders {
		entry.MatchedKeys = append(entry.MatchedKeys, "reminder:"+r.Key)
		if r.Score > entry.TopScore {
			entry.TopScore = r.Score
		}
	}
	for _, s := range statements {
		entry.MatchedKeys = append(entry.MatchedKeys, "statement:"+s.Key)
		if s.Score > entry.TopScore {
			entry.TopScore = s.Score
		}
	}
	for _, e := range events {
		entry.MatchedKeys = append(entry.MatchedKeys, "event:"+e.Key)
		if e.Score > entry.TopScore {
			entry.TopScore = e.Score
		}
	}
}

func writeJSONLine(w io.Writer, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func writeClaudeHookOutput(w io.Writer, eventName, additionalContext string) error {
	if additionalContext == "" {
		return nil
	}
	return writeJSONLine(w, claudeHookOutput{
		HookSpecificOutput: claudeHookSpecificOutput{
			HookEventName:     eventName,
			AdditionalContext: additionalContext,
		},
	})
}

func writeAgyEmptyOutput(w io.Writer) {
	_, _ = io.WriteString(w, "{}\n")
}

func writeAgySessionOutput(w io.Writer, msg string) error {
	if msg == "" {
		writeAgyEmptyOutput(w)
		return nil
	}
	return writeJSONLine(w, agyInjectOutput{
		InjectSteps: []agyInjectStep{
			{SystemMessage: &agySystemMessageInner{SystemMessage: msg}},
		},
	})
}

func writeAgyStepsOutput(w io.Writer, sessionMsg, promptMsg string) error {
	steps := make([]agyInjectStep, 0, 2)
	if sessionMsg != "" {
		steps = append(steps, agyInjectStep{
			SystemMessage: &agySystemMessageInner{SystemMessage: sessionMsg},
		})
	}
	if promptMsg != "" {
		steps = append(steps, agyInjectStep{EphemeralMessage: promptMsg})
	}
	if len(steps) == 0 {
		writeAgyEmptyOutput(w)
		return nil
	}
	return writeJSONLine(w, agyInjectOutput{InjectSteps: steps})
}
