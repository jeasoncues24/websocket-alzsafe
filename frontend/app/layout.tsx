import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { AdminAuthCheck } from "@/components/admin-auth-check";

const fontSans = localFont({
  src: "../public/fonts/PlusJakartaSans-Variable.woff2",
  variable: "--font-sans",
  display: "swap",
  weight: "200 800",
});



export const metadata: Metadata = {
  title: "WhatsApp API Admin",
  description: "Panel administrativo para gestión de WhatsApp API",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="es" suppressHydrationWarning>
      <body className={`${fontSans.variable} antialiased`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <AdminAuthCheck>{children}</AdminAuthCheck>
        </ThemeProvider>
      </body>
    </html>
  );
}
