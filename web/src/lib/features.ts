import { type LucideIcon, Shield, History, RotateCw, HardDrive, FileCheck, Wrench } from "lucide-react"

export interface Feature {
  title: string
  description: string
  icon: LucideIcon
  category: 'reliability' | 'performance' | 'usability' | 'scalability'
}

export const features: Feature[] = [
  {
    title: "Write-Ahead Logging",
    description: "Durable operations with configurable fsync and group commit for data safety.",
    icon: Shield,
    category: "reliability"
  },
  {
    title: "Crash Recovery",
    description: "Automatic recovery via WAL replay and checkpointing on startup.",
    icon: History,
    category: "reliability"
  },
  {
    title: "Data Integrity",
    description: "Per-file checksums with validation and repair capabilities.",
    icon: FileCheck,
    category: "reliability"
  },
  {
    title: "Admin Tools",
    description: "Built-in compaction and integrity verification commands.",
    icon: Wrench,
    category: "usability"
  },
  {
    title: "Smart Checkpointing",
    description: "Periodic snapshots with WAL truncation for optimal recovery.",
    icon: RotateCw,
    category: "performance"
  },
  {
    title: "Efficient Storage",
    description: "Human-readable JSON with atomic WAL-based updates.",
    icon: HardDrive,
    category: "performance"
  }
]