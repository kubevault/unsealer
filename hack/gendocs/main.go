package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/appscode/go/runtime"
	"github.com/spf13/cobra/doc"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"kubevault.dev/unsealer/pkg/cmds"
)

const (
	version = "0.2.0"
)

var (
	tplFrontMatter = template.Must(template.New("index").Parse(`---
title: Reference | Vault Unsealer
description: Vault Unsealer CLI Reference
menu:
  docs_{{ .Version }}:
    identifier: reference-unsealer
    name: Vault Unsealer
    weight: 40
    parent: reference
menu_name: docs_{{ .Version }}
---
`))

	_ = template.Must(tplFrontMatter.New("cmd").Parse(`---
title: {{ .Name }}
menu:
  docs_{{ .Version }}:
    identifier: {{ .ID }}
    name: {{ .Name }}
    parent: reference-unsealer
{{- if .RootCmd }}
    weight: 0
{{ end }}
menu_name: docs_{{ .Version }}
section_menu_id: reference
{{- if .RootCmd }}
url: /docs/{{ .Version }}/reference/unsealer/
aliases:
- /docs/{{ .Version }}/reference/unsealer/{{ .ID }}/
{{- end }}
---
`))
)

// ref: https://github.com/spf13/cobra/blob/master/doc/md_docs.md
func main() {
	rootCmd := cmds.NewRootCmd()
	dir := runtime.GOPath() + "/src/kubevault.dev/docs/docs/reference/unsealer"
	fmt.Printf("Generating cli markdown tree in: %v\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		data := struct {
			ID      string
			Name    string
			Version string
			RootCmd bool
		}{
			strings.Replace(base, "_", "-", -1),
			strings.Title(strings.Replace(base, "_", " ", -1)),
			version,
			!strings.ContainsRune(base, '_'),
		}
		var buf bytes.Buffer
		if err := tplFrontMatter.ExecuteTemplate(&buf, "cmd", data); err != nil {
			log.Fatalln(err)
		}
		return buf.String()
	}

	linkHandler := func(name string) string {
		return "/docs/reference/unsealer/" + name
	}
	utilruntime.Must(doc.GenMarkdownTreeCustom(rootCmd, dir, filePrepender, linkHandler))

	index := filepath.Join(dir, "_index.md")
	f, err := os.OpenFile(index, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	err = tplFrontMatter.ExecuteTemplate(f, "index", struct{ Version string }{version})
	if err != nil {
		log.Fatalln(err)
	}
	if err := f.Close(); err != nil {
		log.Fatalln(err)
	}
}
