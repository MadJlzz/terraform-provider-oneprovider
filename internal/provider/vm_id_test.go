package provider

import (
	"encoding/json"
	"testing"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUnitAPIIDJSONToTerraformString(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{name: "number ID", payload: `33`, want: "33"},
		{name: "string ID", payload: `"33"`, want: "33"},
		{name: "exactly above JavaScript safe integer", payload: `9007199254740993`, want: "9007199254740993"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id vm.APIID
			if err := json.Unmarshal([]byte(tt.payload), &id); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			got := types.StringValue(id.String())
			if got.ValueString() != tt.want {
				t.Errorf("types.StringValue(%q).ValueString() = %q, want %q", id.String(), got.ValueString(), tt.want)
			}
		})
	}
}

func TestUnitAPIIDImportIDsToVMInstanceResourceModel(t *testing.T) {
	var location vm.LocationReadResponse
	if err := json.Unmarshal([]byte(`{"id":33}`), &location); err != nil {
		t.Fatalf("unmarshal location: %v", err)
	}

	var size vm.SizeReadResponse
	if err := json.Unmarshal([]byte(`{"id":45}`), &size); err != nil {
		t.Fatalf("unmarshal size: %v", err)
	}

	var template vm.TemplateReadResponse
	if err := json.Unmarshal([]byte(`{"id":1194}`), &template); err != nil {
		t.Fatalf("unmarshal template: %v", err)
	}

	model := vmInstanceResourceModel{
		LocationId:     types.StringValue(location.Id.String()),
		InstanceSizeId: types.StringValue(size.Id.String()),
		TemplateId:     types.StringValue(template.Id.String()),
	}

	if got := model.LocationId.ValueString(); got != "33" {
		t.Errorf("LocationId = %q, want %q", got, "33")
	}
	if got := model.InstanceSizeId.ValueString(); got != "45" {
		t.Errorf("InstanceSizeId = %q, want %q", got, "45")
	}
	if got := model.TemplateId.ValueString(); got != "1194" {
		t.Errorf("TemplateId = %q, want %q", got, "1194")
	}
}
