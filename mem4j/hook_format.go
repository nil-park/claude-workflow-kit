package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

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

func formatSessionContext(activeScopes []string, statePath string, insts []Instruction, allScopes []Scope) string {
	if len(insts) == 0 && len(allScopes) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("다음은 사용자 공유 기억 저장소(mem4j)에서 현재 세션의 활성 스코프 기준으로 조회한 정보이다.\n")
	fmt.Fprintf(&b, "활성 스코프: %s\n", strings.Join(activeScopes, ", "))
	if statePath != "" {
		fmt.Fprintf(&b, "세션 상태 파일: %s\n", filepath.ToSlash(statePath))
	}
	if len(insts) > 0 {
		b.WriteString("\n### 적용 중인 상시 지침\n\n")
		writeMerged(&b, insts)
	}
	if len(allScopes) > 0 {
		b.WriteString("### 등록된 전체 스코프 목록\n\n")
		writeScopes(&b, allScopes)
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatReminderContext(activeScopes []string, reminders []Reminder) string {
	if len(reminders) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("다음은 사용자 공유 기억 저장소(mem4j)에서 현재 프롬프트와 관련해 조회한 조건부 리마인더이다.\n")
	fmt.Fprintf(&b, "활성 스코프: %s\n\n", strings.Join(activeScopes, ", "))
	writeReminders(&b, reminders)
	return strings.TrimRight(b.String(), "\n")
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

func writeAgyPromptOutput(w io.Writer, msg string) error {
	if msg == "" {
		writeAgyEmptyOutput(w)
		return nil
	}
	return writeJSONLine(w, agyInjectOutput{
		InjectSteps: []agyInjectStep{
			{EphemeralMessage: msg},
		},
	})
}
