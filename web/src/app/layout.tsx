import "./globals.css"
import { ThemeProvider } from "@/lib/theme/themeProvider"
import type { Metadata } from "next"
import { Inter } from "next/font/google"
import Link from "next/link"
import { ModeToggle } from "@/components/theme/toggle"

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
  title: "smolDB",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <main className="flex min-h-screen w-screen items-center flex-col justify-start sm:px-8 px-4">
            <nav className="z-10 h-20 items-center justify-between w-full max-w-screen-md flex">
              <Link href="/">
                <div className="text-lg font-medium">
                  smolDB
                </div>
              </Link>
              <ModeToggle />
            </nav>

            <div className="z-10 w-full max-w-screen-md items-start justify-start flex flex-col">
              {children}
            </div>
          </main>
        </ThemeProvider>
      </body>
    </html>
  )
}