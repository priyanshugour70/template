"use client";

import { Building2, ExternalLink, Mail, MapPin, Phone, Plus } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useCreateOrganization, useOrganizations } from "@/hooks/tenant/useTenantQueries";
import { usePermissions } from "@/providers";

function slugify(s: string): string {
  return s
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export default function OrganizationsPage() {
  const orgsQ = useOrganizations();
  const create = useCreateOrganization();
  const { has } = usePermissions();
  const [open, setOpen] = useState(false);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Organizations</h1>
          <p className="text-muted-foreground mt-1">
            Workspaces inside your tenant. Each can have its own users, roles, and subscription.
          </p>
        </div>
        {has("org.create") && (
          <Button onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" />
            New organization
          </Button>
        )}
      </div>

      {orgsQ.isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-60" />
          ))}
        </div>
      ) : !orgsQ.data?.length ? (
        <Card>
          <CardContent className="p-10 text-center text-sm text-muted-foreground">
            No organizations yet.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {orgsQ.data.map((o) => (
            <Card key={o.id} className="overflow-hidden">
              <CardHeader>
                <div className="flex items-start gap-3">
                  {o.logoUrl ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={o.logoUrl} alt={o.name} className="h-12 w-12 rounded-md object-cover" />
                  ) : (
                    <div className="h-12 w-12 rounded-md bg-primary/10 flex items-center justify-center">
                      <Building2 className="h-5 w-5 text-primary" />
                    </div>
                  )}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold truncate">{o.name}</h3>
                      {o.isDefault && <Badge>default</Badge>}
                    </div>
                    <div className="text-xs text-muted-foreground truncate">/{o.slug}</div>
                  </div>
                  <Badge variant={o.status === "active" ? "success" : "warning"}>
                    {o.status}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                {o.description && (
                  <p className="text-muted-foreground line-clamp-2">{o.description}</p>
                )}
                <dl className="space-y-1.5 text-sm">
                  {o.industry && (
                    <Row label="Industry" value={o.industry} />
                  )}
                  {o.size && <Row label="Team size" value={o.size} />}
                  {(o.city || o.country) && (
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <MapPin className="h-3.5 w-3.5" />
                      <span>{[o.city, o.country].filter(Boolean).join(", ")}</span>
                    </div>
                  )}
                  {o.contactEmail && (
                    <div className="flex items-center gap-2 text-muted-foreground truncate">
                      <Mail className="h-3.5 w-3.5" />
                      <span className="truncate">{o.contactEmail}</span>
                    </div>
                  )}
                  {o.contactPhone && (
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <Phone className="h-3.5 w-3.5" />
                      <span>{o.contactPhone}</span>
                    </div>
                  )}
                  {o.websiteUrl && (
                    <div className="flex items-center gap-2">
                      <ExternalLink className="h-3.5 w-3.5 text-muted-foreground" />
                      <a
                        href={o.websiteUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="text-primary hover:underline truncate"
                      >
                        {o.websiteUrl}
                      </a>
                    </div>
                  )}
                </dl>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New organization</DialogTitle>
            <DialogDescription>Create another workspace inside your tenant.</DialogDescription>
          </DialogHeader>
          <CreateOrgForm
            pending={create.isPending}
            error={create.isError ? (create.error as Error).message : null}
            onCancel={() => setOpen(false)}
            onSubmit={async (values) => {
              await create.mutateAsync(values);
              setOpen(false);
            }}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium truncate max-w-[60%]">{value}</span>
    </div>
  );
}

function CreateOrgForm(props: {
  pending: boolean;
  error: string | null;
  onCancel: () => void;
  onSubmit: (v: {
    name: string;
    slug: string;
    description?: string;
    industry?: string;
    size?: string;
    websiteUrl?: string;
    contactEmail?: string;
    contactPhone?: string;
    country?: string;
    city?: string;
  }) => void;
}) {
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [industry, setIndustry] = useState("");
  const [size, setSize] = useState("");
  const [website, setWebsite] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [country, setCountry] = useState("");
  const [city, setCity] = useState("");

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        props.onSubmit({
          name: name.trim(),
          slug: slug || slugify(name),
          description: description.trim() || undefined,
          industry: industry.trim() || undefined,
          size: size.trim() || undefined,
          websiteUrl: website.trim() || undefined,
          contactEmail: email.trim() || undefined,
          contactPhone: phone.trim() || undefined,
          country: country.trim() || undefined,
          city: city.trim() || undefined,
        });
      }}
      className="space-y-4 max-h-[70vh] overflow-y-auto"
    >
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="org-name">Name</Label>
          <Input
            id="org-name"
            required
            value={name}
            onChange={(e) => {
              const v = e.target.value;
              setName(v);
              if (!slug) setSlug(slugify(v));
            }}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="org-slug">Slug</Label>
          <Input
            id="org-slug"
            required
            value={slug}
            onChange={(e) => setSlug(slugify(e.target.value))}
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="org-desc">Description</Label>
        <Input id="org-desc" value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="org-industry">Industry</Label>
          <Input id="org-industry" value={industry} onChange={(e) => setIndustry(e.target.value)} placeholder="SaaS, Retail…" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="org-size">Team size</Label>
          <Input id="org-size" value={size} onChange={(e) => setSize(e.target.value)} placeholder="1-10, 50-200…" />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="org-website">Website</Label>
          <Input id="org-website" value={website} onChange={(e) => setWebsite(e.target.value)} placeholder="https://" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="org-email">Contact email</Label>
          <Input id="org-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="org-phone">Contact phone</Label>
          <Input id="org-phone" value={phone} onChange={(e) => setPhone(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="org-country">Country</Label>
          <Input id="org-country" value={country} onChange={(e) => setCountry(e.target.value)} />
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="org-city">City</Label>
        <Input id="org-city" value={city} onChange={(e) => setCity(e.target.value)} />
      </div>
      {props.error && <p className="text-sm text-destructive">{props.error}</p>}
      <DialogFooter>
        <Button type="button" variant="ghost" onClick={props.onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={props.pending || !name}>
          {props.pending ? "Creating…" : "Create organization"}
        </Button>
      </DialogFooter>
    </form>
  );
}
