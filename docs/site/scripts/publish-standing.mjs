#!/usr/bin/env node
// Puts the pages that belong to no version at the root of the assembly.
//
// The standing build renders the whole site with the root as its base. Only a
// few things of it are published: each page of no version, the assets those
// pages load, and the search index. The other pages of that build stay behind,
// because their addresses at the root belong to the stubs that lead into the
// default version.
import { cpSync, existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { extname, join } from 'node:path';

import { StandingPages } from '../src/lib/docs-origin.mjs';
import { isVersionFolder } from './gen-versions.mjs';

// BUNDLES holds what a page loads: the bundles and images of the build. Their
// names carry a hash of their content, so this folder is merged into the one
// that the latest release may already have put at the root.
const BUNDLES = '_astro';

// SEARCH is the search index. The root keeps the index of the latest release
// when it has one; the index of this build is published only into a root
// without one.
const SEARCH = 'pagefind';

// ROOT_FILE_TYPES are the files of public/ that a page names at the root, such
// as its icon and the picture of a shared link.
const ROOT_FILE_TYPES = new Set(['.svg', '.png', '.jpg', '.webp', '.ico']);

// publish copies the pages and what they load from a standing build into the
// assembly, and returns the names it wrote. A page that the build does not
// hold throws: past this step the artifact uploads, and the address of the
// page would answer with nothing.
export function publish(distDir, siteDir, pages = StandingPages) {
	const written = [];
	for (const page of pages) {
		if (isVersionFolder(page.split('/')[0])) throw new Error(`page ${page} has the name of a version`);
		if (!existsSync(join(distDir, page, 'index.html'))) {
			throw new Error(`the standing build holds no page ${page}`);
		}
		rmSync(join(siteDir, page), { recursive: true, force: true });
		cpSync(join(distDir, page), join(siteDir, page), { recursive: true });
		written.push(`${page}/`);
	}
	if (existsSync(join(distDir, BUNDLES))) {
		cpSync(join(distDir, BUNDLES), join(siteDir, BUNDLES), { recursive: true });
		written.push(`${BUNDLES}/`);
	}
	if (existsSync(join(distDir, SEARCH)) && !existsSync(join(siteDir, SEARCH))) {
		cpSync(join(distDir, SEARCH), join(siteDir, SEARCH), { recursive: true });
		written.push(`${SEARCH}/`);
	}
	for (const entry of readdirSync(distDir, { withFileTypes: true })) {
		if (!entry.isFile() || !ROOT_FILE_TYPES.has(extname(entry.name))) continue;
		// A file that the latest release already put at the root stays.
		if (existsSync(join(siteDir, entry.name))) continue;
		cpSync(join(distDir, entry.name), join(siteDir, entry.name));
		written.push(entry.name);
	}
	return written;
}

function selftest() {
	const assert = (condition, message) => {
		if (!condition) throw new Error(message);
	};
	const tmp = mkdtempSync(join(tmpdir(), 'unswell-docs-standing-'));
	try {
		const dist = join(tmp, 'dist');
		const site = join(tmp, '_site');
		for (const folder of ['report/part', 'guide', '_astro', 'pagefind']) mkdirSync(join(dist, folder), { recursive: true });
		mkdirSync(join(site, 'edge'), { recursive: true });
		writeFileSync(join(dist, 'report', 'index.html'), 'the report');
		writeFileSync(join(dist, 'report', 'part', 'index.html'), 'a page below it');
		writeFileSync(join(dist, 'guide', 'index.html'), 'a page of a version');
		writeFileSync(join(dist, 'index.html'), 'the front page of the build');
		writeFileSync(join(dist, '404.html'), 'not found');
		writeFileSync(join(dist, 'sitemap-index.xml'), 'sitemap');
		writeFileSync(join(dist, 'CNAME'), 'host');
		writeFileSync(join(dist, '_astro', 'page.css'), 'css');
		writeFileSync(join(dist, 'pagefind', 'index.js'), 'search');
		writeFileSync(join(dist, 'card.jpg'), 'jpg');
		writeFileSync(join(dist, 'favicon.svg'), 'svg');

		const written = publish(dist, site, ['report']);
		assert(readFileSync(join(site, 'report', 'index.html'), 'utf8') === 'the report', 'the page is at the root');
		assert(existsSync(join(site, 'report', 'part', 'index.html')), 'a page below it comes along');
		assert(existsSync(join(site, '_astro', 'page.css')), 'the bundles of the build are published');
		assert(existsSync(join(site, 'pagefind', 'index.js')), 'the search index is published');
		assert(existsSync(join(site, 'card.jpg')) && existsSync(join(site, 'favicon.svg')), 'the files a page names at the root are published');
		assert(!existsSync(join(site, 'guide')), 'a page of a version stays behind');
		assert(!existsSync(join(site, 'index.html')), 'the front page of the build stays behind');
		assert(!existsSync(join(site, '404.html')), 'the not-found page of the build stays behind');
		assert(!existsSync(join(site, 'sitemap-index.xml')), 'the sitemap of the build stays behind');
		assert(!existsSync(join(site, 'CNAME')), 'the CNAME is the business of another step');
		assert(written.includes('report/') && written.includes('_astro/'), 'the names are reported');

		// A root that the latest release already filled keeps its own search
		// index and its own files, and gains the bundles of this build.
		const filled = join(tmp, '_filled');
		mkdirSync(join(filled, '_astro'), { recursive: true });
		mkdirSync(join(filled, 'pagefind'), { recursive: true });
		writeFileSync(join(filled, '_astro', 'release.css'), 'css of the release');
		writeFileSync(join(filled, 'pagefind', 'index.js'), 'search of the release');
		writeFileSync(join(filled, 'favicon.svg'), 'icon of the release');
		publish(dist, filled, ['report']);
		assert(readFileSync(join(filled, 'pagefind', 'index.js'), 'utf8') === 'search of the release', 'the search index of the release stays');
		assert(readFileSync(join(filled, 'favicon.svg'), 'utf8') === 'icon of the release', 'a root file of the release stays');
		assert(existsSync(join(filled, '_astro', 'release.css')) && existsSync(join(filled, '_astro', 'page.css')), 'the bundles are merged');
		assert(existsSync(join(filled, 'card.jpg')), 'a file that only this build has is added');

		let refused = false;
		try {
			publish(dist, site, ['missing']);
		} catch {
			refused = true;
		}
		assert(refused, 'a page that the build does not hold is refused');
		refused = false;
		try {
			publish(dist, site, ['edge']);
		} catch {
			refused = true;
		}
		assert(refused, 'a page with the name of a version is refused');
		console.log('publish-standing.mjs --selftest: OK');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}
}

function main() {
	const [first, second] = process.argv.slice(2);
	if (first === '--selftest') {
		selftest();
		return;
	}
	if (!first || !second) {
		console.error('usage: node scripts/publish-standing.mjs <standing-dist> <site-dir> | --selftest');
		process.exitCode = 2;
		return;
	}
	const written = publish(first, second);
	console.log(`published ${written.join(', ')} into ${second}`);
}

if (process.argv[1] && process.argv[1].endsWith('publish-standing.mjs')) {
	main();
}
