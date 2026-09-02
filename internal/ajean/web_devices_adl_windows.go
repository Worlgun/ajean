//go:build windows

// web_devices_adl_windows.go — température GPU AMD réelle sous Windows, via
// l'ADL (AMD Display Library, atiadlxx.dll) chargée directement en mémoire.
//
// Pourquoi : aucun compteur de performance Windows (WMI/perfmon) n'expose la
// température d'un GPU — c'est une donnée que même le Gestionnaire des tâches
// ne lit que depuis une API DirectX privée ajoutée pour lui (Windows 11
// 22H2+), hors de portée d'un simple appel PowerShell (voir
// windowsAdapterStats dans web_devices.go, qui gère VRAM/utilisation par ce
// biais mais pas la température). L'ADL, elle, est l'API officielle qu'AMD
// fournit pour ça — les mêmes fonctions qu'utilisent GPU-Z, HWiNFO, etc.
// Layout des structures vérifié contre le header public d'AMD
// (github.com/GPUOpen-LibrariesAndSDKs/display-library, adl_structures.h /
// adl_defines.h), pas deviné : les fonctions ADL/ADL2 utilisées ici plantent
// silencieusement (ou renvoient n'importe quoi) si la moindre structure a la
// mauvaise taille, donc pas de place pour l'approximation.
//
// Best-effort total : n'importe quelle étape qui échoue (DLL absente, init
// ADL en erreur, capteur non supporté) rend juste "pas de température", jamais
// une erreur qui remonterait à l'appelant.
package ajean

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	adlKernel32       = syscall.NewLazyDLL("kernel32.dll")
	adlProcLocalAlloc = adlKernel32.NewProc("LocalAlloc")

	adlDLL                 = syscall.NewLazyDLL("atiadlxx.dll")
	adlProcMainCreate      = adlDLL.NewProc("ADL_Main_Control_Create")
	adlProcMainDestroy     = adlDLL.NewProc("ADL_Main_Control_Destroy")
	adlProcNumAdapters     = adlDLL.NewProc("ADL_Adapter_NumberOfAdapters_Get")
	adlProcAdapterInfo     = adlDLL.NewProc("ADL_Adapter_AdapterInfo_Get")
	adlProcMain2Create     = adlDLL.NewProc("ADL2_Main_Control_Create")
	adlProcMain2Destroy    = adlDLL.NewProc("ADL2_Main_Control_Destroy")
	adlProcPMLogStart      = adlDLL.NewProc("ADL2_Adapter_PMLog_Start")
	adlProcQueryPMLogData  = adlDLL.NewProc("ADL2_New_QueryPMLogData_Get")
)

const (
	adlVendorAMD      = 1002 // PCI vendor ID
	adlPMLogMaxSensor = 256
	adlSensorHotspot  = 27 // ADL_PMLOG_TEMPERATURE_HOTSPOT
	adlSensorEdge     = 8  // ADL_PMLOG_TEMPERATURE_EDGE
	adlSensorBoardPow = 73 // ADL_PMLOG_BOARD_POWER — newer driver revisions only
	adlSensorAsicPow  = 23 // ADL_PMLOG_ASIC_POWER — long-standing fallback, widely supported
	adlSensorGfxPow   = 30 // ADL_PMLOG_GFX_POWER — narrower (graphics core only), last resort
	adlSensorFanRPM   = 14 // ADL_PMLOG_FAN_RPM
	adlSensorFanPct   = 15 // ADL_PMLOG_FAN_PERCENTAGE
)

func adlMallocCallback(size uintptr) uintptr {
	ret, _, _ := adlProcLocalAlloc.Call(0 /* LMEM_FIXED */, size)
	return ret
}

