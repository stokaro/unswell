#!/usr/bin/env node
// Refuses an internal link that resolves to nothing.
//
// The check reads the built output rather than the Markdown sources. A link is
// written in four places -- page prose, the sidebar, a component, a redirect --
// and only the emitted HTML shows what a reader will actually click. It also
// means a broken link is caught under the same `base` the deploy publishes,
// which is where a versioned site usually breaks.
import { existsSync, readdirSync, readFileSync, mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, posix, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const defaultDist = join(scriptDir, '..', 'dist');

// HREF_RE collects href and src values. Both can address a page, and a stale
// asset path fails a reader exactly as loudly as a stale page path.
const HREF_RE = /(?:href|src)="([^"]*)"/g;

// EXTERNAL_RE matches anything with a scheme or a protocol-relative prefix, plus
// the addressing schemes that never resolve to a file.
const EXTERNAL_RE = /^(?:[a-z][a-z0-9+.-]*:|\/\/)/i;

export function collectFiles(root, prefix = '') {
	const found = [];
	for (const entry of readdirSync(join(root, prefix), { withFileTypes: true })) {
		const relative = posix.join(prefix, entry.name);
		if (entry.isDirectory()) {
			found.push(...collectFiles(root, relative));
		} else {
			found.push(relative);
		}
	}
	return found;
}

// extractLinks returns the distinct internal targets of one HTML document,
// already stripped of the fragment and the query.
export function extractLinks(html) {
	const targets = new Set();
	for (const match of html.matchAll(HREF_RE)) {
		const raw = match[1].trim();
		if (raw === '' || raw.startsWith('#') || EXTERNAL_RE.test(raw)) continue;
		const withoutFragment = raw.split('#')[0].split('?')[0];
		if (withoutFragment === '') continue;
		targets.add(withoutFragment);
	}
	return [...targets];
}

// resolveTarget maps one link to the file the server would answer with.
//
// The two coordinate systems have to be joined here: an emitted file sits at a
// path within dist/, while a link is written in the addresses the server uses,
// which carry the version base in front. `fromRoute` is the file path, so the
// base goes in front of it before a relative link is resolved against it.
//
// A directory address is answered by its index.html; an extensionless address
// is answered either by a file of that exact name or by such an index.
export function resolveTarget(target, fromRoute, base) {
	const fromAddress = posix.join(base, fromRoute);
	const absolute = target.startsWith('/')
		? target
		: posix.resolve(posix.dirname(fromAddress), target);
	if (!absolute.startsWith(base)) return null;
	const withinSite = absolute.slice(base.length);
	const candidates = [];
	if (withinSite === '' || withinSite.endsWith('/')) {
		candidates.push(posix.join(withinSite, 'index.html'));
	} else {
		candidates.push(withinSite, posix.join(withinSite, 'index.html'), `${withinSite}.html`);
	}
	return candidates.map((candidate) => candidate.replace(/^\/+/, ''));
}

export function check(distDir, base) {
	const files = collectFiles(distDir);
	const present = new Set(files);
	const problems = [];
	for (const file of files) {
		if (!file.endsWith('.html')) continue;
		const html = readFileSync(join(distDir, file), 'utf8');
		for (const target of extractLinks(html)) {
			const candidates = resolveTarget(target, file, base);
			if (candidates === null) {
				problems.push(`${file}: ${target} leaves this version's base ${base}`);
				continue;
			}
			if (!candidates.some((candidate) => present.has(candidate))) {
				problems.push(`${file}: ${target} resolves to nothing`);
			}
		}
	}
	return problems;
}

function selftest() {
	const assert = (condition, message) => {
		if (!condition) throw new Error(message);
	};

	assert(extractLinks('<a href="https://example.com/">x</a>').length === 0, 'an external link is skipped');
	assert(extractLinks('<a href="#top">x</a>').length === 0, 'a bare fragment is skipped');
	assert(extractLinks('<a href="mailto:a@b.c">x</a>').length === 0, 'a mail address is skipped');
	assert(extractLinks('<a href="/edge/a/#b">x</a>')[0] === '/edge/a/', 'the fragment is stripped');

	const tmp = mkdtempSync(join(tmpdir(), 'unswell-docs-links-'));
	try {
		const dist = join(tmp, 'dist');
		mkdirSync(join(dist, 'overview'), { recursive: true });
		mkdirSync(join(dist, 'rules'), { recursive: true });
		writeFileSync(
			join(dist, 'overview', 'index.html'),
			'<a href="/edge/rules/">rules</a><a href="https://unswell.dev/">home</a>'
		);
		writeFileSync(join(dist, 'rules', 'index.html'), '<a href="../overview/">back</a>');
		assert(check(dist, '/edge/').length === 0, 'absolute and relative links both resolve');

		writeFileSync(join(dist, 'rules', 'index.html'), '<a href="/edge/missing/">gone</a>');
		const problems = check(dist, '/edge/');
		assert(problems.length === 1, 'one broken link is one problem');
		assert(problems[0].includes('resolves to nothing'), 'the problem names the failure');

		writeFileSync(join(dist, 'rules', 'index.html'), '<a href="/other/page/">wrong base</a>');
		assert(check(dist, '/edge/')[0].includes('leaves this version'), 'a link outside the base is refused');
		console.log('check-links.mjs --selftest: OK');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}
}

function main() {
	if (process.argv.includes('--selftest')) {
		selftest();
		return;
	}
	const version = process.env.DOCS_VERSION || 'edge';
	const base = `/${version}/`;
	const distIndex = process.argv.indexOf('--dist');
	const distDir = distIndex === -1 ? defaultDist : resolve(process.argv[distIndex + 1]);
	if (!existsSync(distDir)) {
		console.error(`error: no built site at ${distDir}; run npm run build first`);
		process.exitCode = 2;
		return;
	}
	const problems = check(distDir, base);
	for (const problem of problems) console.error(`broken link: ${problem}`);
	if (problems.length > 0) {
		console.error(`${problems.length} internal links resolve to nothing`);
		process.exitCode = 1;
		return;
	}
	console.log(`check-links.mjs: every internal link under ${base} resolves`);
}

if (process.argv[1] && process.argv[1].endsWith('check-links.mjs')) {
	main();
}
