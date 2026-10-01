import { Skeleton } from "@listou/ui";

export default function Loading() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-10 w-72" />
      {[0, 1, 2].map((i) => (
        <Skeleton key={i} className="rounded-card h-28" />
      ))}
    </div>
  );
}
