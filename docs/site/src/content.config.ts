import { defineCollection } from 'astro:content';
import { z } from 'astro/zod';
import { docsLoader, i18nLoader } from '@astrojs/starlight/loaders';
import { docsSchema, i18nSchema } from '@astrojs/starlight/schema';

// Page metadata the layout and the head need beyond Starlight's own fields.
//
// `eyebrow` is the small uppercase label above the page title. `accent` names a
// word of the title that is set in italic teal; the title renderer matches it
// once, so a word that does not occur in the title is a visible mistake rather
// than a silent one.
const pageMetadata = z.object({
	eyebrow: z.string().min(1),
	accent: z.string().min(1).optional(),
	lead: z.string().min(1).optional(),
	noindex: z.boolean().optional(),
});

export const collections = {
	docs: defineCollection({
		loader: docsLoader(),
		schema: docsSchema({ extend: pageMetadata }),
	}),
	// Starlight's interface strings, where this design words them differently:
	// src/content/i18n/en.json.
	i18n: defineCollection({ loader: i18nLoader(), schema: i18nSchema() }),
};
