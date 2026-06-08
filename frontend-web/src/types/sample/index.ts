export type SampleStatus = "active" | "inactive" | "archived";

export interface Sample {
  id: number;
  name: string;
  slug: string;
  status: SampleStatus;
  createdAt: string;
  updatedAt: string;
}

export interface SampleListResponse {
  items: Sample[];
  total: number;
}
