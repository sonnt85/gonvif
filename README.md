# gonvif

[![Go Reference](https://pkg.go.dev/badge/github.com/sonnt85/gonvif.svg)](https://pkg.go.dev/github.com/sonnt85/gonvif)

ONVIF client library for Go — discover and control IP cameras via the ONVIF protocol.

## Installation

```bash
go get github.com/sonnt85/gonvif
```

## Features

- WS-Discovery to find ONVIF devices on a network interface
- Connect to ONVIF devices by address with optional HTTP client
- Automatic capability discovery (Media, PTZ, Imaging, Events, Analytics)
- WS-Security (UsernameToken) authentication
- Call arbitrary ONVIF SOAP methods via `CallMethod`
- Structured types for Device, Media, PTZ, Imaging, Event, and Analytics services
- ISO 8601 duration support

## Usage

```go
// Discover NVT devices on an interface
devices := gonvif.GetAvailableDevicesAtSpecificEthernetInterface("eth0")

// Connect to a known camera
dev, err := gonvif.NewDevice(gonvif.DeviceParams{
    Xaddr:    "192.168.1.100",
    Username: "admin",
    Password: "secret",
})
if err != nil {
    log.Fatal(err)
}

// List discovered service endpoints
fmt.Println(dev.GetServices())

// Call a ONVIF method (e.g., GetProfiles from media package)
resp, err := dev.CallMethod(media.GetProfiles{})
```

## API

### Types

- `Device` — represents an ONVIF device; holds endpoints and device info
- `DeviceParams` — connection parameters (address, credentials, HTTP client)
- `DeviceType` — NVD / NVS / NVA / NVT constants
- `DeviceInfo` — device information returned by GetDeviceInformation

### Functions

- `NewDevice(params DeviceParams) (*Device, error)` — connect to a device and discover its capabilities
- `GetAvailableDevicesAtSpecificEthernetInterface(iface string) []Device` — WS-Discovery probe on a network interface

### Device Methods

- `CallMethod(method interface{}) (*http.Response, error)` — call any ONVIF SOAP method
- `GetServices() map[string]string` — return discovered service endpoint URLs
- `GetDeviceInfo() DeviceInfo` — return device information
- `GetEndpoint(name string) string` — return endpoint URL for a named service

### Sub-packages

- `device` — Device service types and operations
- `media` — Media service types
- `ptz` — PTZ service types
- `event` — Event service types (Subscribe, PullMessages, etc.)
- `Imaging` — Imaging service types
- `analytics` — Analytics service types
- `wsdiscovery` — WS-Discovery probe implementation
- `gosoap` — SOAP message builder with WS-Security support
- `onvifutils` — ONVIF utility helpers
- `xsd` — XSD built-in and ONVIF schema types

## Author

**sonnt85** — [thanhson.rf@gmail.com](mailto:thanhson.rf@gmail.com)

## License

MIT License - see [LICENSE](LICENSE) for details.
