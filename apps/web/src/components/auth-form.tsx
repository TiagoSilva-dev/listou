"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { LoginInput, RegisterInput, User } from "@listou/types";
import { Button, Field, Input } from "@listou/ui";
import { z } from "zod";
import { api, applyFieldErrors } from "@/lib/api";

const userResponse = z.object({ user: User });

export function LoginForm({ next }: { next: string }) {
  const router = useRouter();
  const [formError, setFormError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({ resolver: zodResolver(LoginInput) });

  return (
    <form
      noValidate
      className="flex flex-col gap-5"
      onSubmit={handleSubmit(async (values) => {
        setFormError(null);
        try {
          await api("/auth/login", userResponse, { method: "POST", body: values });
          router.replace(next);
          router.refresh();
        } catch (err) {
          setFormError(applyFieldErrors(err, setError));
        }
      })}
    >
      <Field label="E-mail" htmlFor="email" error={errors.email?.message}>
        <Input
          id="email"
          type="email"
          autoComplete="email"
          inputMode="email"
          placeholder="voce@exemplo.com"
          aria-invalid={!!errors.email}
          {...register("email")}
        />
      </Field>
      <Field label="Senha" htmlFor="password" error={errors.password?.message}>
        <Input
          id="password"
          type="password"
          autoComplete="current-password"
          aria-invalid={!!errors.password}
          {...register("password")}
        />
      </Field>
      {formError ? (
        <p role="alert" className="rounded-control bg-danger-soft text-danger px-4 py-3 text-sm">
          {formError}
        </p>
      ) : null}
      <Button type="submit" size="lg" block loading={isSubmitting}>
        Entrar
      </Button>
      <p className="text-ink-muted text-center text-sm">
        Ainda não tem conta?{" "}
        <Link
          href={`/criar-conta?next=${encodeURIComponent(next)}`}
          className="text-primary font-semibold hover:underline"
        >
          Criar conta grátis
        </Link>
      </p>
    </form>
  );
}

export function RegisterForm({ next }: { next: string }) {
  const router = useRouter();
  const [formError, setFormError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({ resolver: zodResolver(RegisterInput) });

  return (
    <form
      noValidate
      className="flex flex-col gap-5"
      onSubmit={handleSubmit(async (values) => {
        setFormError(null);
        try {
          await api("/auth/register", userResponse, { method: "POST", body: values });
          router.replace(next);
          router.refresh();
        } catch (err) {
          setFormError(applyFieldErrors(err, setError));
        }
      })}
    >
      <Field label="Seu nome" htmlFor="name" error={errors.name?.message}>
        <Input
          id="name"
          autoComplete="name"
          placeholder="Como você gosta de ser chamado"
          aria-invalid={!!errors.name}
          {...register("name")}
        />
      </Field>
      <Field label="E-mail" htmlFor="email" error={errors.email?.message}>
        <Input
          id="email"
          type="email"
          autoComplete="email"
          inputMode="email"
          placeholder="voce@exemplo.com"
          aria-invalid={!!errors.email}
          {...register("email")}
        />
      </Field>
      <Field
        label="Senha"
        htmlFor="password"
        hint="Pelo menos 8 caracteres."
        error={errors.password?.message}
      >
        <Input
          id="password"
          type="password"
          autoComplete="new-password"
          aria-invalid={!!errors.password}
          {...register("password")}
        />
      </Field>
      {formError ? (
        <p role="alert" className="rounded-control bg-danger-soft text-danger px-4 py-3 text-sm">
          {formError}
        </p>
      ) : null}
      <Button type="submit" size="lg" block loading={isSubmitting}>
        Criar minha conta
      </Button>
      <p className="text-ink-muted text-center text-sm">
        Já tem conta?{" "}
        <Link
          href={`/entrar?next=${encodeURIComponent(next)}`}
          className="text-primary font-semibold hover:underline"
        >
          Entrar
        </Link>
      </p>
    </form>
  );
}
