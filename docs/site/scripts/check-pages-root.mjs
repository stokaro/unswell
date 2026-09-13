#!/usr/bin/env node
// Two assertions about what a build is allowed to publish.
//
// The CNAME must name the domain this site claims. A CNAME naming anything else
// hands the domain back to GitHub's default host on the next deploy, and the
// certificate that was approved for it stops matching.
//
// No emitted text may contain a github.io address. The Pages default host
// answers every one of them, so such a link works in a browser and is still
// wrong: it publishes an address the project does not control the redirect of,
// and search engines index the copy rather than the domain.
import { existsSync, readdirSync, readFileSync, mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, extname, join, posix, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { Host } from '../src/lib/docs-origin.mjs';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const siteRoot = join(scriptDir, '..');

// The formats a reader or a crawler follows an address out of: the pages, the
// sitemap, robots.txt, and the extensionless root files.
//
// Bundled JavaScript is deliberately not scanned. Starlight's interface strings
// and Pagefind's own bundle carry a translator's personal github.io address in
// their locale data. It is upstream content, it is not an address of this site,
// and refusing it would only mean vendoring a search bundle to satisfy a gate.
// What this check is for is an address the project publishes as its own.
const TEXT_EXTENSIONS = new Set(['.html', '.xml', '.txt', '']);

export const FORBIDDEN_HOST = 'github.io';

export function readCname(path) {
	return readFileSync(path, 'utf8').trim();
}

export function collectTextFiles(root, prefix = '') {
	const found = [];
	for (const entry of readdirSync(join(root, prefix), { withFileTypes: true })) {
		const relative = posix.join(prefix, entry.name);
		if (entry.isDirectory()) {
			found.push(...collectTextFiles(root, relative));
		} else if (TEXT_EXTENSIONS.has(extname(entry.name))) {
			found.push(relative);
		}
	}
	return found;
}

export function findForbiddenHost(root) {
	const hits = [];
	for (const file of collectTextFiles(root)) {
		if (readFileSync(join(root, file), 'utf8').includes(FORBIDDEN_HOST)) {
			hits.push(file);
		}
	}
	return hits;
}

function selftest() {
	const assert = (condition, message) => {
		if (!condition) throw new Error(message);
	};

	const tmp = mkdtempSync(join(tmpdir(), 'unswell-docs-pages-root-'));
	try {
		const dist = join(tmp, 'dist');
		mkdirSync(join(dist, 'overview'), { recursive: true });
		writeFileSync(join(dist, 'CNAME'), `${Host}\n`);
		assert(readCname(join(dist, 'CNAME')) === Host, 'the CNAME reads back trimmed');

		writeFileSync(join(dist, 'overview', 'index.html'), '<a href="https://unswell.dev/">home</a>');
		assert(findForbiddenHost(dist).length === 0, 'a clean build has no default-host address');

		writeFileSync(join(dist, 'overview', 'index.html'), '<a href="https://stokaro.github.io/unswell/">x</a>');
		const hits = findForbiddenHost(dist);
		assert(hits.length === 1, 'one offending file is one hit');
		assert(hits[0] === 'overview/index.html', 'the hit names the file');

		writeFileSync(join(dist, 'overview', 'index.html'), 'clean');
		writeFileSync(join(dist, 'sitemap-0.xml'), `<loc>https://stokaro.github.io/x/</loc>`);
		assert(findForbiddenHost(dist).length === 1, 'a sitemap is scanned too');
		console.log('check-pages-root.mjs --selftest: OK');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}
}

function main() {
	if (process.argv.includes('--selftest')) {
		selftest();
		return;
	}
	const distIndex = process.argv.indexOf('--dist');
	const distDir = distIndex === -1 ? join(siteRoot, 'dist') : resolve(process.argv[distIndex + 1]);
	if (!existsSync(distDir)) {
		console.error(`error: no built site at ${distDir}; run npm run build first`);
		process.exitCode = 2;
		return;
	}
	const problems = [];
	const cnamePath = join(distDir, 'CNAME');
	if (!existsSync(cnamePath)) {
		problems.push(`no CNAME in ${distDir}`);
	} else {
		const cname = readCname(cnamePath);
		if (cname !== Host) problems.push(`CNAME names ${cname}, not ${Host}`);
	}
	for (const file of findForbiddenHost(distDir)) {
		problems.push(`${file} contains a ${FORBIDDEN_HOST} address`);
	}
	for (const problem of problems) console.error(`pages root: ${problem}`);
	if (problems.length > 0) {
		process.exitCode = 1;
		return;
	}
	console.log(`check-pages-root.mjs: ${distDir} claims ${Host} and names no ${FORBIDDEN_HOST} address`);
}

if (process.argv[1] && process.argv[1].endsWith('check-pages-root.mjs')) {
	main();
}
