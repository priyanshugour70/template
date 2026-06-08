/** Common DTO bits shared across feature types. Mirrors backend conventions. */

export type ID = string;
export type ISODate = string;
export type JSONValue = string | number | boolean | null | JSONValue[] | { [k: string]: JSONValue };
export type JSONObject = Record<string, JSONValue>;

export interface BaseEntity {
  id: ID;
  createdAt: ISODate;
  updatedAt: ISODate;
  deletedAt?: ISODate | null;
  createdBy?: ID | null;
  updatedBy?: ID | null;
  deletedBy?: ID | null;
}

export interface PaginatedQuery {
  page?: number;
  limit?: number;
  q?: string;
  sort?: string;
}
