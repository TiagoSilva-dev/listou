import type { InputHTMLAttributes, ReactNode, TextareaHTMLAttributes } from "react";
import { cn } from "../cn";

const controlBase = [
  "w-full rounded-control border border-line bg-surface px-4 text-[0.9375rem] text-ink",
  "placeholder:text-ink-muted/70 transition-shadow duration-(--duration-fast)",
  "focus:border-primary/40 focus:shadow-focus focus:outline-none",
  "aria-invalid:border-danger/50",
].join(" ");

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn(controlBase, "h-12", className)} {...props} />;
}

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea className={cn(controlBase, "min-h-28 py-3 leading-relaxed", className)} {...props} />
  );
}

export interface FieldProps {
  label: string;
  htmlFor: string;
  hint?: string;
  error?: string;
  optional?: boolean;
  children: ReactNode;
  className?: string;
}

export function Field({ label, htmlFor, hint, error, optional, children, className }: FieldProps) {
  return (
    <div className={cn("flex flex-col gap-2", className)}>
      <label htmlFor={htmlFor} className="text-ink flex items-baseline gap-2 text-sm font-semibold">
        {label}
        {optional ? <span className="text-ink-muted text-xs font-normal">opcional</span> : null}
      </label>
      {children}
      {error ? (
        <p role="alert" className="text-danger text-sm">
          {error}
        </p>
      ) : hint ? (
        <p className="text-ink-muted text-sm">{hint}</p>
      ) : null}
    </div>
  );
}
