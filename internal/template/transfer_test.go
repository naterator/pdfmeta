package template

import (
	"bytes"
	"testing"

	"pdfmeta/internal/model"
)

func TestMarshalRecordsWrapsAndSorts(t *testing.T) {
	t.Parallel()
	payload, err := MarshalRecords([]model.TemplateRecord{{Name: "zeta"}, {Name: "alpha"}})
	if err != nil {
		t.Fatalf("MarshalRecords: %v", err)
	}
	if !bytes.Contains(payload, []byte(`"templates"`)) {
		t.Fatalf("expected wrapper object:\n%s", payload)
	}
	alpha := bytes.Index(payload, []byte(`"name": "alpha"`))
	zeta := bytes.Index(payload, []byte(`"name": "zeta"`))
	if alpha == -1 || zeta == -1 || alpha > zeta {
		t.Fatalf("expected sorted records:\n%s", payload)
	}
}

func TestUnmarshalRecordsSupportsWrapperAndArray(t *testing.T) {
	t.Parallel()

	wrapped, err := UnmarshalRecords([]byte(`{"templates":[{"name":"b"},{"name":"a"}]}`))
	if err != nil {
		t.Fatalf("UnmarshalRecords wrapped: %v", err)
	}
	if len(wrapped) != 2 || wrapped[0].Name != "a" || wrapped[1].Name != "b" {
		t.Fatalf("unexpected wrapped records: %+v", wrapped)
	}

	array, err := UnmarshalRecords([]byte(`[{"name":"d"},{"name":"c"}]`))
	if err != nil {
		t.Fatalf("UnmarshalRecords array: %v", err)
	}
	if len(array) != 2 || array[0].Name != "c" || array[1].Name != "d" {
		t.Fatalf("unexpected array records: %+v", array)
	}
}
