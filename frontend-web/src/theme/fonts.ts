import { Inter } from "next/font/google";

/** Inter as the default app font. Mount the CSS variable on <html className={inter.variable}>. */
export const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});
