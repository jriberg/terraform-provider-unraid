package provider

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jriberg/terraform-provider-unraid/internal/client"
)

// ---- unraid_shares ----

type shareModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Comment    types.String `tfsdk:"comment"`
	FreeKB     types.Int64  `tfsdk:"free_kb"`
	UsedKB     types.Int64  `tfsdk:"used_kb"`
	SizeKB     types.Int64  `tfsdk:"size_kb"`
	Include    []string     `tfsdk:"include"`
	Exclude    []string     `tfsdk:"exclude"`
	Cache      types.Bool   `tfsdk:"cache"`
	Allocator  types.String `tfsdk:"allocator"`
	SplitLevel types.String `tfsdk:"split_level"`
	Floor      types.String `tfsdk:"floor"`
	Cow        types.String `tfsdk:"cow"`
	LuksStatus types.String `tfsdk:"luks_status"`
}

type sharesModel struct {
	Shares []shareModel `tfsdk:"shares"`
}

func newSharesDataSource() datasource.DataSource {
	return &readOnlyDataSource[sharesModel]{
		name: "shares",
		schema: schema.Schema{
			MarkdownDescription: "All user shares. Requires `READ_ANY` on `SHARE`. " +
				"Shares are read-only: the Unraid API has no mutations for creating or changing shares.",
			Attributes: map[string]schema.Attribute{
				"shares": cObjList("List of user shares.", map[string]schema.Attribute{
					"id":          cStr("Share identifier."),
					"name":        cStr("Share name."),
					"comment":     cStr("Share comment."),
					"free_kb":     cInt("Free space in KB on the disks/pool backing the share (not per share)."),
					"used_kb":     cInt("Used space in KB on the disks/pool backing the share (not per share)."),
					"size_kb":     cInt("Total size in KB."),
					"include":     cStrList("Disks included in the share."),
					"exclude":     cStrList("Disks excluded from the share."),
					"cache":       cBool("Whether the share uses a cache pool."),
					"allocator":   cStr("Allocation method, e.g. `highwater`."),
					"split_level": cStr("Split level."),
					"floor":       cStr("Minimum free space."),
					"cow":         cStr("Copy-on-write setting."),
					"luks_status": cStr("Encryption status."),
				}),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (sharesModel, error) {
			r, err := client.GetShares(ctx, c)
			if err != nil {
				return sharesModel{}, err
			}
			out := sharesModel{Shares: []shareModel{}}
			for _, s := range r.Shares {
				out.Shares = append(out.Shares, shareModel{
					ID:         types.StringValue(s.Id),
					Name:       types.StringPointerValue(s.Name),
					Comment:    types.StringPointerValue(s.Comment),
					FreeKB:     bigPtr(s.Free),
					UsedKB:     bigPtr(s.Used),
					SizeKB:     bigPtr(s.Size),
					Include:    s.Include,
					Exclude:    s.Exclude,
					Cache:      types.BoolPointerValue(s.Cache),
					Allocator:  types.StringPointerValue(s.Allocator),
					SplitLevel: types.StringPointerValue(s.SplitLevel),
					Floor:      types.StringPointerValue(s.Floor),
					Cow:        types.StringPointerValue(s.Cow),
					LuksStatus: types.StringPointerValue(s.LuksStatus),
				})
			}
			return out, nil
		},
	}
}

// ---- unraid_array ----

type arrayDiskModel struct {
	ID         types.String `tfsdk:"id"`
	Idx        types.Int64  `tfsdk:"idx"`
	Name       types.String `tfsdk:"name"`
	Device     types.String `tfsdk:"device"`
	SizeKB     types.Int64  `tfsdk:"size_kb"`
	Status     types.String `tfsdk:"status"`
	Type       types.String `tfsdk:"type"`
	Rotational types.Bool   `tfsdk:"rotational"`
	Temp       types.Int64  `tfsdk:"temp"`
	FsType     types.String `tfsdk:"fs_type"`
	FsSizeKB   types.Int64  `tfsdk:"fs_size_kb"`
	FsFreeKB   types.Int64  `tfsdk:"fs_free_kb"`
	FsUsedKB   types.Int64  `tfsdk:"fs_used_kb"`
	NumErrors  types.Int64  `tfsdk:"num_errors"`
	IsSpinning types.Bool   `tfsdk:"is_spinning"`
}

type arrayModel struct {
	ID          types.String     `tfsdk:"id"`
	State       types.String     `tfsdk:"state"`
	FreeKB      types.String     `tfsdk:"free_kb"`
	UsedKB      types.String     `tfsdk:"used_kb"`
	TotalKB     types.String     `tfsdk:"total_kb"`
	Parities    []arrayDiskModel `tfsdk:"parities"`
	Disks       []arrayDiskModel `tfsdk:"disks"`
	Caches      []arrayDiskModel `tfsdk:"caches"`
	BootDevices []arrayDiskModel `tfsdk:"boot_devices"`
}

var arrayDiskAttrs = map[string]schema.Attribute{
	"id":          cStr("Disk identifier."),
	"idx":         cInt("Slot number (parity 0/29, data 1-28, cache 30-53, flash 54)."),
	"name":        cStr("Slot name, e.g. `disk1`."),
	"device":      cStr("Device name, e.g. `sdb`."),
	"size_kb":     cInt("Disk size in KB."),
	"status":      cStr("Disk status, e.g. `DISK_OK`."),
	"type":        cStr("`DATA`, `PARITY`, `CACHE`, `BOOT` or `FLASH`."),
	"rotational":  cBool("`true` for HDD, `false` for SSD."),
	"temp":        cInt("Temperature in °C (null if unavailable)."),
	"fs_type":     cStr("File system type."),
	"fs_size_kb":  cInt("File system size in KB."),
	"fs_free_kb":  cInt("File system free space in KB."),
	"fs_used_kb":  cInt("File system used space in KB."),
	"num_errors":  cInt("Unrecoverable I/O errors."),
	"is_spinning": cBool("Whether the disk is spinning."),
}

func toArrayDisk(d client.ArrayDiskFields) arrayDiskModel {
	return arrayDiskModel{
		ID:         types.StringValue(d.Id),
		Idx:        types.Int64Value(int64(d.Idx)),
		Name:       types.StringPointerValue(d.Name),
		Device:     types.StringPointerValue(d.Device),
		SizeKB:     bigPtr(d.Size),
		Status:     strEnum(d.Status),
		Type:       types.StringValue(string(d.Type)),
		Rotational: types.BoolPointerValue(d.Rotational),
		Temp:       intPtr(d.Temp),
		FsType:     types.StringPointerValue(d.FsType),
		FsSizeKB:   bigPtr(d.FsSize),
		FsFreeKB:   bigPtr(d.FsFree),
		FsUsedKB:   bigPtr(d.FsUsed),
		NumErrors:  bigPtr(d.NumErrors),
		IsSpinning: types.BoolPointerValue(d.IsSpinning),
	}
}

func mapDisks(in []client.ArrayDiskFields) []arrayDiskModel {
	out := make([]arrayDiskModel, len(in))
	for i, d := range in {
		out[i] = toArrayDisk(d)
	}
	return out
}

func newArrayDataSource() datasource.DataSource {
	return &readOnlyDataSource[arrayModel]{
		name: "array",
		schema: schema.Schema{
			MarkdownDescription: "State, capacity and disks of the Unraid array. Requires `READ_ANY` on `ARRAY`.",
			Attributes: map[string]schema.Attribute{
				"id":           cStr("Array identifier."),
				"state":        cStr("Array state, e.g. `STARTED` or `STOPPED`."),
				"free_kb":      cStr("Free capacity in KB (string, as returned by the API)."),
				"used_kb":      cStr("Used capacity in KB."),
				"total_kb":     cStr("Total capacity in KB."),
				"parities":     cObjList("Parity disks.", arrayDiskAttrs),
				"disks":        cObjList("Data disks.", arrayDiskAttrs),
				"caches":       cObjList("Cache pool devices.", arrayDiskAttrs),
				"boot_devices": cObjList("Boot devices (flash or internal boot).", arrayDiskAttrs),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (arrayModel, error) {
			r, err := client.GetArray(ctx, c)
			if err != nil {
				return arrayModel{}, err
			}
			a := r.Array
			kb := a.Capacity.Kilobytes
			return arrayModel{
				ID:          types.StringValue(a.Id),
				State:       types.StringValue(string(a.State)),
				FreeKB:      types.StringValue(kb.Free),
				UsedKB:      types.StringValue(kb.Used),
				TotalKB:     types.StringValue(kb.Total),
				Parities:    mapDisks(a.Parities),
				Disks:       mapDisks(a.Disks),
				Caches:      mapDisks(a.Caches),
				BootDevices: mapDisks(a.BootDevices),
			}, nil
		},
	}
}

// ---- unraid_disks ----

type diskModel struct {
	ID               types.String  `tfsdk:"id"`
	Device           types.String  `tfsdk:"device"`
	Type             types.String  `tfsdk:"type"`
	Name             types.String  `tfsdk:"name"`
	Vendor           types.String  `tfsdk:"vendor"`
	SizeBytes        types.Float64 `tfsdk:"size_bytes"`
	SerialNum        types.String  `tfsdk:"serial_num"`
	FirmwareRevision types.String  `tfsdk:"firmware_revision"`
	InterfaceType    types.String  `tfsdk:"interface_type"`
	SmartStatus      types.String  `tfsdk:"smart_status"`
	Temperature      types.Float64 `tfsdk:"temperature"`
	IsSpinning       types.Bool    `tfsdk:"is_spinning"`
}

type disksModel struct {
	Disks []diskModel `tfsdk:"disks"`
}

func newDisksDataSource() datasource.DataSource {
	return &readOnlyDataSource[disksModel]{
		name: "disks",
		schema: schema.Schema{
			MarkdownDescription: "All physical disks detected by the server. Requires `READ_ANY` on `DISK`.",
			Attributes: map[string]schema.Attribute{
				"disks": cObjList("List of physical disks.", map[string]schema.Attribute{
					"id":                cStr("Disk identifier."),
					"device":            cStr("Device path, e.g. `/dev/sdb`."),
					"type":              cStr("Disk type, e.g. `SSD` or `HDD`."),
					"name":              cStr("Model name."),
					"vendor":            cStr("Manufacturer."),
					"size_bytes":        cFloat("Size in bytes."),
					"serial_num":        cStr("Serial number."),
					"firmware_revision": cStr("Firmware revision."),
					"interface_type":    cStr("`SATA`, `SAS`, `USB`, `PCIE` or `UNKNOWN`."),
					"smart_status":      cStr("`OK` or `UNKNOWN`."),
					"temperature":       cFloat("Temperature in °C (null if unavailable)."),
					"is_spinning":       cBool("Whether the disk is spinning."),
				}),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (disksModel, error) {
			r, err := client.GetDisks(ctx, c)
			if err != nil {
				return disksModel{}, err
			}
			out := disksModel{Disks: []diskModel{}}
			for _, d := range r.Disks {
				out.Disks = append(out.Disks, diskModel{
					ID:               types.StringValue(d.Id),
					Device:           types.StringValue(d.Device),
					Type:             types.StringValue(d.Type),
					Name:             types.StringValue(d.Name),
					Vendor:           types.StringValue(d.Vendor),
					SizeBytes:        types.Float64Value(d.Size),
					SerialNum:        types.StringValue(d.SerialNum),
					FirmwareRevision: types.StringValue(d.FirmwareRevision),
					InterfaceType:    types.StringValue(string(d.InterfaceType)),
					SmartStatus:      types.StringValue(string(d.SmartStatus)),
					Temperature:      types.Float64PointerValue(d.Temperature),
					IsSpinning:       types.BoolValue(d.IsSpinning),
				})
			}
			return out, nil
		},
	}
}
