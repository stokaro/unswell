#!/usr/bin/env node
// Writes the files that belong at the Pages ROOT rather than inside a versioned
// documentation directory.
//
// The root holds one directory per version plus a few files that address the
// site as a whole. gen-versions.mjs writes versions.json, index.html, 404.html
// and a stub for each page of the default version. This script writes the rest: the CNAME that claims the domain,
// the .nojekyll marker, and robots.txt with the sitemap of the default version.
//
// The deploy assembles `_site/` from scratch on every run and uploads it whole,
// so there is no incremental Pages state a file can survive in. A root file
// exists after a deploy only because that deploy wrote it, which is why this is
// a workflow step rather than a file somebody uploaded once.
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { Host, Origin, StandingPages } from '../src/lib/docs-origin.mjs';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const defaultSiteRoot = join(scriptDir, '..');

// COPIED_ASSETS are published byte for byte from docs/site/public/. They live
// there rather than beside this script so that a change to them runs the docs
// workflow, whose path filter watches docs/site/**. Astro also copies public/
// into every version's dist/, which is harmless: the addresses that matter are
// the root ones this script writes.
//
// og-image.png is the picture of a shared link. Every page of every version
// names it at the root of the site, so a version folder's own copy is not the
// file a chat client asks for.
export const COPIED_ASSETS = [
	{ name: 'CNAME', source: 'public/CNAME' },
	{ name: '.nojekyll', source: 'public/.nojekyll' },
	{ name: 'og-image.png', source: 'public/og-image.png' },
];

// GENERATED_ROOT_FILES names what gen-versions.mjs writes into the same
// directory. This script does not write them; naming them here lets the Pages
// root gate compare the whole expected root against a real assembly.
export const GENERATED_ROOT_FILES = ['versions.json', 'index.html', '404.html'];

// renderRobots points crawlers at the sitemap of the version the apex serves.
// Every other version is reachable and indexable, but only one of them should
// be the copy a search engine prefers, and that is the one the canonical links
// of the default version already name.
//
// The pages of no version are in no version's sitemap, so they get a small
// sitemap of their own at the root.
//
// When the root holds the latest release, its sitemap is at the root as well.
export function renderRobots(defaultSlug, standing = [], atRoot = false) {
	const own = standing.length > 0 ? `Sitemap: ${Origin}/${STANDING_SITEMAP}\n` : '';
	const sitemap = atRoot ? `${Origin}/sitemap-index.xml` : `${Origin}/${defaultSlug}/sitemap-index.xml`;
	return `User-agent: *
Allow: /

Sitemap: ${sitemap}
${own}`;
}

export const STANDING_SITEMAP = 'sitemap-standing.xml';

export function renderStandingSitemap(standing) {
	const urls = standing.map((page) => `  <url><loc>${Origin}/${page}/</loc></url>\n`).join('');
	return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls}</urlset>
