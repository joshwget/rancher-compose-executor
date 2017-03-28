package template

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

func Apply(contents []byte, variables map[string]interface{}) ([]byte, error) {
	// Skip templating if contents begin with '# notemplating'
	trimmedContents := strings.TrimSpace(string(contents))
	if strings.HasPrefix(trimmedContents, "#notemplating") || strings.HasPrefix(trimmedContents, "# notemplating") {
		return contents, nil
	}

	t, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(string(contents))
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	t.Execute(&buf, map[string]map[string]interface{}{
		"Values": variables,
	})
	return buf.Bytes(), nil
}
