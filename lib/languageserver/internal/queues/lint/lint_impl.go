package lint

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/filesystem"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"

	parsing "github.com/glennsarti/sentinel-utils/lib/parsing/default"
	cwalker "github.com/glennsarti/sentinel-utils/lib/walkers/sentinel_config"

	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues/generic"
	"github.com/glennsarti/sentinel-utils/lib/linting"
)

var _ queues.LintQueue = &lintQueue{}

func NewLintQueue(
	queueSize int,
	rootUri lsp.DocumentURI,
	fsys filesystem.SessionFS,
	dispatchQueue queues.ClientNotifyDispatchQueue,
	logger *log.Logger,
) (queues.LintQueue, error) {
	lq := &lintQueue{
		logger:          logger,
		fsys:            fsys,
		rootUri:         rootUri,
		dispatchQueue:   dispatchQueue,
		issueIndex:      0,
		filesWithIssues: make(map[string]int, 0),
	}
	lq.baseq = generic.NewGenericQueue(1, queueSize, lq.process)

	return lq, nil
}

type lintQueue struct {
	logger          *log.Logger
	baseq           *generic.GenericQueue[queues.LintQueueRequest]
	fsys            filesystem.SessionFS
	rootUri         lsp.DocumentURI
	dispatchQueue   queues.ClientNotifyDispatchQueue
	muWriter        sync.Mutex
	issueIndex      int
	filesWithIssues map[string]int
}

type allIssues = map[string]slint.Issues

func (lq *lintQueue) Enqueue(req queues.LintQueueRequest) error {
	lq.baseq.Enqueue(req)
	return nil
}
func (lq *lintQueue) Start(ctx context.Context) error      { return lq.baseq.Start(ctx) }
func (lq *lintQueue) StartAsync(ctx context.Context) error { return lq.baseq.StartAsync(ctx) }
func (lq *lintQueue) Stop()                                {}
func (lq *lintQueue) Name() string                         { return "lintQueue" }
func (lq *lintQueue) Logger() *log.Logger                  { return lq.logger }

func (lq *lintQueue) process(job queues.LintQueueRequest) error {
	rootPath, err := lq.fsys.UriToPath(lq.rootUri)
	if err != nil {
		return errors.New("failed to convert root URI to path: " + err.Error())
	}

	// Expensive but fine
	pf := parsing.NewDefaultParsingFactory(lq.fsys)
	walker := cwalker.NewSentinelConfigWalker(
		lq.fsys,
		rootPath,
		job.SentinelVersion,
		pf,
	)
	if walker == nil {
		return errors.New("failed to create walker")
	}

	issuesList := make(allIssues, 0)

	if err := linting.Lint(walker, pf, func(lintFile slint.File, issues slint.Issues) {
		if len(issues) > 0 {
			if _, ok := issuesList[lintFile.Path()]; ok {
				issuesList[lintFile.Path()] = append(issuesList[lintFile.Path()], issues...)
			} else {
				issuesList[lintFile.Path()] = issues
			}
		}
	}); err != nil {
		return err
	}

	if err := lq.sendIssues(&issuesList); err != nil {
		return err
	}

	return nil
}

func (lq *lintQueue) sendIssues(issues *allIssues) error {
	lq.muWriter.Lock()
	defer lq.muWriter.Unlock()

	lq.issueIndex++

	for filePath, fileIssues := range *issues {
		fileUri, err := lq.fsys.PathToUri(filePath)
		if err != nil {
			return err
		}

		lq.filesWithIssues[string(fileUri)] = lq.issueIndex

		resp := lsp.PublishDiagnosticsParams{
			URI:         fileUri,
			Diagnostics: make([]lsp.Diagnostic, len(fileIssues)),
		}

		for idx, issue := range fileIssues {
			if issue != nil {
				resp.Diagnostics[idx] = lq.toDiagnostic(*issue)
			}
		}

		if err := lq.dispatchQueue.Enqueue(queues.ClientNotifyDispatchRequest{
			Method: "textDocument/publishDiagnostics",
			Params: resp,
		}); err != nil {
			return err
		}
	}

	for uri, i := range lq.filesWithIssues {
		if i != lq.issueIndex {
			if err := lq.dispatchQueue.Enqueue(queues.ClientNotifyDispatchRequest{
				Method: "textDocument/publishDiagnostics",
				Params: lsp.PublishDiagnosticsParams{
					URI:         lsp.DocumentURI(uri),
					Diagnostics: []lsp.Diagnostic{},
				},
			}); err != nil {
				return err
			}
			delete(lq.filesWithIssues, uri)
		}
	}

	return nil
}

func (lq *lintQueue) toDiagnostic(issue slint.Issue) lsp.Diagnostic {
	d := lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      uint32(issue.Range.Start.Line),
				Character: uint32(issue.Range.Start.Column),
			},
			End: lsp.Position{
				Line:      uint32(issue.Range.End.Line),
				Character: uint32(issue.Range.End.Column),
			},
		},
		Message:  issue.Detail,
		Code:     issue.RuleId,
		Source:   "sentinel-lint",
		Severity: lsp.SeverityError,
	}

	switch issue.Severity {
	case slint.Information:
		d.Severity = lsp.SeverityInformation
	case slint.Warning:
		d.Severity = lsp.SeverityWarning
	}

	if issue.Related != nil {
		d.RelatedInformation = make([]lsp.DiagnosticRelatedInformation, len(*issue.Related))
		for idx, rv := range *issue.Related {
			relatedUri, _ := lq.fsys.PathToUri(rv.Range.Filename)
			d.RelatedInformation[idx] = lsp.DiagnosticRelatedInformation{
				Message: rv.Summary,
				Location: lsp.Location{
					URI: relatedUri,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      uint32(rv.Range.Start.Line),
							Character: uint32(rv.Range.Start.Column),
						},
						End: lsp.Position{
							Line:      uint32(rv.Range.End.Line),
							Character: uint32(rv.Range.End.Column),
						},
					},
				},
			}
		}
	}

	return d
}
