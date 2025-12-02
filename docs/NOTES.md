# NOTES
A document of general notes and random sporadic stuff that myself (Les) and maybe other developers can use to dump stuff in.

**General rule of thumb for this document:** if you point or section gets too big, it likely needs its own document.

## Infrustructure for custom themes (Using --imag-100 - --imag-5)
The idea is that currently I just have two test themes, [`viz-blue`](../viz/src/lib/styles/scss/viz-blue.scss) and [`viz-black`](../viz/src/lib/styles/scss/viz-black.scss) written in SCSS, compiled by Vite and imported at runtime by the [`hooks.server.ts`](../viz/src/hooks.server.ts) file. However, this app is a SPA and embedded into the [`cmd/api`](../cmd/api/main.go) file (name needs to change to server tbh) and whatever SCSS is used needs to get compiled again.

We could use [`https://github.com/bep/godartsass`](https://github.com/bep/godartsass) to compile at runtime or maybe just a seperate process that could be launched once when necessary.

Themes are based on these five SCSS variables:

````scss
// viz-black.scss
$bg-color: #191919;
$base-color: #1f1f1f;
$text-color: #f7f9f9;
$primary-color: #1885df;
$secondary-color: #06589b;
````

A [`viz-theme.scss.tmpl`](../viz/src/lib/styles/scss/viz-theme.scss.tmpl) file could be used maybe based on what the current structure of the theme files look like and then all those variables can just be injected from wherever they are set (user settings for example)

idk this makes sense to me now and I'm sure a final solution could be figured out and way better than this idea. I just don't want non-developers to be writing CSS just make their environment look good for themselves.

Authored by [`@garvageart`](https://github.com/garvageart)