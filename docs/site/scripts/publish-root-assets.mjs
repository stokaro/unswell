#!/usr/bin/env node
// Writes the files that belong at the Pages ROOT rather than inside a versioned
// documentation directory.
//
// The root holds one directory per version plus a few files that address the
// site as a whole. gen-versions.mjs writes two of them, versions.json and
// index.html. This script writes the rest: the CNAME that claims the domain,
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

import { Host, Origin } from '../src/lib/docs-origin.mjs';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const defaultSiteRoot = join(scriptDir, '..');

// COPIED_ASSETS are published byte for byte from docs/site/public/. They live
// there rather than beside this script so that a change to them runs the docs
// workflow, whose path filter watches docs/site/**. Astro also copies public/
// into every version's dist/, which is harmless: the addresses that matter are
// the root ones this script writes.
export const COPIED_ASSETS = [
	{ name: 'CNAME', source: 'public/CNAME' },
	{ name: '.nojekyll', source: 'public/.nojekyll' },
];

// GENERATED_ROOT_FILES names what gen-versions.mjs writes into the same
// directory. This script does not write them; naming them here lets the Pages
// root gate compare the whole expected root against a real assembly.
export const GENERATED_ROOT_FILES = ['versions.json', 'index.html'];

// renderRobots points crawlers at the sitemap of the version the apex serves.
// Every other version is reachable and indexable, but only one of them should
// be the copy a search engine prefers, and that is the one the canonical links
// of the default version already name.
export function renderRobots(defaultSlug) {
	return `User-agent: *
Allow: /

Sitemap: ${Origin}/${defaultSlug}/sitemap-index.xml
`;
}

// publish copies and writes every root file into siteDir and returns the names
// it wrote. A missing or empty source throws rather than being skipped: past
// this step the artifact uploads, and a missing CNAME takes the domain down
// with nothing red anywhere.
export function publish(siteDir, defaultSlug, siteRoot = defaultSiteRoot) {
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
	writeFileSync(join(siteDir, 'robots.txt'), renderRobots(defaultSlug));
	written.push('robots.txt');
	return written;
}

function selftest() {
	const assert = (condition, message) => {
		if (!condition) throw new Error(message);
	};

	assert(renderRobots('edge').includes(`${Origin}/edge/sitemap-index.xml`), 'robots names the sitemap');
	assert(renderRobots('edge').includes('User-agent: *'), 'robots addresses every crawler');

	const tmp = mkdtempSync(join(tmpdir(), 'unswell-docs-root-'));
	try {
		const fakeSiteRoot = join(tmp, 'site');
		mkdirSync(join(fakeSiteRoot, 'public'), { recursive: true });
		writeFileSync(join(fakeSiteRoot, 'public', 'CNAME'), `${Host}\n`);
		writeFileSync(join(fakeSiteRoot, 'public', '.nojekyll'), '');
		const assembly = join(tmp, '_site');
		mkdirSync(assembly);
		const written = publish(assembly, 'v0.1.0-alpha.3', fakeSiteRoot);
		assert(written.includes('CNAME'), 'the CNAME is published');
		assert(written.includes('.nojekyll'), 'the Jekyll marker is published');
		assert(statSync(join(assembly, 'robots.txt')).size > 0, 'robots.txt has content');

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
		console.error('usage: node scripts/publish-root-assets.mjs <site-dir> <default-version> | --selftest');
		process.exitCode = 2;
		return;
	}
	const written = publish(siteDir, defaultSlug);
	console.log(`published ${written.join(', ')} into ${siteDir}`);
}

if (process.argv[1] && process.argv[1].endsWith('publish-root-assets.mjs')) {
	main();
}
