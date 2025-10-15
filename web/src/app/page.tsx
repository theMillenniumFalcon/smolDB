"use client"

import React from "react";
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/seperator"
import { Docker } from "@/components/docker";
import { FeatureCard } from "@/components/ui/feature-card";
import { features } from "@/lib/features";

export default function Home() {
  return (
    <div className="w-full mt-8">
      <h1 className="md:text-3xl sm:text-2xl text-xl mb-4 font-semibold">
        smolDB
      </h1>
      <div className="text-muted-foreground text-sm md:text-base">
        A JSON document-based database that relies on key based access to achieve
        O(1) access time. Check the repository out on{" "}
        <a
          href="https://github.com/themillenniumfalcon/smolDB"
          target="_blank"
        >
          <Button className="p-0 h-auto text-sm md:text-base" variant="link">
            GitHub
          </Button>
        </a>
        .
      </div>

      <Separator className="my-8" />

      {/* Features Section */}
      <div className="space-y-4 mb-8">
        <h2 className="text-sm font-medium text-muted-foreground">Key Features</h2>
        <div className="grid md:grid-cols-2 gap-2">
          {features.map((feature, index) => (
            <FeatureCard key={index} feature={feature} />
          ))}
        </div>
      </div>

      <p className="text-muted-foreground text-sm md:text-base">Get started quickly with the API via Docker:</p>
      <Docker />
      <Separator className="my-8" />
      <div className="text-muted-foreground text-sm md:text-base">
        Built by{" "}
        <a
          href="https://github.com/themillenniumfalcon"
          target="_blank"
        >
          <Button className="p-0 h-auto text-sm md:text-base" variant="link">
            Nishank
          </Button>
        </a>
        .
      </div>
    </div>
  );
}