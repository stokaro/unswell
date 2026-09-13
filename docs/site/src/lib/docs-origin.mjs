// Where the documentation is published, declared once.
//
// The site is served at the apex of its own domain, so a page address is the
// origin plus the version: there is no path segment between them. Everything
// else derives from the two values here. astro.config.mjs builds `site` and
// `base`, the version generator builds the apex redirect, and the gates build
// the addresses they fetch.
//
// This module is plain ESM with no imports so that astro.config.mjs, the .mjs
// gates and the version generator can all read it. A workflow or a Markdown
// page cannot import it and names the address literally; the docs workflow
// asserts that the published CNAME and the built output agree with this file.

// Origin is the scheme and host the documentation is served from, without a
// trailing slash.
export const Origin = 'https://unswell.dev';

// Host is the same address as the Pages CNAME file spells it.
export const Host = 'unswell.dev';

// BasePath is the site-root-relative prefix every page of one version lives
// under, with leading and trailing slashes. It is a function because the
// version is chosen per build.
export function BasePath(version) {
	return `/${version}/`;
}

// RootURL is the absolute URL of a file published at the site root, beside the
// version folders rather than inside one.
export function RootURL(name) {
	return `${Origin}/${name}`;
}

// PageURL is the absolute URL of one page of one version. Route is written the
// way the site's own routes are: no leading slash, one trailing slash.
export function PageURL(version, route) {
	return `${Origin}${BasePath(version)}${route}`;
}

// PlaygroundOrigin is where Unswell runs in a browser tab. It is a site of its
// own rather than a page here, so the documentation never loads it for a reader
// who did not ask for it.
export const PlaygroundOrigin = 'https://play.unswell.dev';

// RepositoryURL is the source of every claim on this site.
export const RepositoryURL = 'https://github.com/stokaro/unswell';
