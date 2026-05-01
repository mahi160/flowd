import { Moon, Sun } from "lucide-react";
import { useEffect, useState } from "preact/hooks";

export function ThemeToggle() {
  const [dark, setDark] = useState(() =>
    typeof window !== "undefined"
      ? window.matchMedia("(prefers-color-scheme: dark)").matches
      : false,
  );
  const toggleTheme = () => {
    setDark((prev) => !prev);
  };

  useEffect(() => {
    if (dark) {
      document.documentElement.dataset.theme = "dark";
    } else {
      document.documentElement.dataset.theme = "light";
    }
  }, [dark]);
  return (
    <button
      onClick={toggleTheme}
      className="size-8 flex items-center justify-center bg-surface rounded-lg shadow-sm cursor-pointer"
    >
      {dark ? <Moon size={16} /> : <Sun size={16} />}
    </button>
  );
}