// adlMallocCallbackPtr is the ONE native function pointer ADL's malloc
// callback ever uses, for the lifetime of the process. syscall.NewCallback
// hands out a fixed-size, never-freed slot each time it's called (~2000 per
// process on this Go runtime) — calling it fresh inside adlAMDHotspotTempC on
// every invocation exhausted that table after enough polls of /api/vram (the
// UI hits it periodically) and crashed the whole ajean-ui process with an
// access violation deep inside the DLL call. sync.OnceValue makes the actual
// syscall.NewCallback call happen exactly once, no matter how many times
// adlAMDHotspotTempC runs.
var adlMallocCallbackPtr = sync.OnceValue(func() uintptr {
	return syscall.NewCallback(adlMallocCallback)
})

// adlAdapterInfo mirrors ADL's AdapterInfo struct (adl_structures.h) exactly:
// every field is either a 4-byte int or a fixed byte array, so there's no
// hidden padding to get wrong on a 64-bit build.
type adlAdapterInfo struct {
	Size           int32
	AdapterIndex   int32
	UDID           [256]byte
	BusNumber      int32
	DeviceNumber   int32
	FunctionNumber int32
	VendorID       int32
	AdapterName    [256]byte
	DisplayName    [256]byte
	Present        int32
	Exist          int32
	DriverPath     [256]byte
	DriverPathExt  [256]byte
	PNPString      [256]byte
	OSDisplayIndex int32
}

// adlPMLogStartInput mirrors ADLPMLogStartInput.
type adlPMLogStartInput struct {
	Sensors    [adlPMLogMaxSensor]uint16
	SampleRate uint32
	Reserved   [15]int32
}

// adlPMLogStartOutput mirrors ADLPMLogStartOutput (a pointer-sized union
// followed by reserved ints — the union collapses to 8 bytes on x64).
type adlPMLogStartOutput struct {
	LoggingAddress uintptr
	Reserved       [14]int32
}

type adlSingleSensorData struct {
	Supported int32
	Value     int32
}

// adlPMLogDataOutput mirrors ADLPMLogDataOutput: a leading size field, then
// an array indexed DIRECTLY by ADL_PMLOG_SENSORS id (not a sparse id/value
// list) — sensors[adlSensorHotspot] is the hotspot reading, no scanning needed.
type adlPMLogDataOutput struct {
	Size    int32
	Sensors [adlPMLogMaxSensor]adlSingleSensorData
}

// adlCallMu serializes EVERY call into atiadlxx.dll. Required, not defensive
// paranoia: a concurrent-load test (50-80 requests in parallel against
// /api/vram, which is what a UI that polls this endpoint actually produces)
// crashed the whole ajean-ui process with an access violation inside the DLL
// — ajean's HTTP server runs each request on its own goroutine, so two
// requests landing on adlAMDHotspotTempC at the same instant is the normal
// case, not a rare edge case. AMD's ADL predates Go/goroutines by a decade
// and gives no documented thread-safety guarantee; holding this lock for the
// DLL's full create→query→destroy sequence is the only way to guarantee it's
// never entered from two goroutines at once. The calls inside are fast
// (single-digit ms), so serializing them isn't a meaningful bottleneck even
// under the same load that used to crash it.
//
// This mutex alone isn't enough, though: a THIRD incident (an "Application
// Hang" Windows itself force-closed after ~65 minutes of no response) hit the
// day this fix shipped. adlAMDSensorsBlocking below has no timeout on any of
// its ADL calls — if the DLL itself ever stalls (a driver hiccup, the GPU
// busy with something else, anything), the goroutine holding adlCallMu never
// returns, so it never unlocks, so every subsequent /api/vram poll (one
// every 3s from the browser) piles up forever waiting on a lock that will
// never be released. One bad DLL call and the app is wedged for good — worse
// than the original crash, since a crash at least gets Windows/the user to
// notice and restart it.
//
// The fix isn't a timeout on the mutex (Go can't force a stuck syscall to
// return early — the goroutine and the lock would still be stuck forever
// either way). It's removing the DLL call from the request path entirely:
// adlPoller below is the ONLY goroutine that ever calls
// adlAMDSensorsBlocking, on its own loop, independent of how many browser
// tabs are polling /api/vram or how often. adlAMDSensors (what callers
// actually use) just reads whatever the poller last cached — instant, never
// blocks, and immune to the DLL by construction. If the DLL does stall, the
// poller goroutine is the only thing that gets stuck; the cache goes stale
// (and adlAMDSensors reports it as unavailable past adlCacheMaxAge) instead
// of ever taking the HTTP server down with it.
var adlCallMu sync.Mutex

