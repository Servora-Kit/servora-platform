import { Link } from '@tanstack/react-router'
import ThemeToggle from './ThemeToggle'

const navLinkClass =
  'rounded-md px-3 py-2 text-sm font-medium text-muted-foreground no-underline transition-colors hover:bg-accent hover:text-accent-foreground'
const navLinkActiveClass = 'text-foreground bg-accent'

export default function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-border bg-background/95 px-4 backdrop-blur supports-backdrop-filter:bg-background/80">
      <nav className="mx-auto flex max-w-5xl flex-wrap items-center gap-x-3 gap-y-2 py-3 sm:py-4">
        <h2 className="m-0 shrink-0 text-base font-semibold tracking-tight">
          <Link
            to="/"
            className="inline-flex items-center gap-2 rounded-full border border-border bg-muted px-3 py-1.5 text-sm font-medium text-foreground no-underline hover:bg-muted/80 sm:px-4 sm:py-2"
          >
            IAM
          </Link>
        </h2>

        <div className="ml-auto flex items-center gap-1.5 sm:ml-0 sm:gap-2">
          <a
            href="https://x.com/tan_stack"
            target="_blank"
            rel="noreferrer"
            className="hidden rounded-md p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground sm:block"
            aria-label="Follow TanStack on X"
          >
            <svg viewBox="0 0 16 16" aria-hidden width="24" height="24">
              <path
                fill="currentColor"
                d="M12.6 1h2.2L10 6.48 15.64 15h-4.41L7.78 9.82 3.23 15H1l5.14-5.84L.72 1h4.52l3.12 4.73L12.6 1zm-.77 12.67h1.22L4.57 2.26H3.26l8.57 11.41z"
              />
            </svg>
          </a>
          <a
            href="https://github.com/TanStack"
            target="_blank"
            rel="noreferrer"
            className="hidden rounded-md p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground sm:block"
            aria-label="TanStack GitHub"
          >
            <svg viewBox="0 0 16 16" aria-hidden width="24" height="24">
              <path
                fill="currentColor"
                d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"
              />
            </svg>
          </a>
          <ThemeToggle />
        </div>

        <div className="order-3 flex w-full flex-wrap items-center gap-x-1 gap-y-1 pb-1 text-sm font-medium sm:order-2 sm:w-auto sm:flex-nowrap sm:pb-0">
          <Link
            to="/"
            className={navLinkClass}
            activeProps={{ className: navLinkActiveClass }}
          >
            Home
          </Link>
          <Link
            to="/about"
            className={navLinkClass}
            activeProps={{ className: navLinkActiveClass }}
          >
            About
          </Link>
          <a
            href="https://tanstack.com/start/latest/docs/framework/react/overview"
            target="_blank"
            rel="noreferrer"
            className={navLinkClass}
          >
            Docs
          </a>
          <details className="relative w-full sm:w-auto">
            <summary className={`cursor-pointer list-none ${navLinkClass}`}>
              Demos
            </summary>
            <div className="mt-2 min-w-56 rounded-lg border border-border bg-popover p-2 shadow-md sm:absolute sm:right-0">
              <Link
                to="/demo/table"
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground no-underline hover:bg-accent hover:text-accent-foreground"
              >
                TanStack Table
              </Link>
              <Link
                to="/demo/form/simple"
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground no-underline hover:bg-accent hover:text-accent-foreground"
              >
                Simple Form
              </Link>
              <Link
                to="/demo/form/address"
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground no-underline hover:bg-accent hover:text-accent-foreground"
              >
                Address Form
              </Link>
              <Link
                to="/demo/store"
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground no-underline hover:bg-accent hover:text-accent-foreground"
              >
                Store
              </Link>
              <Link
                to="/demo/tanstack-query"
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground no-underline hover:bg-accent hover:text-accent-foreground"
              >
                TanStack Query
              </Link>
              <Link
                to="/demo/sentry/testing"
                className="block rounded-md px-3 py-2 text-sm text-muted-foreground no-underline hover:bg-accent hover:text-accent-foreground"
              >
                Sentry
              </Link>
            </div>
          </details>
        </div>
      </nav>
    </header>
  )
}
