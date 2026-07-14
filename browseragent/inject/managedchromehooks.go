package inject

// ManagedChromeHooks records injectable managed-chrome hooks for CLI doctests.
type ManagedChromeHooks struct {
	LaunchFn func(args []string) error
}

// ManagedChromeTestHooks is the doctest assignment target for open-managed-chrome LaunchFn.
var ManagedChromeTestHooks *ManagedChromeHooks