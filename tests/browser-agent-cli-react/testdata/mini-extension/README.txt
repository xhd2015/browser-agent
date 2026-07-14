Mini MV3 fixture shape for browser-agent embedded extension.

Production: stage Chrome-Ext-Browser-Agent build into
browseragent/embedded/extension/** for //go:embed.

CI must not run webpack/npm to execute this doctest tree. Ship a minimal
manifest+background (this folder) under the embed path.

Required: manifest.json "version", host_permissions mentioning 43761.
