package provider

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		{name: "opaque string ID", payload: `"image-uuid"`, want: "image-uuid"},
		{name: "null ID", payload: `null`, want: ""},
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

func TestParseRequiredNumericID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{name: "empty", value: "", wantErr: true},
		{name: "opaque", value: "vm-36", wantErr: true},
		{name: "negative", value: "-36", wantErr: true},
		{name: "fraction", value: "36.5", wantErr: true},
		{name: "exponent", value: "36e0", wantErr: true},
		{name: "overflow", value: "9223372036854775808", wantErr: true},
		{name: "valid decimal", value: "36", want: 36},
		{name: "zero is decimal", value: "0", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRequiredNumericID(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseRequiredNumericID(%q) succeeded with %d", tt.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseRequiredNumericID(%q) error = %v", tt.value, err)
			}
			if got != tt.want {
				t.Errorf("parseRequiredNumericID(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestRequiredCreateResponseIDRejectsEmpty(t *testing.T) {
	if _, err := requiredCreateResponseID(vm.APIID("")); err == nil {
		t.Fatal("requiredCreateResponseID(\"\") succeeded")
	}
	if got, err := requiredCreateResponseID(vm.APIID("36")); err != nil || got != "36" {
		t.Fatalf("requiredCreateResponseID(\"36\") = %q, %v; want 36, nil", got, err)
	}
}

func TestWaitForVMReadyWithEmptyIDDoesNotPoll(t *testing.T) {
	called := false
	err := waitForVMReady(context.Background(), time.Second, vm.APIID(""), func(context.Context, string) (*vm.InstanceReadResponse, error) {
		called = true
		return nil, nil
	})
	if err == nil {
		t.Fatal("waitForVMReady() succeeded for empty response ID")
	}
	if called {
		t.Fatal("waitForVMReady() called the API getter for an empty response ID")
	}
}

func TestAvailableSizesToTerraform(t *testing.T) {
	ctx := context.Background()

	valid := []vm.StringOrNumber{
		vm.StringOrNumber("71"),
		vm.StringOrNumber("9007199254740993"),
	}
	got, diags := availableSizesToTerraform(ctx, valid)
	if diags.HasError() {
		t.Fatalf("availableSizesToTerraform() diagnostics = %v", diags)
	}
	if got.IsNull() || got.IsUnknown() {
		t.Fatalf("availableSizesToTerraform() = %v, want known list", got)
	}
	if len(got.Elements()) != len(valid) {
		t.Fatalf("availableSizesToTerraform() has %d elements, want %d", len(got.Elements()), len(valid))
	}
	large, ok := got.Elements()[1].(types.Number)
	if !ok {
		t.Fatalf("element type = %T, want types.Number", got.Elements()[1])
	}
	wantLarge, _, err := big.ParseFloat("9007199254740993", 10, 256, big.ToNearestEven)
	if err != nil {
		t.Fatalf("parse expected large number: %v", err)
	}
	if large.ValueBigFloat().Cmp(wantLarge) != 0 {
		t.Errorf("large available size = %s, want %s", large.ValueBigFloat(), wantLarge)
	}

	nilList, nilDiags := availableSizesToTerraform(ctx, nil)
	if nilDiags.HasError() {
		t.Fatalf("nil available sizes diagnostics = %v", nilDiags)
	}
	if !nilList.IsNull() {
		t.Fatalf("nil available sizes = %v, want Terraform null", nilList)
	}

	emptyList, emptyDiags := availableSizesToTerraform(ctx, []vm.StringOrNumber{})
	if emptyDiags.HasError() {
		t.Fatalf("empty available sizes diagnostics = %v", emptyDiags)
	}
	if emptyList.IsNull() || len(emptyList.Elements()) != 0 {
		t.Fatalf("empty available sizes = %v, want known empty list", emptyList)
	}
}

func TestAvailableSizesToTerraformDiagnosesInvalidEntryIndex(t *testing.T) {
	_, diags := availableSizesToTerraform(context.Background(), []vm.StringOrNumber{
		vm.StringOrNumber("71"),
		vm.StringOrNumber(""),
	})
	if !diags.HasError() {
		t.Fatal("availableSizesToTerraform() returned no diagnostic for empty selected element")
	}
	foundPath := false
	for _, diagnostic := range diags {
		if withPath, ok := diagnostic.(diag.DiagnosticWithPath); ok && withPath.Path().String() == "available_sizes[1]" {
			foundPath = true
		}
	}
	if !foundPath {
		t.Fatalf("diagnostics = %v, want path available_sizes[1]", diags)
	}
}
