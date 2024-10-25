"use client"

import React from "react";

import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/seperator"
import { Docker } from "@/components/docker";

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
          <Button className=" p-0 h-auto text-sm md:text-base" variant="link">
            GitHub
          </Button>
        </a>
        .
      </div>
      <Separator className="my-8" />
      <p className="text-muted-foreground text-sm md:text-base">Get started quickly with Docker:</p>
      <Docker />
      <Separator className="my-8" />
      <div className="text-muted-foreground">
        Built by{" "}
        <a
          href="https://nishank.vercel.app"
          target="_blank"
        >
          <Button className=" p-0 h-auto text-base" variant="link">
            Nishank
          </Button>
        </a>
        .
      </div>
    </div>
  );
}