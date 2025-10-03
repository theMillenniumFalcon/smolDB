"use client"

import React from "react";
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/seperator"
import { Docker } from "@/components/docker";
import { Shield } from "lucide-react";

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

      {/* Durability Feature */}
      <div className="bg-muted/50 rounded-lg p-4 mb-8 border border-border">
        <div className="flex items-center gap-2 mb-2">
          <Shield className="w-4 h-4 text-primary" />
          <h2 className="text-base font-semibold">Production-Ready Durability</h2>
        </div>
        <p className="text-muted-foreground text-sm">
          Now includes Write-Ahead Logging with configurable fsync modes, group commit, and automatic crash recovery.
        </p>
      </div>

      <p className="text-muted-foreground text-sm md:text-base">Get started quickly with the API via Docker:</p>
      <Docker />
      <Separator className="my-8" />
      <div className="text-muted-foreground text-sm md:text-base">
        Built by{" "}
        <a
          href="https://nishank.vercel.app"
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