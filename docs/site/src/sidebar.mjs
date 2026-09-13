// The navigation, as a value in its own module.
//
// A plain Node script cannot import astro.config.mjs, because Starlight's entry
// point is TypeScript inside node_modules and Node refuses to strip types
// there. The link gate reads this module instead, so the sidebar the gate
// checks is the sidebar the site renders.
import { PlaygroundOrigin } from './lib/docs-origin.mjs';

export const sidebar = [
	{
		label: 'Start',
		items: [
			{ label: 'Overview', slug: '' },
			{ label: 'Getting started', slug: 'getting-started' },
			{ label: 'Playground', link: PlaygroundOrigin, attrs: { target: '_blank', rel: 'noopener' } },
		],
	},
	{
		label: 'Understand',
		items: [
			{ label: 'How it works', slug: 'how-it-works' },
			{ label: 'Rules catalog', slug: 'rules' },
		],
	},
	{
		label: 'Configure',
		items: [{ label: 'Configuration & profiles', slug: 'configuration' }],
	},
	{
		label: 'Operate',
		items: [
			{ label: 'Reports, gate & exit codes', slug: 'reports' },
			{ label: 'CI, Action & MCP', slug: 'ci-action-mcp' },
		],
	},
];

// routes lists every internal page the sidebar names, as the site spells its
// own routes: no leading slash, one trailing slash. The link gate resolves
// internal targets against this list.
export const routes = sidebar
	.flatMap((group) => group.items)
	.filter((item) => typeof item.slug === 'string')
	.map((item) => `${item.slug}/`);
