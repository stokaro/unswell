/**
 * Rewrites a root-relative Markdown link to this version's base.
 *
 * Every page of the site lives under `/<version>/`, and the version is chosen
 * per build. Neither spelling of a plain Markdown link survives that on its own:
 * a relative `getting-started/` resolves against the current page rather than
 * the site, so it breaks on every page except the index, and an absolute
 * `/getting-started/` lands outside the version folder. Astro rewrites neither.
 *
 * So pages write `/getting-started/`, meaning "this site's own route", and this
 * plugin turns it into `/<version>/getting-started/` while the page is still
 * hast. A link that already carries the base is left alone, and so is anything
 * with a scheme, a protocol-relative prefix, or a bare fragment.
 *
 * A Sätteri hast plugin, registered on `markdown.processor` in astro.config.mjs.
 * Components take the same rewrite from `import.meta.env.BASE_URL` instead:
 * their attributes never reach the Markdown pipeline.
 */

const EXTERNAL = /^(?:[a-z][a-z0-9+.-]*:|\/\/)/i;

export default function markdownBaseLinks(base) {
	if (!base.startsWith('/') || !base.endsWith('/')) {
		throw new Error(`base must have a leading and a trailing slash: ${base}`);
	}
	return {
		name: 'unswell-base-links',
		element: {
			filter: ['a'],
			visit(node, ctx) {
				const href = node.properties?.href;
				if (typeof href !== 'string') return;
				if (href === '' || href.startsWith('#') || EXTERNAL.test(href)) return;
				if (!href.startsWith('/')) return;
				if (href === base || href.startsWith(base)) return;
				ctx.setProperty(node, 'href', `${base}${href.slice(1)}`);
			},
		},
	};
}
