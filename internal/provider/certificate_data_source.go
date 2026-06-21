package provider

import (
	"context"
	"fmt"

	"github.com/menscho/terraform-provider-certer/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CertificateDataSource{}
var _ datasource.DataSourceWithConfigure = &CertificateDataSource{}

type CertificateDataSource struct {
	client *client.Client
}

type CertificateDataSourceModel struct {
	CertificateID types.String   `tfsdk:"certificate_id"`
	Domain        types.String   `tfsdk:"domain"`
	Sans          []types.String `tfsdk:"sans"`
	Issued        types.Bool     `tfsdk:"issued"`
	Certificate   types.String   `tfsdk:"certificate"`
	PrivateKey    types.String   `tfsdk:"private_key"`
	CertFilename  types.String   `tfsdk:"cert_filename"`
	KeyFilename   types.String   `tfsdk:"key_filename"`
}

func NewCertificateDataSource() datasource.DataSource {
	return &CertificateDataSource{}
}

func (d *CertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_data"
}

func (d *CertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches certificate PEM data and private keys for a given certificate configuration ID (UUID) from certer.",
		Attributes: map[string]schema.Attribute{
			"certificate_id": schema.StringAttribute{
				MarkdownDescription: "The unique UUID identifier of the certificate configuration to fetch.",
				Required:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The primary domain name of the certificate.",
				Computed:            true,
			},
			"sans": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Subject Alternative Names (SANs) for this certificate.",
				Computed:            true,
			},
			"issued": schema.BoolAttribute{
				MarkdownDescription: "Whether the certificate has been issued and stored.",
				Computed:            true,
			},
			"certificate": schema.StringAttribute{
				MarkdownDescription: "The PEM-encoded certificate chain.",
				Computed:            true,
				Sensitive:           true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "The PEM-encoded private key.",
				Computed:            true,
				Sensitive:           true,
			},
			"cert_filename": schema.StringAttribute{
				MarkdownDescription: "The file name of the certificate in the storage directory.",
				Computed:            true,
			},
			"key_filename": schema.StringAttribute{
				MarkdownDescription: "The file name of the private key in the storage directory.",
				Computed:            true,
			},
		},
	}
}

func (d *CertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}

	d.client = c
}

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certs, err := d.client.GetCertificateData(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch certificate data: %s", err))
		return
	}

	found := false
	for _, c := range certs {
		if c.ID == data.CertificateID.ValueString() {
			found = true
			data.Domain = types.StringValue(c.Domain)
			sansVal := []types.String{}
			for _, s := range c.Sans {
				sansVal = append(sansVal, types.StringValue(s))
			}
			data.Sans = sansVal
			data.Issued = types.BoolValue(c.Issued)
			data.Certificate = types.StringValue(c.Certificate)
			data.PrivateKey = types.StringValue(c.PrivateKey)
			data.CertFilename = types.StringValue(c.CertFilename)
			data.KeyFilename = types.StringValue(c.KeyFilename)
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("Certificate Not Found", fmt.Sprintf("No certificate found for ID %q. Make sure it is configured and has been issued.", data.CertificateID.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
