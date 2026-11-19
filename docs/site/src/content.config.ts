import { defineCollection } from 'astro:content';
import { z } from 'astro/zod';
import { docsLoader, i18nLoader } from '@astrojs/starlight/loaders';
import { docsSchema, i18nSchema } from '@astrojs/starlight/schema';

import { Standing, StandingPages } from './lib/docs-origin.mjs';

// The pages of no version are built by the standing build alone. Every other
// build loads the collection without them, so such a page has one address and
// is in no version's routes, sitemap, or search index.
const loader = docsLoader();
const versioned = (process.env.DOCS_VERSION || 'edge') !== Standing;
const docs = {
	...loader,
	load: async (context: Parameters<typeof loader.load>[0]) => {
		await loader.load(context);
		if (versioned) {
			for (const page of StandingPages) context.store.delete(page);
		}
	},
};

// Page metadata the layout and the head need beyond Starlight's own fields.
//
// `eyebrow` is the small uppercase label above the page title. `accent` names a
// word of the title that is set in italic teal; the title renderer matches it
// once, so a word that does not occur in the title is a visible mistake rather
// than a silent one.
//
// `card` is the picture of a shared link for one page: a file of public/, 1200
// by 630 pixels, and the words that describe it. A page without one gets the
// card of the site.
const pageMetadata = z.object({
	eyebrow: z.string().min(1),
	accent: z.string().min(1).optional(),
	lead: z.string().min(1).optional(),
	noindex: z.boolean().optional(),
	card: z.object({ file: z.string().regex(/^[a-z0-9-]+\.(?:jpg|png)$/), alt: z.string().min(1) }).optional(),
});

export const collections = {
	docs: defineCollection({
		loader: docs,
		schema: docsSchema({ extend: pageMetadata }),
	}),
	// Starlight's interface strings, where this design words them differently:
	// src/content/i18n/en.json.
	i18n: defineCollection({ loader: i18nLoader(), schema: i18nSchema() }),
};
