import { useState, useEffect } from "react";

const ThemeModes = {
  Light: "light",
  Dark: "dark",
  System: "system",
} as const;

type ThemeMode = (typeof ThemeModes)[keyof typeof ThemeModes];

const THEME_KEY = "theme-mode";

function getSystemTheme(): "light" | "dark" {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

function resolveTheme(mode: ThemeMode): "light" | "dark" {
  if (mode === ThemeModes.System) return getSystemTheme();
  return mode;
}

function applyTheme(resolved: "light" | "dark") {
  document.documentElement.setAttribute("data-theme", resolved);
}

export function useTheme() {
  const [mode, setMode] = useState<ThemeMode>(() => {
    try {
      return (localStorage.getItem(THEME_KEY) as ThemeMode) || ThemeModes.System;
    } catch {
      return ThemeModes.System;
    }
  });

  useEffect(() => {
    const resolved = resolveTheme(mode);
    applyTheme(resolved);
    try {
      localStorage.setItem(THEME_KEY, mode);
    } catch {
      // ignore
    }
  }, [mode]);

  useEffect(() => {
    if (mode !== ThemeModes.System) return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => applyTheme(getSystemTheme());
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, [mode]);

  return { mode, setMode, ThemeModes };
}
