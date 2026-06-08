"use client";

import { useQuery } from "@tanstack/react-query";

import { sampleService } from "@/services/sample";

export function useSamples(page = 1, limit = 20, search?: string) {
  return useQuery({
    queryKey: ["samples", page, limit, search ?? ""],
    queryFn: async () => {
      const res = await sampleService.list(page, limit, search);
      if (!res.success) throw new Error(res.error?.message ?? "list samples failed");
      return res.data!;
    },
  });
}
