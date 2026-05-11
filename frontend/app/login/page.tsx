"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import {
  Eye,
  EyeOff,
  Loader2,
  MessageSquareText,
  Sun,
  Moon,
  ArrowRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { cn } from "@/lib/utils";
import { LoginCarousel } from "./carousel";

export default function LoginPage() {
  const router = useRouter();
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [fieldErrors, setFieldErrors] = useState({ username: "", password: "" });

  useEffect(() => {
    setMounted(true);
  }, []);

  function validateField(name: "username" | "password", value: string) {
    if (name === "username" && !value.trim()) {
      setFieldErrors((p) => ({ ...p, username: "El usuario es requerido" }));
    } else if (name === "password" && value.length > 0 && value.length < 4) {
      setFieldErrors((p) => ({ ...p, password: "Contraseña demasiado corta" }));
    } else {
      setFieldErrors((p) => ({ ...p, [name]: "" }));
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!username.trim()) {
      setFieldErrors((p) => ({ ...p, username: "El usuario es requerido" }));
      return;
    }

    setLoading(true);

    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error || "Credenciales inválidas");
        return;
      }

      localStorage.setItem("admin_token", data.token);
      router.push("/dashboard");
    } catch {
      setError("Error de conexión. Verifica tu red e intenta nuevamente.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-dvh flex">
      <aside className="hidden lg:flex lg:w-1/2 relative">
        <LoginCarousel />
      </aside>

      <div className="flex-1 flex flex-col bg-background">
        <header className="flex items-center justify-between px-5 py-4 lg:justify-end">
          <div className="flex items-center gap-2 lg:hidden">
            <div className="h-7 w-7 rounded-lg bg-primary/20 border border-primary/30 flex items-center justify-center">
              <MessageSquareText className="h-4 w-4 text-primary" />
            </div>
            <span className="font-semibold text-sm">wsapi</span>
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            aria-label="Cambiar tema"
          >
            {mounted && theme === "dark" ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </Button>
        </header>

        <main className="flex-1 flex items-center justify-center px-6 py-8">
          <div className="w-full max-w-sm space-y-8">
            <div className="space-y-1">
              <div className="flex items-center gap-2 mb-4 lg:hidden">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-primary" />
                </span>
                <span className="text-xs text-muted-foreground">
                  Sistema en línea
                </span>
              </div>
              <h2 className="text-2xl font-bold tracking-tight">
                Iniciar sesión
              </h2>
              <p className="text-sm text-muted-foreground">
                Panel administrativo · WhatsApp API
              </p>
            </div>

            <form
              onSubmit={handleSubmit}
              className="space-y-5"
              noValidate
              autoComplete="on"
            >
              {error && (
                <Alert variant="destructive" role="alert" aria-live="polite">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <div className="space-y-1.5">
                <Label htmlFor="username">Usuario</Label>
                <Input
                  id="username"
                  type="text"
                  placeholder="admin"
                  value={username}
                  onChange={(e) => {
                    setUsername(e.target.value);
                    if (fieldErrors.username) validateField("username", e.target.value);
                  }}
                  onBlur={() => validateField("username", username)}
                  autoComplete="username"
                  aria-invalid={!!fieldErrors.username}
                  aria-describedby={fieldErrors.username ? "username-error" : undefined}
                  className={cn(
                    fieldErrors.username &&
                      "border-destructive focus-visible:ring-destructive/30"
                  )}
                />
                {fieldErrors.username && (
                  <p id="username-error" className="text-xs text-destructive">
                    {fieldErrors.username}
                  </p>
                )}
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="password">Contraseña</Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      if (fieldErrors.password) validateField("password", e.target.value);
                    }}
                    onBlur={() => validateField("password", password)}
                    autoComplete="current-password"
                    aria-invalid={!!fieldErrors.password}
                    aria-describedby={fieldErrors.password ? "password-error" : undefined}
                    className={cn(
                      "pr-10",
                      fieldErrors.password &&
                        "border-destructive focus-visible:ring-destructive/30"
                    )}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((v) => !v)}
                    aria-label={showPassword ? "Ocultar contraseña" : "Mostrar contraseña"}
                    className="absolute right-0 top-0 h-full w-10 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </button>
                </div>
                {fieldErrors.password && (
                  <p id="password-error" className="text-xs text-destructive">
                    {fieldErrors.password}
                  </p>
                )}
              </div>

              <Button
                type="submit"
                className="w-full group"
                size="lg"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Verificando...
                  </>
                ) : (
                  <>
                    Acceder al panel
                    <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-0.5" />
                  </>
                )}
              </Button>
            </form>
          </div>
        </main>

        <footer className="pb-6 text-center text-xs text-muted-foreground">
          &copy; {new Date().getFullYear()} wsapi &middot; Acceso restringido a administradores
        </footer>
      </div>
    </div>
  );
}
