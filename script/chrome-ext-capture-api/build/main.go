package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/xgo/support/cmd"
)

const extensionDir = "Chrome-Ext-Capture-API"

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(_ []string) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	extRoot := filepath.Join(root, extensionDir)
	if _, err := os.Stat(extRoot); err != nil {
		return fmt.Errorf("%s not found; run this from the project-api-capture repo root", extensionDir)
	}

	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is not installed; install Node.js from https://nodejs.org/")
	}

	nodeModules := filepath.Join(extRoot, "node_modules")
	if _, err := os.Stat(nodeModules); err != nil {
		fmt.Println("Installing Chrome extension dependencies...")
		if err := cmd.Debug().Dir(extRoot).Run("npm", "install"); err != nil {
			return fmt.Errorf("npm install failed: %w", err)
		}
	}

	fmt.Println("Building Chrome extension...")
	if err := cmd.Debug().Dir(extRoot).Run("npm", "run", "build"); err != nil {
		return fmt.Errorf("npm run build failed: %w", err)
	}

	buildDir, err := filepath.Abs(filepath.Join(extRoot, "build"))
	if err != nil {
		return fmt.Errorf("resolve build directory: %w", err)
	}

	fmt.Println()
	fmt.Println("Chrome extension build complete.")
	fmt.Println()
	printLoadInstructions(buildDir)
	return nil
}

func printLoadInstructions(buildDir string) {
	fmt.Println("Load the extension in Chrome:")
	fmt.Println()
	fmt.Println("  1. Open chrome://extensions")
	fmt.Println("  2. Enable Developer mode (top-right toggle)")
	fmt.Println("  3. Click Load unpacked")
	fmt.Printf("  4. Select this folder:\n\n     %s\n\n", buildDir)
	fmt.Println("After code changes, rebuild with:")
	fmt.Println()
	fmt.Println("  go run ./script/chrome-ext-capture-api/build")
	fmt.Println()
	fmt.Println("Then click Reload on the extension card in chrome://extensions.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  - Open the tab you want to record")
	fmt.Println("  - Click the extension icon -> Start Recording")
	fmt.Println("  - Perform the API flow in that tab")
	fmt.Println("  - Click Stop & Download HAR")
	fmt.Println()
	fmt.Println("Note: Chrome shows a \"Debugger attached\" banner while recording.")
	fmt.Println("      Only one debugger can attach per tab at a time.")
}