// adlSensors is every reading adlAMDSensorsBlocking was able to obtain in one
// PMLog query — each field's companion bool is false when that specific
// sensor isn't supported on the card (common for e.g. board power on older
// ADL revisions), independently of whether any of the others succeeded.
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

// adlCacheMaxAge bounds how long a cached reading is trusted once the
// poller stops updating it (DLL stuck, adapter unplugged mid-session,
// whatever) — past this, adlAMDSensors reports "unavailable" instead of a
// silently frozen, increasingly wrong number.
const adlCacheMaxAge = 15 * time.Second

// adlPollInterval matches the UI's own /api/vram poll rate (09-stream.js:
// setInterval(loadVram, 3000)) — no point refreshing the cache faster than
// anything will ever read it.
const adlPollInterval = 3 * time.Second

var (
	adlCacheMu  sync.RWMutex
	adlCacheVal adlSensors
	adlCacheAt  time.Time

	adlPollerOnce sync.Once
)

// adlAMDHotspotTempC returns just the temperature from the cache, kept as a
// thin wrapper since it was the original entry point and other callers may
// still reasonably want only this one value.
func adlAMDHotspotTempC() (int, bool) {
	s := adlAMDSensors()
	return s.TempC, s.HasTemp
}

// adlAMDSensors returns the poller's last cached reading — never touches the
// DLL itself, so it can be called from an HTTP handler on every poll without
// any risk of blocking on it. Starts the background poller on first call
// (lazy: a machine with no AMD GPU, or with the web UI never opened, never
// spins it up at all).
func adlAMDSensors() adlSensors {
	adlPollerOnce.Do(func() { go adlPoller() })
	adlCacheMu.RLock()
	defer adlCacheMu.RUnlock()
	if time.Since(adlCacheAt) > adlCacheMaxAge {
		return adlSensors{}
	}
	return adlCacheVal
}

// adlPoller owns the only goroutine allowed to call adlAMDSensorsBlocking.
// Runs for the lifetime of the process once started — deliberately simple
// over demand-aware (stop when no one's asked in a while, restart on
// demand): the DLL calls are cheap (single-digit ms) and this fully avoids
// the far worse failure mode of racing a stop/restart against a stuck call.
func adlPoller() {
	for {
		v := adlAMDSensorsBlocking()
		adlCacheMu.Lock()
		adlCacheVal, adlCacheAt = v, time.Now()
		adlCacheMu.Unlock()
		time.Sleep(adlPollInterval)
	}
}

