import { createFileRoute, Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

export const Route = createFileRoute('/')({ component: App })

function App() {
  return (
    <main className="mx-auto max-w-5xl px-4 pb-10 pt-8">
      <section className="rounded-2xl border border-border bg-card px-6 py-10 shadow-sm sm:px-10 sm:py-14">
        <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          IAM · Template
        </p>
        <h1 className="mb-5 max-w-3xl text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
          Start simple, ship quickly.
        </h1>
        <p className="mb-8 max-w-2xl text-base text-muted-foreground sm:text-lg">
          This base template keeps things light: Tailwind + shadcn/ui, TanStack
          Router, Store, Form, Table, Query. Use it as the frontend template for
          Servora microservices.
        </p>
        <div className="flex flex-wrap gap-3">
          <Button asChild variant="default">
            <Link to="/about">About</Link>
          </Button>
          <Button asChild variant="outline">
            <a
              href="https://tanstack.com/router"
              target="_blank"
              rel="noopener noreferrer"
            >
              Router Guide
            </a>
          </Button>
        </div>
      </section>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          [
            'Type-Safe Routing',
            'Routes and links stay in sync across every page.',
          ],
          [
            'Server Functions',
            'Call server code from your UI without API boilerplate.',
          ],
          [
            'Streaming by Default',
            'Ship progressively rendered responses for faster loads.',
          ],
          [
            'Tailwind + shadcn',
            'Utility-first styling and reusable UI components.',
          ],
        ].map(([title, desc]) => (
          <article
            key={title}
            className="rounded-xl border border-border bg-card p-5 transition-colors hover:border-border/80"
          >
            <h2 className="mb-2 text-base font-semibold text-foreground">
              {title}
            </h2>
            <p className="m-0 text-sm text-muted-foreground">{desc}</p>
          </article>
        ))}
      </section>

      <section className="mt-8 rounded-2xl border border-border bg-card p-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Quick Start
        </p>
        <ul className="m-0 list-disc space-y-2 pl-5 text-sm text-muted-foreground">
          <li>
            Edit <code>src/routes/index.tsx</code> to customize the home page.
          </li>
          <li>
            Update <code>src/components/Header.tsx</code> and{' '}
            <code>src/components/Footer.tsx</code> for navigation and branding.
          </li>
          <li>
            Add routes under <code>src/routes</code> and theme tokens in{' '}
            <code>src/styles.css</code>.
          </li>
        </ul>
      </section>
    </main>
  )
}
