"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { MessageSquareText } from "lucide-react";
import { cn } from "@/lib/utils";

const APP_VERSION = process.env.NEXT_PUBLIC_APP_VERSION ?? "2026.1.0";

interface Slide {
  title: string;
  highlight?: string;
  subtitle?: string;
  bigNumber?: string;
  bigNumberSuffix?: string;
  handwritten?: string;
  footer: string;
}

const slides: Slide[] = [
  {
    title: "Gestiona tus instancias de",
    highlight: "WhatsApp.",
    subtitle:
      "Conecta, monitorea y escala tus conversaciones empresariales desde un solo panel.",
    footer: "Menos fricción, más eficiencia.\nControla tu mensajería con wsapi.",
  },
  {
    title: "Mensajería en tiempo real,",
    highlight: "sin interrupciones.",
    bigNumber: "99.9%",
    bigNumberSuffix: "de uptime garantizado",
    handwritten: "WebSocket nativo",
    footer:
      "Con wsapi, cada mensaje llega cuando importa.\nEl control está en tus manos.",
  },
  {
    title: "API lista para integrar a",
    highlight: "cualquier sistema.",
    subtitle:
      "REST + WebSocket con autenticación JWT y roles granulares para tu operación.",
    footer: "De la startup al enterprise.\nHazlo posible con wsapi.",
  },
];

const countries = [
  { code: "PE", name: "Perú", colors: ["#D91023", "#FFFFFF", "#D91023"] },
  { code: "CO", name: "Colombia", colors: ["#FCD116", "#003893", "#CE1126"] },
  { code: "EC", name: "Ecuador", colors: ["#FFD100", "#0033A0", "#FF0000"] },
];

function ChatBubblePattern() {
  return (
    <svg
      className="absolute inset-0 w-full h-full opacity-[0.04]"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <defs>
        <pattern
          id="carousel-chat-pattern"
          x="0"
          y="0"
          width="160"
          height="130"
          patternUnits="userSpaceOnUse"
        >
          <rect x="8" y="10" width="90" height="32" rx="14" fill="white" />
          <path d="M 18 42 L 8 54 L 34 42 Z" fill="white" />
          <rect x="62" y="72" width="78" height="28" rx="13" fill="white" />
          <path d="M 125 100 L 140 108 L 128 100 Z" fill="white" />
          <rect x="8" y="88" width="44" height="22" rx="11" fill="white" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#carousel-chat-pattern)" />
    </svg>
  );
}

export function LoginCarousel() {
  const [current, setCurrent] = useState(0);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const resetInterval = useCallback(() => {
    if (intervalRef.current) clearInterval(intervalRef.current);
    intervalRef.current = setInterval(() => {
      setCurrent((c) => (c + 1) % slides.length);
    }, 5000);
  }, []);

  useEffect(() => {
    resetInterval();
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [resetInterval]);

  const goTo = useCallback(
    (index: number) => {
      setCurrent(index);
      resetInterval();
    },
    [resetInterval]
  );

  return (
    <div className="motion-fade-in relative flex h-full w-full flex-col justify-between overflow-hidden bg-zinc-950 p-8 lg:p-12">
      <ChatBubblePattern />

      {/* Header: logo + flags */}
      <div className="relative z-10 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex size-10 shrink-0 items-center justify-center rounded-xl border border-primary/30 bg-primary/20 shadow-sm">
            <MessageSquareText className="h-5 w-5 text-primary" />
          </div>
          <span className="text-white font-semibold text-lg tracking-tight">
            wsapi
          </span>
        </div>

        <div className="flex items-center gap-2">
          {countries.map((country) => (
            <div
              key={country.code}
              className="flex h-6 w-8 overflow-hidden rounded border border-white/20"
              aria-label={country.name}
              title={country.name}
            >
              {country.colors.map((color, i) => (
                <div
                  key={i}
                  className="h-full flex-1"
                  style={{ backgroundColor: color }}
                />
              ))}
            </div>
          ))}
        </div>
      </div>

      {/* Slides */}
      <div className="relative z-10 flex-1 flex flex-col justify-center overflow-hidden">
        <div
          className="flex motion-transform duration-[var(--motion-duration-slow)] ease-[var(--motion-ease-emphasized)]"
          style={{ transform: `translateX(-${current * 100}%)` }}
        >
          {slides.map((slide, index) => (
            <div key={index} className="flex w-full flex-none flex-col gap-4 pr-4">
              <h1 className="text-balance text-3xl font-bold leading-tight text-white lg:text-4xl xl:text-5xl">
                {slide.title}
                {slide.highlight && (
                  <span className="block text-primary">{slide.highlight}</span>
                )}
              </h1>

              {slide.bigNumber && (
                <div className="my-4">
                  <p className="text-5xl font-bold text-primary lg:text-6xl xl:text-7xl">
                    {slide.bigNumber}
                  </p>
                  {slide.bigNumberSuffix && (
                    <p className="mt-2 text-lg text-zinc-300">
                      {slide.bigNumberSuffix}
                    </p>
                  )}
                </div>
              )}

              {slide.handwritten && (
                <div className="inline-block rounded-lg bg-white/10 px-6 py-3">
                  <p className="font-serif text-2xl italic text-primary">
                    {slide.handwritten}
                  </p>
                </div>
              )}

              {slide.subtitle && (
                <>
                  <div className="h-1 w-16 rounded-full bg-primary" />
                  <p className="text-pretty text-lg text-zinc-400 lg:text-xl">
                    {slide.subtitle}
                  </p>
                </>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Footer */}
      <div className="relative z-10 flex flex-col gap-4">
        <p className="whitespace-pre-line text-sm text-zinc-400 lg:text-base">
          {slides[current].footer}
        </p>

        <div className="flex items-center gap-2" role="tablist">
          {slides.map((_, index) => (
            <button
              key={index}
              role="tab"
              aria-selected={index === current}
              aria-label={`Ir a slide ${index + 1}`}
              onClick={() => goTo(index)}
              className={cn(
                "motion-transition h-2 rounded-full",
                index === current
                  ? "w-8 bg-white"
                  : "w-2 bg-white/30 hover:bg-white/50"
              )}
            />
          ))}
        </div>

        <div className="flex items-center justify-between text-xs text-zinc-500">
          <span className="flex items-center gap-2">
            <span className="relative flex size-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75" />
              <span className="relative inline-flex size-2 rounded-full bg-primary" />
            </span>
            Sistema en línea
          </span>
          <span>v{APP_VERSION}</span>
        </div>
      </div>
    </div>
  );
}
