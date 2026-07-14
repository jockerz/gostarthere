# Project Intelligence Guide

## Build / Lint / Test Commands

```bash
npm run dev              # Start dev server (vite)
npm run build            # Production build
npm run preview          # Preview production build
npm run check            # Type check via svelte-check
npm run check:watch      # Type check in watch mode
npm run prepare          # Sync svelte-kit types
```

There is no test framework or linter configured. Run `npm run check` for type checking before committing.

## Tech Stack

| Technology | Purpose |
|---|---|
| **SvelteKit 2** | Fullstack framework with SSR |
| **Svelte 5** | Runes-based reactive components |
| **shadcn/ui** (Svelte port, Lyra style) | Component library |
| **Tailwind CSS v4** | Utility-first CSS |
| **TypeScript** (strict) | Static typing |
| **Lucide** | Icons via `@lucide/svelte` |
| **Vite 8** | Build tool |

## Project Structure

```
src/
├── lib/
│   ├── components/
│   │   ├── ui/           # shadcn/ui components (add via shadcn-svelte)
│   │   └── custom/       # Project-specific components
│   │       └── [feature]/
│   │           ├── ComponentName.svelte
│   │           └── types.ts
│   ├── hooks/            # Svelte rune-based reusable logic
│   ├── services/         # API calls & integrations
│   ├── stores/           # Svelte stores (if not using runes-only)
│   ├── types/            # Shared TypeScript types
│   ├── utils.ts          # Utilities (cn, etc.)
│   └── index.ts
├── routes/
│   ├── +layout.svelte    # Root layout
│   ├── +page.svelte      # Home page
│   ├── layout.css        # Tailwind + shadcn/ui theme
│   ├── auth/             # Auth pages (no nav layout)
│   ├── dashboard/        # Requires authentication
│   └── admin/            # Requires auth + admin role
├── app.d.ts              # App namespace types
└── app.html              # HTML entry point
```

## Path Aliases

```typescript
import { cn } from "$lib/utils";    // $lib → src/lib (SvelteKit default)
import { cn } from "@/utils";       // @/    → src/lib (configured in svelte.config.js)
import { PUBLIC_API_URL } from "$env/static/public";  // Env vars
```

## Code Style Guidelines

### General
- **Indentation**: tabs
- **Quotes**: double quotes for imports/strings (e.g. `import { foo } from "bar"`)
- **Semicolons**: required
- **Trailing commas**: none in package.json configs (project style varies — follow existing file)
- **TypeScript**: strict mode enabled

### Imports Order
1. External packages (svelte, shadcn-svelte, etc.)
2. `$env/*` aliases
3. `$lib/*` or `@/*` aliases (with a blank line between groups)
4. Relative imports

### Svelte 5 (Runes)
```svelte
<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";

  interface Props extends HTMLAttributes<HTMLDivElement> {
    variant?: "default" | "primary";
    size?: "sm" | "md" | "lg";
  }

  let { variant = "default", size = "md", ...rest }: Props = $props();
</script>

<div class={cn("component", variant)} {...rest}>
  {#if size === "lg"}<slot />{:else}<slot />{/if}
</div>
```
- Use `$props()` for component props. Use `$state()` for local reactive state. Use `$derived()` for computed values. Use `$effect()` for side effects.
- Use `{@render children()}` for rendering children (Svelte 5 snippets), not `<slot />`.
- Keep template logic minimal; use `{#if}`/`{:else if}`/`{:else}` and `{#each}` blocks.

### Components
- **shadcn/ui** components are in `src/lib/components/ui/` — add via `npx shadcn-svelte@latest add [name]`
- **Custom components** go in `src/lib/components/custom/[feature]/` — use PascalCase for filenames, co-locate types in `types.ts`

### Naming
| Item | Convention | Example |
|---|---|---|
| Components | PascalCase | `UserCard.svelte` |
| Routes | kebab-case | `/user-settings` |
| Stores/Services | camelCase | `userStore.ts` |
| Types | PascalCase | `ApiResponse.ts` |
| Utils | camelCase | `formatDate.ts` |

### CSS
```css
/* Tailwind v4 @import syntax (no @tailwind directives) */
@import "tailwindcss";
@import "tw-animate-css";
@import "shadcn-svelte/tailwind.css";
```
- Use `cn()` utility (`clsx` + `tailwind-merge`) for conditional classes:
  ```typescript
  import { cn } from "$lib/utils";
  <div class={cn("base-class", condition && "extra-class")} />
  ```
- CSS custom properties for theming in `layout.css` (oklch color space)

### Environment Variables
- Public: `PUBLIC_` prefix, accessed via `$env/static/public`
- Private: `import { SECRET } from "$env/static/private"`
- Add to `.env.local` (gitignored) and `.env.example` (committed)

## shadcn/ui

```bash
npx shadcn-svelte@latest add button
npx shadcn-svelte@latest add card dialog
npx shadcn-svelte@latest add form input
```

Uses **Lyra style**, **lucide** icons, **neutral base color**. Config in `components.json`. Do not manually edit generated `src/lib/components/ui/*` files — extend in `custom/` instead.

## VSCode Setup

Recommended extensions (in `.vscode/extensions.json`):
- `svelte.svelte-vscode`
- `bradlc.vscode-tailwindcss`

CSS files are associated with `tailwindcss` language mode.
