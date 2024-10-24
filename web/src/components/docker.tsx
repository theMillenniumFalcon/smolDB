"use client"

import React, { useState, useCallback } from "react"
import copy from "copy-to-clipboard"
import { motion, AnimatePresence, MotionConfig } from "framer-motion"
import { toast } from "sonner"

const variants = {
  visible: { opacity: 1, scale: 1 },
  hidden: { opacity: 0, scale: 0.5 },
}

export function Docker() {
  const [copying, setCopying] = useState(0)

  const onCopy = useCallback(() => {
    copy(`docker run -p 8080:8080 nishank02/smoldb:latest`)
    toast.success('Text copied')
    setCopying((c) => c + 1)
    setTimeout(() => {
      setCopying((c) => c - 1)
    }, 2000)
  }, [])

  return (
    <div className="border border-border px-1 rounded mt-1 cursor-copy" onClick={onCopy}>
        <code className="block whitespace-pre-wrap break-all px-3 pr-[62px] rounded text-sm md:text-base relative h-auto items-center text-white">
                <div className="my-4 text-secondary-foreground">
                    {`docker run -p 8080:8080 nishank02/smoldb:latest`}
                </div>
            <button aria-label="Copy code" className="absolute right-1.5 top-1/2 -translate-y-1/2 cursor-pointer rounded-md bg-border text-white w-[26px] h-[26px] flex justify-center items-center">
                <MotionConfig transition={{ duration: 0.15 }}>
                    <AnimatePresence initial={false} mode="wait">
                        {copying ? (
                            <motion.div animate="visible" exit="hidden" initial="hidden" key="check" variants={variants}>
                                <svg
                                    viewBox="0 0 24 24"
                                    width="14"
                                    height="14"
                                    stroke="white"
                                    strokeWidth="1.5"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    fill="none"
                                    shapeRendering="geometricPrecision"
                                >
                                    <path d="M20 6L9 17l-5-5"></path>
                                </svg>
                            </motion.div>
                        ) : (
                            <motion.div animate="visible" exit="hidden" initial="hidden" key="copy" variants={variants}>
                                <svg
                                    viewBox="0 0 24 24"
                                    width="14"
                                    height="14"
                                    stroke="white"
                                    strokeWidth="1.5"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    fill="none"
                                    shapeRendering="geometricPrecision"
                                >
                                    <path d="M8 17.929H6c-1.105 0-2-.912-2-2.036V5.036C4 3.91 4.895 3 6 3h8c1.105 0 2 .911 2 2.036v1.866m-6 .17h8c1.105 0 2 .91 2 2.035v10.857C20 21.09 19.105 22 18 22h-8c-1.105 0-2-.911-2-2.036V9.107c0-1.124.895-2.036 2-2.036z"></path>
                                </svg>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </MotionConfig>
            </button>
        </code>
    </div>
  )
}