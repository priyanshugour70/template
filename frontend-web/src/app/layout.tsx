import type { Metadata, Viewport } from "next";

import { AuthProvider, PermissionsProvider, QueryProvider, ThemeProvider } from "@/providers";
import { inter } from "@/theme/fonts";
import { LIGHT_THEME, DARK_THEME } from "@/theme/palettes";

import "./globals.css";

export const metadata: Metadata = {
  title: "App",
  description: "Web frontend template",
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: LIGHT_THEME.background },
    { media: "(prefers-color-scheme: dark)", color: DARK_THEME.background },
  ],
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning className={inter.variable}>
      <body className={`${inter.className} min-h-screen bg-background text-foreground antialiased`}>
        <ThemeProvider>
          <QueryProvider>
            <AuthProvider>
              <PermissionsProvider>{children}</PermissionsProvider>
            </AuthProvider>
          </QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
