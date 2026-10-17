package tripo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

// Produces names what a successful Task's Output holds, so callers can say
// what to do next.
type Produces string

const (
	ProducesMesh         Produces = "mesh"
	ProducesImage        Produces = "image"
	ProducesViews        Produces = "views"
	ProducesSplat        Produces = "splat"
	ProducesRigCheck     Produces = "rig_check"
	ProducesSegmentation Produces = "segmentation"
)

// Operation is one task-creating Tripo endpoint with request type R.
//
// R's json tags are both the MCP argument names and Tripo's payload keys
// (ADR-0005); an `enum:"a,b"` tag constrains a field in the schema and in
// validation; Input fields accept local paths, which are uploaded. A field
// named model is resolved through the Catalog.
type Operation[R any] struct {
	Name        string
	Path        string
	Description string
	Produces    Produces
	// Free marks the rare Operation that creates a Task without spending
	// credits (rig_check).
	Free bool
	// check validates cross-field rules against the resolved model version.
	check func(req *R, model string) error
	// wire rewrites the encoded body into Tripo's wire form, if it differs.
	wire func(body map[string]any)
}

// Submission is a created Task.
type Submission struct {
	TaskID    string   `json:"task_id"`
	Operation string   `json:"operation"`
	Model     string   `json:"model,omitempty"`
	Produces  Produces `json:"produces"`
	Warnings  []string `json:"warnings,omitempty"`
}

// Submit validates req, uploads local-file Inputs and creates the Task. It
// costs credits once Tripo accepts it; everything that can be checked
// locally is checked before any upload or API call.
func (op Operation[R]) Submit(ctx context.Context, c *Client, req R) (*Submission, error) {
	if _, _, _, err := op.encode(&req); err != nil {
		return nil, err
	}
	if err := c.resolveInputs(ctx, &req); err != nil {
		return nil, err
	}
	body, model, warnings, err := op.encode(&req)
	if err != nil {
		return nil, err
	}
	if op.wire != nil {
		op.wire(body)
	}

	var created struct {
		TaskID string `json:"task_id"`
	}
	if err := c.call(ctx, http.MethodPost, op.Path, body, &created); err != nil {
		return nil, err
	}
	if created.TaskID == "" {
		return nil, fmt.Errorf("tripo: %s response has no task_id", op.Name)
	}
	return &Submission{TaskID: created.TaskID, Operation: op.Name, Model: model, Produces: op.Produces, Warnings: warnings}, nil
}

// encode turns a request into its JSON body, validated against the
// Operation's schema, with the model version resolved.
func (op Operation[R]) encode(req *R) (map[string]any, string, []string, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("tripo: encoding %s request: %w", op.Name, err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, "", nil, fmt.Errorf("tripo: encoding %s request: %w", op.Name, err)
	}
	resolved, err := resolvedSchema[R]()
	if err != nil {
		return nil, "", nil, err
	}
	if err := resolved.Validate(body); err != nil {
		return nil, "", nil, fmt.Errorf("tripo: invalid %s request: %w", op.Name, err)
	}
	for _, field := range requiredStrings(reflect.TypeFor[R]()) {
		if s, _ := body[field].(string); strings.TrimSpace(s) == "" {
			return nil, "", nil, fmt.Errorf("tripo: invalid %s request: %s is required", op.Name, field)
		}
	}
	var model string
	var warnings []string
	if _, hasModel := reflect.TypeFor[R]().FieldByName("Model"); hasModel {
		if model, warnings, err = applyModel(op.Name, body); err != nil {
			return nil, "", nil, err
		}
	}
	for field, operation := range catalogFields(reflect.TypeFor[R]()) {
		value, _ := body[field].(string)
		if value == "" {
			continue
		}
		if version, ok := lookupModel(operation, value); ok {
			body[field] = version.ID
		} else {
			warnings = append(warnings, fmt.Sprintf("%s %q is not in the built-in catalog (checked %s); sent to Tripo as-is", field, value, CatalogDate))
		}
	}
	if op.check != nil {
		if err := op.check(req, model); err != nil {
			return nil, "", nil, fmt.Errorf("tripo: invalid %s request: %w", op.Name, err)
		}
	}
	return body, model, warnings, nil
}

// InputSchema is the JSON Schema of R: properties and descriptions from its
// json and jsonschema tags, enums from its enum tags.
func (op Operation[R]) InputSchema() (*jsonschema.Schema, error) {
	return schemaFor(reflect.TypeFor[R]())
}

// SchemaFor builds the JSON Schema of a request type the same way
// Operation.InputSchema does.
func SchemaFor[T any]() (*jsonschema.Schema, error) {
	return schemaFor(reflect.TypeFor[T]())
}

func schemaFor(t reflect.Type) (*jsonschema.Schema, error) {
	s, err := jsonschema.ForType(t, &jsonschema.ForOptions{})
	if err != nil {
		return nil, fmt.Errorf("tripo: schema for %s: %w", t, err)
	}
	applyEnums(s, t)
	return s, nil
}

// catalogFields maps the top-level JSON fields of t tagged
// `catalog:"<operation>"` to that Operation; their values are resolved
// through the Catalog like model is.
func catalogFields(t reflect.Type) map[string]string {
	fields := map[string]string{}
	for _, field := range reflect.VisibleFields(t) {
		if op := field.Tag.Get("catalog"); op != "" {
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			fields[name] = op
		}
	}
	return fields
}

// requiredStrings lists the top-level string fields of t without omitempty,
// which the schema marks required but which must also be non-empty.
func requiredStrings(t reflect.Type) []string {
	var fields []string
	for _, field := range reflect.VisibleFields(t) {
		name, opts, _ := strings.Cut(field.Tag.Get("json"), ",")
		if field.Type.Kind() == reflect.String && name != "" && name != "-" && !strings.Contains(opts, "omitempty") {
			fields = append(fields, name)
		}
	}
	return fields
}

var resolvedSchemas sync.Map // reflect.Type → *jsonschema.Resolved

func resolvedSchema[R any]() (*jsonschema.Resolved, error) {
	t := reflect.TypeFor[R]()
	if r, ok := resolvedSchemas.Load(t); ok {
		return r.(*jsonschema.Resolved), nil
	}
	s, err := schemaFor(t)
	if err != nil {
		return nil, err
	}
	r, err := s.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("tripo: resolving schema for %s: %w", t, err)
	}
	resolvedSchemas.Store(t, r)
	return r, nil
}

// applyEnums copies `enum:"a,b,c"` struct tags onto the matching schema
// properties, following embedded and nested structs and slices.
func applyEnums(s *jsonschema.Schema, t reflect.Type) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if s.Items != nil {
			applyEnums(s.Items, t.Elem())
		}
	case reflect.Struct:
		for i := range t.NumField() {
			field := t.Field(i)
			if field.Anonymous {
				applyEnums(s, field.Type)
				continue
			}
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "" || name == "-" {
				continue
			}
			prop := s.Properties[name]
			if prop == nil {
				continue
			}
			if tag := field.Tag.Get("enum"); tag != "" {
				target := prop
				if prop.Items != nil {
					target = prop.Items
				}
				for _, v := range strings.Split(tag, ",") {
					target.Enum = append(target.Enum, v)
				}
			}
			applyEnums(prop, field.Type)
		}
	}
}
