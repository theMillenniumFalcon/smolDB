import { type LucideIcon, Shield, History, RotateCw, HardDrive } from "lucide-react"

export interface Feature {
  title: string
  description: string
  icon: LucideIcon
  category: 'reliability' | 'performance' | 'usability' | 'scalability'
}

export const features: Feature[] = [
  {
    title: "Production Durability",
    description: "Write-Ahead Logging with configurable fsync modes and group commit.",
    icon: Shield,
    category: "reliability"
  },
  {
    title: "Crash Recovery",
    description: "Automatic recovery via checkpoints and WAL replay.",
    icon: History,
    category: "reliability"
  },
  {
    title: "Smart Checkpointing",
    description: "Periodic snapshots with WAL truncation for optimal performance.",
    icon: RotateCw,
    category: "performance"
  },
  {
    title: "Efficient Storage",
    description: "Atomic WAL truncation with human-readable JSON format.",
    icon: HardDrive,
    category: "performance"
  }
]