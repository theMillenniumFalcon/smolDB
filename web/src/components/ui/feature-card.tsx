import { type Feature } from "@/lib/features"

interface FeatureCardProps {
  feature: Feature
}

export function FeatureCard({ feature }: FeatureCardProps) {
  const Icon = feature.icon
  
  return (
    <div className="bg-muted/50 rounded-lg p-3 border border-border">
      <div className="flex items-center gap-2 mb-1">
        <Icon className="w-4 h-4 text-primary" />
        <h3 className="text-sm font-medium">{feature.title}</h3>
      </div>
      <p className="text-muted-foreground text-xs">
        {feature.description}
      </p>
    </div>
  )
}