// adlAMDSensorsBlocking does the actual ADL work — temperature, power draw,
// and fan speed of the first present AMD adapter ADL reports, in a single
// PMLog query (cheaper than querying each sensor separately, and keeps them
// from drifting apart from momentary readings taken microseconds apart).
// Called ONLY from adlPoller (see the adlCallMu doc comment above for why
// nothing else may call this directly). On a machine with more than
// one physical AMD GPU this only ever reports the first one — ADL enumerates
// one entry per display OUTPUT, not per physical card, so reliably telling
// two real cards apart would need deduplicating by bus/device/function, not
// attempted here since it's not this machine's situation.
func adlAMDSensorsBlocking() adlSensors {
	adlCallMu.Lock()
	defer adlCallMu.Unlock()

	if err := adlDLL.Load(); err != nil {
		return adlSensors{}
	}

	if r, _, _ := adlProcMainCreate.Call(adlMallocCallbackPtr(), 1); int32(r) != 0 {
		return adlSensors{}
	}
	defer adlProcMainDestroy.Call()

	var numAdapters int32
	adlProcNumAdapters.Call(uintptr(unsafe.Pointer(&numAdapters)))
	if numAdapters <= 0 {
		return adlSensors{}
	}
	infos := make([]adlAdapterInfo, numAdapters)
	adlProcAdapterInfo.Call(uintptr(unsafe.Pointer(&infos[0])), uintptr(int(unsafe.Sizeof(adlAdapterInfo{}))*int(numAdapters)))

	var amdIdx int32 = -1
	for i := range infos {
		if infos[i].Exist != 0 && infos[i].VendorID == adlVendorAMD {
			amdIdx = infos[i].AdapterIndex
			break
		}
	}
	if amdIdx < 0 {
		return adlSensors{}
	}

	var ctx uintptr
	if r, _, _ := adlProcMain2Create.Call(adlMallocCallbackPtr(), 1, uintptr(unsafe.Pointer(&ctx))); int32(r) != 0 || ctx == 0 {
		return adlSensors{}
	}
	defer adlProcMain2Destroy.Call(ctx)

	// Ask for temp + power + fan sensors in one go; 0 (ADL_SENSOR_MAXTYPES)
	// terminates the requested list. Errors here are ignored on purpose: on
	// this driver, a repeat Start (logging already active from a previous
	// call) errors out while the query still returns live data just fine.
	var start adlPMLogStartInput
	start.Sensors[0] = adlSensorHotspot
	start.Sensors[1] = adlSensorEdge
	start.Sensors[2] = adlSensorBoardPow
	start.Sensors[3] = adlSensorAsicPow
	start.Sensors[4] = adlSensorGfxPow
	start.Sensors[5] = adlSensorFanRPM
	start.Sensors[6] = adlSensorFanPct
	start.SampleRate = 1000
	var startOut adlPMLogStartOutput
	adlProcPMLogStart.Call(ctx, uintptr(amdIdx), uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&startOut)))

	var data adlPMLogDataOutput
	data.Size = int32(unsafe.Sizeof(data))
	if r, _, _ := adlProcQueryPMLogData.Call(ctx, uintptr(amdIdx), uintptr(unsafe.Pointer(&data))); int32(r) != 0 {
		return adlSensors{}
	}

	var out adlSensors
	if hot := data.Sensors[adlSensorHotspot]; hot.Supported != 0 {
		out.TempC, out.HasTemp = int(hot.Value), true
	} else if edge := data.Sensors[adlSensorEdge]; edge.Supported != 0 {
		out.TempC, out.HasTemp = int(edge.Value), true
	}
	// Power: prefer the most complete reading available. BOARD_POWER (total
	// draw at the connectors) is only on newer driver revisions; ASIC_POWER
	// (whole chip) is the long-standing, widely-supported fallback; GFX_POWER
	// (graphics core only, excludes memory/VRMs) is the last resort.
	if p := data.Sensors[adlSensorBoardPow]; p.Supported != 0 {
		out.PowerW, out.HasPow = int(p.Value), true
	} else if p := data.Sensors[adlSensorAsicPow]; p.Supported != 0 {
		out.PowerW, out.HasPow = int(p.Value), true
	} else if p := data.Sensors[adlSensorGfxPow]; p.Supported != 0 {
		out.PowerW, out.HasPow = int(p.Value), true
	}
	if rpm := data.Sensors[adlSensorFanRPM]; rpm.Supported != 0 {
		out.FanRPM, out.HasRPM = int(rpm.Value), true
	}
	if pct := data.Sensors[adlSensorFanPct]; pct.Supported != 0 {
		out.FanPct, out.HasPct = int(pct.Value), true
	}
	return out
}
