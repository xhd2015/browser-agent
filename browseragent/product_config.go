package browseragent

// ProductConfig parameterizes ports, names, and features for browser-agent
// vs browser-trace product shells (Go mirror of react/src/products/*).
type ProductConfig struct {
	ID               string   // "browser-agent" | "browser-trace"
	DisplayName      string
	CLIName          string
	ControlPort      int      // 43761 | 43759
	Features         []string
	PageMarkerGlobal string
	ExtensionDirName string
}

// ProductBrowserAgent is the browser-agent product constants (port 43761).
var ProductBrowserAgent = ProductConfig{
	ID:               "browser-agent",
	DisplayName:      "Browser Agent",
	CLIName:          "browser-agent",
	ControlPort:      43761,
	Features:         []string{"browser-agent"},
	PageMarkerGlobal: "__BROWSER_AGENT_EXT__",
	ExtensionDirName: "Chrome-Ext-Browser-Agent",
}

// ProductBrowserTrace is the browser-trace product constants (port 43759)
// exported for dual-product design / shared React parameterization.
var ProductBrowserTrace = ProductConfig{
	ID:               "browser-trace",
	DisplayName:      "Browser Trace",
	CLIName:          "browser-trace",
	ControlPort:      43759,
	Features:         []string{"browser-trace"},
	PageMarkerGlobal: "__BROWSER_TRACE_EXT__",
	ExtensionDirName: "Chrome-Ext-Capture-API",
}
