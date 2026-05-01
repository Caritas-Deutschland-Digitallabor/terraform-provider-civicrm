package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &TagsetResource{}
	_ resource.ResourceWithConfigure   = &TagsetResource{}
	_ resource.ResourceWithImportState = &TagsetResource{}
)

type TagsetResource struct {
	client *Client
}

type TagsetResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Label       types.String `tfsdk:"label"`
	Description types.String `tfsdk:"description"`
	UsedFor     types.List   `tfsdk:"used_for"`
	Color       types.String `tfsdk:"color"`
}

func NewTagsetResource() resource.Resource {
	return &TagsetResource{}
}

func (r *TagsetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tagset"
}

func (r *TagsetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CiviCRM Tagset. A Tagset is a named container for tags (is_tagset=true). Tags belonging to a tagset reference it via parent_id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the tagset.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The machine name of the tagset (must be unique, no spaces).",
				Required:    true,
			},
			"label": schema.StringAttribute{
				Description: "The display label of the tagset.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the tagset.",
				Optional:    true,
			},
			"used_for": schema.ListAttribute{
				Description: "Entity types this tagset can be used for (e.g., 'civicrm_contact', 'civicrm_activity').",
				Optional:    true,
				ElementType: types.StringType,
			},
			"color": schema.StringAttribute{
				Description: "The color for the tagset in hex format (e.g., '#ff0000').",
				Optional:    true,
			},
		},
	}
}

func (r *TagsetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *TagsetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TagsetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating tagset", map[string]any{"name": plan.Name.ValueString()})

	values := map[string]any{
		"name":      plan.Name.ValueString(),
		"is_tagset": true,
	}

	if !plan.Label.IsNull() {
		values["label"] = plan.Label.ValueString()
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	}

	if !plan.UsedFor.IsNull() {
		var usedFor []string
		diags = plan.UsedFor.ElementsAs(ctx, &usedFor, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values["used_for"] = usedFor
	}

	if !plan.Color.IsNull() {
		values["color"] = plan.Color.ValueString()
	}

	result, err := r.client.Create("Tag", values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tagset",
			"Could not create tagset: "+err.Error(),
		)
		return
	}

	var d diag.Diagnostics
	r.mapResultToState(ctx, result, &plan, &d)
	resp.Diagnostics.Append(d...)

	tflog.Debug(ctx, "Created tagset", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *TagsetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TagsetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading tagset", map[string]any{"id": state.ID.ValueInt64()})

	result, err := r.client.GetByID("Tag", state.ID.ValueInt64(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tagset",
			"Could not read tagset ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	var d diag.Diagnostics
	r.mapResultToState(ctx, result, &state, &d)
	resp.Diagnostics.Append(d...)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *TagsetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TagsetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state TagsetResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating tagset", map[string]any{"id": state.ID.ValueInt64()})

	values := map[string]any{
		"name":      plan.Name.ValueString(),
		"is_tagset": true,
	}

	if !plan.Label.IsNull() {
		values["label"] = plan.Label.ValueString()
	} else {
		values["label"] = nil
	}

	if !plan.Description.IsNull() {
		values["description"] = plan.Description.ValueString()
	} else {
		values["description"] = nil
	}

	if !plan.UsedFor.IsNull() {
		var usedFor []string
		diags = plan.UsedFor.ElementsAs(ctx, &usedFor, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		values["used_for"] = usedFor
	} else {
		values["used_for"] = nil
	}

	if !plan.Color.IsNull() {
		values["color"] = plan.Color.ValueString()
	} else {
		values["color"] = nil
	}

	result, err := r.client.Update("Tag", state.ID.ValueInt64(), values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating tagset",
			"Could not update tagset ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	plan.ID = state.ID
	var d diag.Diagnostics
	r.mapResultToState(ctx, result, &plan, &d)
	resp.Diagnostics.Append(d...)

	tflog.Debug(ctx, "Updated tagset", map[string]any{"id": plan.ID.ValueInt64()})

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *TagsetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TagsetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting tagset", map[string]any{"id": state.ID.ValueInt64()})

	err := r.client.Delete("Tag", state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting tagset",
			"Could not delete tagset ID "+strconv.FormatInt(state.ID.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted tagset", map[string]any{"id": state.ID.ValueInt64()})
}

func (r *TagsetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Could not parse import ID as integer: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func (r *TagsetResource) mapResultToState(ctx context.Context, result map[string]any, model *TagsetResourceModel, diags *diag.Diagnostics) {
	if id, ok := GetInt64(result, "id"); ok {
		model.ID = types.Int64Value(id)
	}

	if name, ok := GetString(result, "name"); ok {
		model.Name = types.StringValue(name)
	}

	if label, ok := GetString(result, "label"); ok && label != "" {
		model.Label = types.StringValue(label)
	} else {
		if name, ok := GetString(result, "name"); ok {
			model.Label = types.StringValue(name)
		}
	}

	if desc, ok := GetString(result, "description"); ok && desc != "" {
		model.Description = types.StringValue(desc)
	} else {
		model.Description = types.StringNull()
	}

	if usedForRaw, ok := result["used_for"]; ok && usedForRaw != nil {
		if usedForSlice, ok := usedForRaw.([]any); ok {
			values := make([]string, 0, len(usedForSlice))
			for _, v := range usedForSlice {
				if s, ok := v.(string); ok {
					values = append(values, s)
				}
			}
			if len(values) > 0 {
				valueList, d := types.ListValueFrom(ctx, types.StringType, values)
				diags.Append(d...)
				model.UsedFor = valueList
			} else {
				model.UsedFor = types.ListNull(types.StringType)
			}
		} else {
			model.UsedFor = types.ListNull(types.StringType)
		}
	} else {
		model.UsedFor = types.ListNull(types.StringType)
	}

	if color, ok := GetString(result, "color"); ok && color != "" {
		model.Color = types.StringValue(color)
	} else {
		model.Color = types.StringNull()
	}
}
