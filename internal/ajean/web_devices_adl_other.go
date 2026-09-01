//go:build !windows

package ajean

// adlSensors mirrors the Windows-only version in web_devices_adl_windows.go
// (kept in sync manually — Go's mutually-exclusive build tags mean only one
// of the two ever compiles for a given OS, so this isn't a real duplicate).
type adlSensors struct {
	TempC   int
	HasTemp bool
	PowerW  int
	HasPow  bool
	FanRPM  int
	HasRPM  bool
	FanPct  int
	HasPct  bool
}

// adlAMDHotspotTempC and adlAMDSensors are no-ops off Windows: ADL
// (atiadlxx.dll) doesn't exist there, and Linux/macOS already get real GPU
// temperature/power/fan through amd-smi / rocm-smi (see backend_gpu_amd.go,
// backend_gpu_rocm.go), so this fallback isn't needed outside Windows.
func adlAMDHotspotTempC() (int, bool) { return 0, false }
func adlAMDSensors() adlSensors       { return adlSensors{} }
