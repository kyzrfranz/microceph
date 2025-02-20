package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"

	"github.com/canonical/microceph/microceph/api/types"
	tests2 "github.com/canonical/microceph/microceph/mocks"
	"github.com/lxc/lxd/lxc/utils"
	"github.com/lxc/lxd/shared/api"
)

// make sure it fails like before on empty config
func TestCmdDiskListExecute(t *testing.T) {
	tests := []struct {
		name    string
		common  *CmdControl
		wantErr string
	}{
		{
			"failure constructing app without state dir",
			&CmdControl{
				FlagStateDir: "",
			},
			"Missing state directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cmdDiskList{
				common: tt.common,
				disk:   nil,
			}
			if err := c.Command().ExecuteContext(context.Background()); (err != nil) && err.Error() != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCmdDiskListRun(t *testing.T) {

	type disks struct {
		mustBeRendered bool
		data           types.Disks
		error          error
	}

	type resources struct {
		mustBeRendered bool
		data           *api.ResourcesStorage
		error          error
	}

	tests := []struct {
		name               string
		showAvailableDisks bool
		formatString       string
		mockDisks          disks
		mockResources      resources
		wantErr            bool
	}{
		{
			name:               "Render empty list",
			showAvailableDisks: false,
			mockDisks: disks{
				mustBeRendered: true,
				data:           types.Disks{},
				error:          nil,
			},
			mockResources: resources{
				mustBeRendered: true,
				data:           &api.ResourcesStorage{},
				error:          nil,
			},
			wantErr: false,
		},
		{
			name:               "Render 2 configured disks",
			showAvailableDisks: false,
			mockDisks: disks{
				mustBeRendered: true,
				data: []types.Disk{
					{
						OSD:      0,
						Path:     "/tmp/folder-1",
						Location: "node-1",
					},
					{
						OSD:      1,
						Path:     "/tmp/folder-2",
						Location: "node-2",
					},
				},
				error: nil,
			},
			mockResources: resources{
				mustBeRendered: true,
				data:           &api.ResourcesStorage{},
				error:          nil,
			},
			wantErr: false,
		},
		{
			name:               "Don't render 2 available disks without flag",
			showAvailableDisks: false,
			mockDisks: disks{
				mustBeRendered: true,
				data:           types.Disks{},
				error:          nil,
			},
			mockResources: resources{
				mustBeRendered: false,
				data: &api.ResourcesStorage{
					Disks: []api.ResourcesStorageDisk{
						{
							Model:    "Virtual Warp Drive",
							Size:     1000,
							DeviceID: uuid.NewUUID().String(),
							Type:     "warp",
						},
						{
							Model:    "Virtual Flux Drive",
							Size:     1000,
							DeviceID: uuid.NewUUID().String(),
							Type:     "flux",
						},
					},
					Total: 0,
				},
				error: nil,
			},
			wantErr: false,
		},
		{
			name:               "Render 2 available disks with --available flag",
			showAvailableDisks: true,
			mockDisks: disks{
				mustBeRendered: true,
				data:           types.Disks{},
				error:          nil,
			},
			mockResources: resources{
				mustBeRendered: true,
				data: &api.ResourcesStorage{
					Disks: []api.ResourcesStorageDisk{
						{
							Model:    "Virtual Warp Drive",
							Size:     1000,
							DeviceID: uuid.NewUUID().String(),
							Type:     "warp",
						},
						{
							Model:    "Virtual Flux Drive",
							Size:     1000,
							DeviceID: uuid.NewUUID().String(),
							Type:     "flux",
						},
					},
					Total: 0,
				},
				error: nil,
			},
			wantErr: false,
		},
		{
			name:               "Only render configured discs by default",
			showAvailableDisks: false,
			mockDisks: disks{
				mustBeRendered: true,
				data: []types.Disk{
					{
						OSD:      0,
						Path:     "/tmp/folder-1",
						Location: "node-1",
					},
					{
						OSD:      1,
						Path:     "/tmp/folder-2",
						Location: "node-2",
					},
					{
						OSD:      2,
						Path:     "/tmp/folder-3",
						Location: "node-3",
					},
				},
				error: nil,
			},
			mockResources: resources{
				mustBeRendered: false,
				data: &api.ResourcesStorage{
					Disks: []api.ResourcesStorageDisk{
						{
							Model:    "Virtual Warp Drive",
							Size:     1000,
							DeviceID: uuid.NewUUID().String(),
							Type:     "warp",
						},
						{
							Model:    "Virtual Flux Drive",
							Size:     1000,
							DeviceID: uuid.NewUUID().String(),
							Type:     "flux",
						},
					},
					Total: 0,
				},
				error: nil,
			},
			wantErr: false,
		},
		{
			name:               "JSON format configured disks",
			showAvailableDisks: false,
			formatString:       "json",
			mockDisks: disks{
				mustBeRendered: true,
				data: []types.Disk{
					{
						OSD:      0,
						Path:     "/tmp/folder-1",
						Location: "node-1",
					},
					{
						OSD:      1,
						Path:     "/tmp/folder-2",
						Location: "node-2",
					},
					{
						OSD:      2,
						Path:     "/tmp/folder-3",
						Location: "node-3",
					},
				},
				error: nil,
			},
			mockResources: resources{
				data: &api.ResourcesStorage{},
			},
			wantErr: false,
		},
		{
			name:               "YAML format configured disks",
			showAvailableDisks: false,
			formatString:       "yaml",
			mockDisks: disks{
				mustBeRendered: true,
				data: []types.Disk{
					{
						OSD:      0,
						Path:     "/tmp/folder-1",
						Location: "node-1",
					},
					{
						OSD:      1,
						Path:     "/tmp/folder-2",
						Location: "node-2",
					},
					{
						OSD:      2,
						Path:     "/tmp/folder-3",
						Location: "node-3",
					},
				},
				error: nil,
			},
			mockResources: resources{
				data: &api.ResourcesStorage{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiMock := tests2.ApiMock{}
			apiMock.On("GetDisks", mock.Anything).Return(tt.mockDisks.data, tt.mockDisks.error)
			apiMock.On("GetResources", mock.Anything).Return(tt.mockResources.data, tt.mockResources.error)
			c := &cmdDiskList{
				apiClient: &apiMock,
			}

			cmd := c.Command()
			c.flgShowAvailableDisks = tt.showAvailableDisks
			c.flagFormat = tt.formatString
			if c.flagFormat == "" {
				c.flagFormat = utils.TableFormatTable
			}

			saveStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			if err := c.Run(cmd, []string{}); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = saveStdout

			outputString := string(out)

			//validate output for table formats
			validateOutput(t, outputString, tt.formatString, tt.mockDisks.data, tt.mockResources.data, tt.mockDisks.mustBeRendered, tt.mockResources.mustBeRendered)

			println(outputString)

		})
	}
}

func validateOutput(t *testing.T, outputString string, formatString string, configuredDisks []types.Disk, availableDisks *api.ResourcesStorage, shouldRenderConfiguredDisks bool, shouldRenderAvailableDisks bool) {
	if shouldRenderConfiguredDisks {
		if formatString == utils.TableFormatTable {
			for _, disk := range configuredDisks {
				assert.Contains(t, outputString, strconv.FormatInt(disk.OSD, 10))
				assert.Contains(t, outputString, disk.Location)
				assert.Contains(t, outputString, disk.Path)
			}
		} else if formatString == utils.TableFormatJSON {
			var disks []types.Disk
			json.Unmarshal([]byte(outputString), &disks)
			assert.Equal(t, configuredDisks, disks)
		} else if formatString == utils.TableFormatYAML {
			var disks []types.Disk
			yaml.Unmarshal([]byte(outputString), &disks)
			assert.Equal(t, configuredDisks, disks)
		}
	} else {
		if formatString == utils.TableFormatTable {
			for _, disk := range configuredDisks {
				assert.NotContains(t, outputString, strconv.FormatInt(disk.OSD, 10))
				assert.NotContains(t, outputString, disk.Location)
				assert.NotContains(t, outputString, disk.Path)
			}
		}
	}

	if shouldRenderAvailableDisks {
		if formatString == utils.TableFormatTable {
			for _, disk := range availableDisks.Disks {
				assert.Contains(t, outputString, disk.Model)
				assert.Contains(t, outputString, fmt.Sprintf("%dB", disk.Size))
				assert.Contains(t, outputString, disk.Type)
				assert.Contains(t, outputString, fmt.Sprintf("/dev/disk/by-id/%s", disk.DeviceID))
			}
		} else if formatString == utils.TableFormatJSON {
			var disks []api.ResourcesStorageDisk
			json.Unmarshal([]byte(outputString), &disks)
			assert.Equal(t, availableDisks, disks)
		} else if formatString == utils.TableFormatYAML {
			var disks []api.ResourcesStorageDisk
			yaml.Unmarshal([]byte(outputString), &disks)
			assert.Equal(t, availableDisks, disks)
		}
	} else {
		if formatString == utils.TableFormatTable {
			for _, disk := range availableDisks.Disks {
				assert.NotContains(t, outputString, disk.Model)
				assert.NotContains(t, outputString, fmt.Sprintf("%dB", disk.Size))
				assert.NotContains(t, outputString, disk.Type)
				assert.NotContains(t, outputString, fmt.Sprintf("/dev/disk/by-id/%s", disk.DeviceID))
			}
		}
	}
}
