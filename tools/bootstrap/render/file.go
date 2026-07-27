package render

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// binaryExts are copied as-is (no content templating).
var binaryExts = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".ico": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".eot": {}, ".otf": {},
	".mp4": {}, ".webm": {}, ".pdf": {}, ".zip": {}, ".gz": {}, ".tar": {},
	".bin": {}, ".exe": {}, ".dll": {}, ".so": {}, ".dylib": {}, ".wasm": {},
}

func isBinaryPath(path string) bool {
	_, ok := binaryExts[strings.ToLower(filepath.Ext(path))]
	return ok
}

// gateData is gate-mode execute root: only Mode (no ModulePath / names).
type gateData struct {
	Mode Mode
}

// File runs gate+normal pipeline for one file; skip => skipped=true.
func File(path string, src []byte, ctx Context) (out []byte, skipped bool, err error) {
	if isBinaryPath(path) {
		return append([]byte(nil), src...), false, nil
	}

	line1, rest, hasNL := splitFirstLine(src)
	gated, include, err := evalGateLine(path, line1, ctx.Mode)
	if err != nil {
		return nil, false, err
	}
	body := src
	if gated {
		if !include {
			return nil, true, nil
		}
		if hasNL {
			body = rest
		} else {
			body = nil
		}
	}

	rendered, err := renderNormal(path, body, ctx)
	if err != nil {
		return nil, false, err
	}
	return rendered, false, nil
}

func splitFirstLine(src []byte) (line1 string, rest []byte, hasNL bool) {
	i := bytes.IndexByte(src, '\n')
	if i < 0 {
		return string(src), nil, false
	}
	line := strings.TrimSuffix(string(src[:i]), "\r")
	return line, src[i+1:], true
}

type gateResult struct {
	called  bool
	include bool
}

func (g *gateResult) when(args ...any) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("when: want exactly one bool argument, got %d", len(args))
	}
	b, isBool := args[0].(bool)
	if !isBool {
		return "", fmt.Errorf("when: argument is %T, want bool", args[0])
	}
	g.called = true
	g.include = b
	return "", nil
}

func modeIsHelper(mode Mode) func(...string) bool {
	return func(allowed ...string) bool {
		for _, a := range allowed {
			if a == string(mode) {
				return true
			}
		}
		return false
	}
}

func evalGateLine(path, line1 string, mode Mode) (gated bool, include bool, err error) {
	if !strings.Contains(line1, "[[") {
		return false, true, nil
	}

	actions, err := extractTemplateActions(line1)
	if err != nil {
		return false, false, fmt.Errorf("render gate %s: %w", path, err)
	}

	var gates []string
	for _, inner := range actions {
		name := firstActionIdent(inner)
		switch {
		case name == "when":
			gates = append(gates, inner)
		case looksLikeGateAttempt(name):
			return false, false, fmt.Errorf("render gate %s: invalid gate action %q", path, strings.TrimSpace(inner))
		}
	}

	switch len(gates) {
	case 0:
		// Plain [[…]]-only line 1 (no comment wrapper) that isn't a gate → hard error.
		if lineIsOnlyTemplateActions(line1, actions) {
			return false, false, fmt.Errorf("render gate %s: line must use when", path)
		}
		return false, true, nil
	case 1:
		return evalGateAction(path, gates[0], mode)
	default:
		return false, false, fmt.Errorf("render gate %s: multiple gate actions on line 1", path)
	}
}

// extractTemplateActions returns inner text of each [[…]] on line (left-to-right).
func extractTemplateActions(line string) ([]string, error) {
	var out []string
	i := 0
	for i < len(line) {
		start := strings.Index(line[i:], "[[")
		if start < 0 {
			break
		}
		start += i
		endRel := strings.Index(line[start+2:], "]]")
		if endRel < 0 {
			return nil, fmt.Errorf("unclosed [[")
		}
		out = append(out, line[start+2:start+2+endRel])
		i = start + 2 + endRel + 2
	}
	return out, nil
}

// firstActionIdent returns the leading function/pipeline identifier inside a [[…]] action.
func firstActionIdent(inner string) string {
	s := strings.TrimSpace(inner)
	if strings.HasPrefix(s, "-") {
		s = strings.TrimSpace(s[1:])
	}
	end := 0
	for end < len(s) {
		c := s[end]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			end++
			continue
		}
		break
	}
	return s[:end]
}

func looksLikeGateAttempt(name string) bool {
	return strings.HasPrefix(name, "when")
}

func lineIsOnlyTemplateActions(line string, actions []string) bool {
	if len(actions) == 0 {
		return false
	}
	rest := line
	for _, inner := range actions {
		token := "[[" + inner + "]]"
		idx := strings.Index(rest, token)
		if idx < 0 {
			return false
		}
		rest = rest[:idx] + rest[idx+len(token):]
	}
	return strings.TrimSpace(rest) == ""
}

func evalGateAction(path, inner string, mode Mode) (gated bool, include bool, err error) {
	gr := &gateResult{}
	funcs := template.FuncMap{
		"when":   gr.when,
		"modeIs": modeIsHelper(mode),
	}
	// Authors must nest: when (modeIs …). No rewrite of bare when modeIs.
	src := "[[" + inner + "]]"
	tmpl, err := template.New("gate:"+path).Delims("[[", "]]").Funcs(funcs).Parse(src)
	if err != nil {
		return false, false, fmt.Errorf("render gate %s: %w", path, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, gateData{Mode: mode}); err != nil {
		return false, false, fmt.Errorf("render gate %s: %w", path, err)
	}
	if !gr.called {
		return false, false, fmt.Errorf("render gate %s: line must use when", path)
	}
	if strings.TrimSpace(buf.String()) != "" {
		return false, false, fmt.Errorf("render gate %s: unexpected gate output %q", path, buf.String())
	}
	return true, gr.include, nil
}

func renderNormal(path string, body []byte, ctx Context) ([]byte, error) {
	// Body FuncMap: modeIs only. when / whenEither / whenMode absent → execute error.
	funcs := template.FuncMap{
		"modeIs": modeIsHelper(ctx.Mode),
	}
	tmpl, err := template.New("file:"+path).Delims("[[", "]]").Funcs(funcs).Parse(string(body))
	if err != nil {
		return nil, fmt.Errorf("render %s: %w", path, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("render %s: %w", path, err)
	}
	return buf.Bytes(), nil
}
