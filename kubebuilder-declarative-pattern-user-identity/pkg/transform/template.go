package transform

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

type Params struct {
	Template *template.Template
}

type Param func(*Params)

// AsTemplate returns a declarative.ObjectTransform that expands a template.Template for each object from a source channel.
func AsTemplate(params ...Param) declarative.ObjectTransform {
	ps := &Params{
		Template: template.New(""),
	}
	for _, setter := range params {
		setter(ps)
	}
	return func(ctx context.Context, o declarative.DeclarativeObject, objects *manifest.Objects) error {
		for _, item := range objects.Items {
			b, err := item.JSON()
			if err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}
			te, err := ps.Template.Parse(string(b))
			if err != nil {
				return fmt.Errorf("parse template: %w", err)
			}
			var buf bytes.Buffer
			if err := te.Execute(&buf, o); err != nil {
				return fmt.Errorf("template: %w", err)
			}
			if bytes.Equal(buf.Bytes(), b) { // no changes
				continue
			}
			newItem, err := manifest.ParseJSONToObject(buf.Bytes()) // refresh the obj
			if err != nil {
				return fmt.Errorf("parse template failed: %w", err)
			}
			*item = *newItem
		}
		return nil
	}
}
