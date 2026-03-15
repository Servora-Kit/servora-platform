import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/about')({
  component: About,
})

function About() {
  return (
    <main className="mx-auto max-w-5xl px-4 py-12">
      <section className="rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          About
        </p>
        <h1 className="mb-3 text-4xl font-bold text-foreground sm:text-5xl">
          A small template with room to grow.
        </h1>
        <p className="m-0 max-w-3xl text-base leading-8 text-muted-foreground">
          TanStack Start gives you type-safe routing, server functions, and
          modern SSR defaults. Use this as the frontend template for Servora
          microservices, then add your own routes, shadcn/ui components, and API
          integration.
        </p>
      </section>
    </main>
  )
}
