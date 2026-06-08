"use server";

import { sampleService } from "@/services/sample";
import type { Sample } from "@/types/sample";

export async function createSampleAction(formData: FormData): Promise<{
  ok: boolean;
  data?: Sample;
  error?: string;
}> {
  const name = String(formData.get("name") ?? "").trim();
  if (!name) return { ok: false, error: "Name is required" };

  const res = await sampleService.create(name);
  if (!res.success || !res.data) {
    return { ok: false, error: res.error?.message ?? "create failed" };
  }
  return { ok: true, data: res.data };
}
