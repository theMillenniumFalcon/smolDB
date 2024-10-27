import "./globals.css"
import type { Metadata } from "next"
import Image from "next/image"
import { Inter } from "next/font/google"
import Link from "next/link"
import { Toaster } from "sonner"

import { ThemeProvider } from "@/lib/theme/themeProvider"
import { ModeToggle } from "@/components/theme/toggle"
import { SuccessIcon } from "@/components/icons/successIcon"

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
  title: "smolDB",
  description: "A JSON document-based database that relies on key based access to achieve O(1) access time."
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
              <Image
                src="/logo.svg"
                alt="logo"
                width="18"
                height="18"
              />
              <ModeToggle />
            </nav>

            <div className="z-10 w-full max-w-screen-md items-start justify-start flex flex-col">
              {children}
              <Toaster
                toastOptions={{
                  classNames: {
                    toast: 'bg-border',
                    title: 'text-secondary-foreground',
                  },
                  style: {
                    border: 'none',
                  },
                }}
                icons={{
                  success: <SuccessIcon />,
                }}
              />
            </div>
          </main>
        </ThemeProvider>
      </body>
    </html>
  )
}