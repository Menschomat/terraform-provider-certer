package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProvider_Metadata(t *testing.T) {
	p := New("test")()
	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "certer" {
		t.Errorf("Expected TypeName 'certer', got %q", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Errorf("Expected Version 'test', got %q", resp.Version)
	}
}