`;
}

// publish copies and writes every root file into siteDir and returns the names
// it wrote. A missing or empty source throws rather than being skipped: past
// this step the artifact uploads, and a missing CNAME takes the domain down
// with nothing red anywhere.
export function publish(siteDir, defaultSlug, siteRoot = defaultSiteRoot, standing = StandingPages, atRoot = false) {
	const written = [];
	for (const asset of COPIED_ASSETS) {
		const source = join(siteRoot, asset.source);
		if (!existsSync(source)) {
			throw new Error(`root asset source is missing: ${asset.source}`);
		}
		copyFileSync(source, join(siteDir, asset.name));
		written.push(asset.name);
	}
	const cname = readFileSync(join(siteDir, 'CNAME'), 'utf8').trim();
	if (cname !== Host) {
		throw new Error(`CNAME names ${cname}, not ${Host}`);
	}
	writeFileSync(join(siteDir, 'robots.txt'), renderRobots(defaultSlug, standing, atRoot));
	written.push('robots.txt');
	if (standing.length > 0) {
		writeFileSync(join(siteDir, STANDING_SITEMAP), renderStandingSitemap(standing));
		written.push(STANDING_SITEMAP);
	}
	return written;
}

function selftest() {
	const assert = (condition, message) => {
		if (!condition) throw new Error(message);
	};

	assert(renderRobots('edge').includes(`${Origin}/edge/sitemap-index.xml`), 'robots names the sitemap');
	assert(renderRobots('edge').includes('User-agent: *'), 'robots addresses every crawler');
	assert(!renderRobots('edge').includes(STANDING_SITEMAP), 'a site without pages of no version names one sitemap');
	assert(renderRobots('v1.0.0', [], true).includes(`Sitemap: ${Origin}/sitemap-index.xml`), 'a release at the root has its sitemap there');
	assert(!renderRobots('v1.0.0', [], true).includes('/v1.0.0/'), 'and robots names no version folder then');
	assert(renderRobots('edge', ['report']).includes(`${Origin}/${STANDING_SITEMAP}`), 'robots names the sitemap of those pages');
	assert(renderStandingSitemap(['report']).includes(`<loc>${Origin}/report/</loc>`), 'that sitemap lists the page at the root');

	const tmp = mkdtempSync(join(tmpdir(), 'unswell-docs-root-'));
	try {
		const fakeSiteRoot = join(tmp, 'site');
		mkdirSync(join(fakeSiteRoot, 'public'), { recursive: true });
		writeFileSync(join(fakeSiteRoot, 'public', 'CNAME'), `${Host}\n`);
		writeFileSync(join(fakeSiteRoot, 'public', '.nojekyll'), '');
		const assembly = join(tmp, '_site');
		mkdirSync(assembly);
		let withoutCard = false;
		try {
			publish(assembly, 'v0.1.0-alpha.3', fakeSiteRoot);
		} catch {
			withoutCard = true;
		}
		assert(withoutCard, 'a site without the picture of a shared link is refused');
		writeFileSync(join(fakeSiteRoot, 'public', 'og-image.png'), 'png');
		const written = publish(assembly, 'v0.1.0-alpha.3', fakeSiteRoot);
		assert(written.includes('CNAME'), 'the CNAME is published');
		assert(written.includes('.nojekyll'), 'the Jekyll marker is published');
		assert(readFileSync(join(assembly, 'og-image.png'), 'utf8') === 'png', 'the picture of a shared link is at the root');
		assert(statSync(join(assembly, 'robots.txt')).size > 0, 'robots.txt has content');
		assert(written.includes(STANDING_SITEMAP), 'the sitemap of the pages of no version is published');
		const plain = join(tmp, '_plain');
		mkdirSync(plain);
		assert(!publish(plain, 'edge', fakeSiteRoot, []).includes(STANDING_SITEMAP), 'no such pages, no such sitemap');

		writeFileSync(join(fakeSiteRoot, 'public', 'CNAME'), 'example.com\n');
		let refused = false;
		try {
			publish(assembly, 'edge', fakeSiteRoot);
		} catch {
			refused = true;
		}
		assert(refused, 'a CNAME naming another host is refused');

		rmSync(join(fakeSiteRoot, 'public', 'CNAME'));
		refused = false;
		try {
			publish(assembly, 'edge', fakeSiteRoot);
		} catch {
			refused = true;
		}
		assert(refused, 'a missing source is refused');
		console.log('publish-root-assets.mjs --selftest: OK');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}
}

function main() {
	const [command, ...rest] = process.argv.slice(2);
	if (command === '--selftest') {
		selftest();
		return;
	}
	const siteDir = command;
	const defaultSlug = rest[0];
	if (!siteDir || !defaultSlug) {
		console.error('usage: node scripts/publish-root-assets.mjs <site-dir> <default-version> [--at-root] | --selftest');
		process.exitCode = 2;
		return;
	}
	const written = publish(siteDir, defaultSlug, defaultSiteRoot, StandingPages, rest.includes('--at-root'));
	console.log(`published ${written.join(', ')} into ${siteDir}`);
}

if (process.argv[1] && process.argv[1].endsWith('publish-root-assets.mjs')) {
	main();
}
