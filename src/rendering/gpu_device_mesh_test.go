/******************************************************************************/
/* gpu_device_mesh_test.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "testing"

func TestGPUDeviceMeshReadRejectsInvalidMeshId(t *testing.T) {
	device := &GPUDevice{}
	if _, _, err := device.MeshRead(MeshId{}); err == nil {
		t.Fatalf("MeshRead should reject an invalid mesh id")
	}
}

func TestMeshIdCounts(t *testing.T) {
	id := MeshId{vertexCount: 3, indexCount: 6}
	if got := id.VertexCount(); got != 3 {
		t.Fatalf("VertexCount = %d, want 3", got)
	}
	if got := id.IndexCount(); got != 6 {
		t.Fatalf("IndexCount = %d, want 6", got)
	}
}
