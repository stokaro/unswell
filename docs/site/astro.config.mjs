// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import sitemap from '@astrojs/sitemap';
import { satteri } from '@astrojs/markdown-satteri';

import markdownBaseLinks from './src/lib/markdown-base-links.mjs';
import { sidebar } from './src/sidebar.mjs';
import { Origin, BuildBase, RepositoryURL, isStandingPath } from './src/lib/docs-origin.mjs';

// One version per build. The deploy runs this build once per version and copies
// each dist/ into its own folder of the Pages artifact, so a page built here
// only ever addresses its own version. One more build, `standing`, holds the
// pages that belong to no version and is served at the root of the site. The
// latest release is built a second way, with DOCS_AT_ROOT=1: its pages are
// served at the root, so the bare domain shows them with no redirect.
const site = Origin;
const base = BuildBase(process.env);

export default defineConfig({
	site,
	base,
	markdown: {
		// Astro 7 renders Markdown and MDX through Satteri. Pages write internal
		// links as site routes, `/rules/`, and the plugin puts this build's
		// version in front of them.
		processor: satteri({
			hastPlugins: [markdownBaseLinks(base, isStandingPath)],
		}),
	},
	integrations: [
		starlight({
			title: 'Unswell',
			description:
				'Unswell is an offline linter for lean English in documentation, comments and string literals.',
			// No `logo` option: SiteTitle.astro renders the product mark inline, so
			// the wordmark can carry two colors and sit beside the version picker.
			customCss: ['./src/styles/fonts.css', './src/styles/unswell.css'],
			lastUpdated: true,
			components: {
				Head: './src/components/Head.astro',
				PageTitle: './src/components/PageTitle.astro',
				Search: './src/components/Search.astro',
				SiteTitle: './src/components/SiteTitle.astro',
				SocialIcons: './src/components/HeaderLinks.astro',
			},
			social: [{ icon: 'github', label: 'GitHub', href: RepositoryURL }],
			expressiveCode: {
				themes: ['github-dark', 'github-light'],
				useStarlightUiThemeColors: true,
			},
			sidebar,
		}),
		// The sitemap lists this version's pages at their absolute addresses. The
		// apex redirect and versions.json are written by the assembly step, not
		// by Astro, so they are deliberately absent from it.
		sitemap(),
	],
});
