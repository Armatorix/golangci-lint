package printers

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"

	"github.com/golangci/golangci-lint/pkg/logutils"
	"github.com/golangci/golangci-lint/pkg/result"
)

// CodeClimateIssue is a subset of the Code Climate spec - https://github.com/codeclimate/spec/blob/master/SPEC.md#data-types
// It is just enough to support GitLab CI Code Quality - https://docs.gitlab.com/ee/user/project/merge_requests/code_quality.html
type CodeClimateIssue struct {
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Fingerprint string `json:"fingerprint"`
	Location    struct {
		Path  string `json:"path"`
		Lines struct {
			Begin int `json:"begin"`
		} `json:"lines"`
	} `json:"location"`
}

type CodeClimate struct {
}

func NewCodeClimate() *CodeClimate {
	return &CodeClimate{}
}

func (p CodeClimate) Print(ctx context.Context, issues []result.Issue) error {
	allIssues := []CodeClimateIssue{}
	for ind := range issues {
		i := &issues[ind]
		var issue CodeClimateIssue
		issue.Description = i.FromLinter + ": " + i.Text
		issue.Location.Path = i.Pos.Filename
		issue.Location.Lines.Begin = i.Pos.Line

		if i.Severity != "" {
			issue.Severity = i.Severity
		}

		// Need a checksum of the issue, so we use MD5 of the filename, text, and first line of source if there is any
		var firstLine string
		if len(i.SourceLines) > 0 {
			firstLine = i.SourceLines[0]
		}

		hash := md5.New() //nolint:gosec
		_, _ = hash.Write([]byte(i.Pos.Filename + i.Text + firstLine))
		issue.Fingerprint = fmt.Sprintf("%X", hash.Sum(nil))

		allIssues = append(allIssues, issue)
	}

	outputJSON, err := json.Marshal(allIssues)
	if err != nil {
		return err
	}

	fmt.Fprint(logutils.StdOut, string(outputJSON))
	return nil
}
