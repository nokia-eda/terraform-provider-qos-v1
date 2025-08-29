package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nokia/eda/apps/terraform-provider-qos/internal/datasource_policy_attachment_list"
	"github.com/nokia/eda/apps/terraform-provider-qos/internal/eda/apiclient"
	"github.com/nokia/eda/apps/terraform-provider-qos/internal/tfutils"
)

const read_ds_policyAttachmentList = "/apps/qos.eda.nokia.com/v1/namespaces/{namespace}/policyattachments"

var (
	_ datasource.DataSource              = (*policyAttachmentListDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*policyAttachmentListDataSource)(nil)
)

func NewPolicyAttachmentListDataSource() datasource.DataSource {
	return &policyAttachmentListDataSource{}
}

type policyAttachmentListDataSource struct {
	client *apiclient.EdaApiClient
}

func (d *policyAttachmentListDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_attachment_list"
}

func (d *policyAttachmentListDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_policy_attachment_list.PolicyAttachmentListDataSourceSchema(ctx)
}

func (d *policyAttachmentListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_policy_attachment_list.PolicyAttachmentListModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract query params from Terraform model
	queryParams, err := tfutils.ModelToStringMap(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error extracting query params", err.Error())
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Read()::API request", map[string]any{
		"path":  read_ds_policyAttachmentList,
		"data":  spew.Sdump(data),
		"query": queryParams,
	})

	t0 := time.Now()
	result := map[string]any{}
	err = d.client.GetByQuery(ctx, read_ds_policyAttachmentList, map[string]string{
		"namespace": tfutils.StringValue(data.Namespace),
	}, queryParams, &result)

	tflog.Info(ctx, "Read()::API returned", map[string]any{
		"path":      read_ds_policyAttachmentList,
		"result":    spew.Sdump(result),
		"timeTaken": time.Since(t0).String(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Error reading resource", err.Error())
		return
	}

	// Convert API response to Terraform model
	err = tfutils.AnyMapToModel(ctx, result, &data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build response from API result", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (r *policyAttachmentListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.EdaApiClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.EdaApiClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}
