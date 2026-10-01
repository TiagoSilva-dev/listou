import { Skeleton } from "@listou/ui";

export default function Loading() {
  return (
    <div className="flex flex-col gap-8">
      <Skeleton className="h-10 w-72" />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="rounded-card h-64" />
        ))}
      </div>
    </div>
  );
}
